#!/usr/bin/python3
import json
import sys
import http
import re
import hvac
import getpass

# Change this value to work with a different vault instance
vaultURL="https://vault.iksdapww-test-us-south-1-2a0beb393d3242574412e5315d3d4662-0002.us-south.containers.appdomain.cloud"


def loadVault(secret_path,remaining_object,variable_values):
  secret_dict={}
  if (type(remaining_object) is dict):
    for current_key in remaining_object:
      current_value=remaining_object[current_key]
      if (type(current_value) is dict):
        loadVault(secret_path + [current_key],current_value, variable_values)
      elif (type(current_value) is str):
        m=re.match(r'{{\s*(.*\S)\s*}}',current_value)
        if (m is None):
          secret_dict[current_key]=current_value
        else:  
          variable_name=m.group(1)
          if (variable_name in variable_values):
            secret_value=variable_values[variable_name]
            secret_dict[current_key]=secret_value
          else:
            print("WARNING unrecognized template variiable - " + variable_name)
            secret_dict[current_key]=current_value
    if (len(secret_dict) > 0):
      storeSecret(secret_path,secret_dict)        
  else:  
    storeSecret(secret_path,remaining_object)

def storeSecret(secret_path,secret_dict):
  vault_path='sandbox/ww/' + '/'.join(secret_path) + '/initialize'
  print(vault_path)
  for (current_key,current_value) in secret_dict.items():
    print('    ',current_key) #,'=',current_value)
  client.secrets.kv.v2.create_or_update_secret(path=vault_path,mount_point='ww',secret=secret_dict)
  print("stored")

if (len(sys.argv) != 3):
  print("Usage: loadvault.py <JSON secrets file path> <JSON variable substitution dictionary file path>")
  print("  You will be prompted for the token to access the vault at:")
  print(vaultURL)
  print("  To change the vault URL, edit 'vaultURL' value in the program (near the top)")
  exit(1)
client=hvac.Client(url=vaultURL,verify=False)
client.token=getpass.getpass(prompt="Enter the token for " + vaultURL)
variable_file=open(sys.argv[-1],'r')
variable_values=json.load(variable_file)
variable_file.close()
json_file=open(sys.argv[-2],'r')
remaining_object=json.load(json_file)
json_file.close()
loadVault([],remaining_object,variable_values)           
