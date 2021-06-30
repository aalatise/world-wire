import os
import sys
import argparse
import certifi
import hvac
import pprint
import json

import warnings
from urllib3.exceptions import InsecureRequestWarning

warnings.filterwarnings("ignore", category=InsecureRequestWarning)

dont_print = ['DB_PWD', 'DB_USER', 'MONGO_CONNECTION_CERT', 'MONGO_ID',
              'SENDGRID_API_KEY',
              'POSTGRESQLPASSWORD', 'POSTGRESQLUSER', 'POSTGRESQLHOST',
              'POSTGRESQL_DB_URL', 'POSTGRESQL_PWD', 'POSTGRESQL_USER',
              'KAFKA_KEY_SASL_PASSWORD', 'KAFKA_BROKER_URL',
              'NODE_ADDRESS', 'NODE_SEED',
              'PKCS11_PIN', 'PKCS11_SLOT',
              'ANCHOR_SH_CRED', 'ANCHOR_SH_PASS', 'ANCHOR_SH_SEC']

secrets_dict = {}

do_print_sandbox = ['template', 'ww']



class VaultListing:

    def __init__(self, v_client: hvac.Client):
        self.client = v_client
        self.listing = dict()
        self.secrets = dict()

    def list_secrets(self, secret_path, mount):
        l_ = self.client.secrets.kv.v2.list_secrets(path=secret_path, mount_point=mount)
        keys_ = l_['data']['keys']
        for k_ in keys_:
            if not k_.endswith('/'):
                self.read_secret(secret_path + k_, mount)
        for k_ in keys_:
            if k_.endswith('/'):
                self.list_secrets(secret_path + k_, mount)


    def blackout_secrets(self, secret):
        for k_ in secret.keys():
            if k_ in dont_print:
                self.secrets[k_] = secret[k_]
                secret[k_] = '...'


    def read_secret(self, secret_path, mount):
        secret_version_response = self.client.secrets.kv.v2.read_secret_version(path=secret_path,
                                                                                mount_point=mount)
        secret_name = mount + '/' + secret_path
        '''
        secret = {'version': secret_version_response['data']['metadata']['version'],
                  'created': secret_version_response['data']['metadata']['created_time'],
                  'data': secret_version_response['data']['data']}
        self.listing[secret_name] = secret
        '''

        # ignore participants
        splits = secret_name.split('/')
        if splits[1] == "sandbox" and splits[2] not in do_print_sandbox:
            return

        # blackout secret values
        s_ = secret_version_response['data']['data']
        if splits[-1] == "participant-template":
            for k_ in s_:
                self.blackout_secrets(s_[k_])
        else:
            self.blackout_secrets(s_)

        self.listing[secret_name] = s_

    def get_listing(self):
        return self.listing

    def get_secrets(self):
        return self.secrets


if __name__ == '__main__':
    # Process arguments
    url = ''
    token = ''
    mount_point = ''
    path = ''
    json_print = True

    parser = argparse.ArgumentParser(description="Traverse a Vault tree")
    parser.add_argument("-u", "--url", help="Vault URL", required=True)
    parser.add_argument("-t", "--token", help="Vault token", required=True)
    parser.add_argument("-m", "--mount-point", help="Mount point of the Vault", default='ww')
    parser.add_argument("-r", "--secrets-root", help="Secrets root node", default='')
    parser.add_argument("-o", "--output-dir", help="JSON output directory", default="./")
    parser.add_argument("-p", "--print", help="Pretty print")

    args = parser.parse_args()

    # Command line arguments
    vault_url: str = args.url
    vault_token: str = args.token
    ww_path: str = args.secrets_root
    ww_mount: str = args.mount_point
    output: str = args.output_dir

    # Create Hashicorp Client
    hvac_client = hvac.Client(url=vault_url, verify=False)
    hvac_client.token = vault_token

    # Vault Listing
    vaultListing = VaultListing(v_client=hvac_client)
    vaultListing.list_secrets(ww_path, ww_mount)

    # Print Vault listing
    if args.print:
        pp = pprint.PrettyPrinter(indent=2)
        pp.pprint(vaultListing.get_listing())
        pp.pprint(vaultListing.get_secrets())

    listingFile = os.path.join(output, "vault-listing.json")
    secretsFile = os.path.join(output, "vault-secrets.json")
    with open(listingFile, "w") as fp:
        json.dump(vaultListing.get_listing(), fp)
        print("JSON listing dumped to {}".format(listingFile))
    with open(secretsFile, "w") as fp:
        json.dump(vaultListing.get_secrets(), fp)
        print("JSON listing dumped to {}".format(secretsFile))
