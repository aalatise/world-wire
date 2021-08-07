package vault

import (
	"errors"
	"github.com/IBM/world-wire/utility/secrets"
	"os"
	"strings"

	global_environment "github.com/IBM/world-wire/utility/global-environment"
)

func (vault *Vault) InitEnv() {
	//getGlobalEnv("initialize")
	vault.getParticipantEnv("initialize")
	vault.getServiceEnv("initialize")
}

// All services

func (vault *Vault) getServiceEnv(var_name string) {
	domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	svcName := os.Getenv(global_environment.ENV_KEY_SERVICE_NAME)
	envVersion := os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION)
	LOGGER.Infof("Initializing service env variables with Domain name: %s, Service name: %s, Environment version: %s", domainId, svcName, envVersion)
	err := vault.getenv(secrets.CredentialInfo{
		Environment: envVersion,
		Domain:      domainId,
		Service:     svcName,
		Variable:    var_name,
	})
	if err != nil {
		LOGGER.Errorf("%v", err)
		panic("Error initializing service with Hashicorp Vault")
	}
}

func (vault *Vault) getParticipantEnv(var_name string) {

	domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	envVersion := os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION)
	LOGGER.Infof("Initializing participant: %s specific env variables", domainId)
	err := vault.getenv(secrets.CredentialInfo{
		Environment: envVersion,
		Domain:      domainId,
		Service:     "participant",
		Variable:    var_name,
	})
	if err != nil {
		LOGGER.Errorf("%v", err)
		panic("Error initializing service with Hashicorp Vault")
	}
}

func (vault *Vault) getenv(credInfo secrets.CredentialInfo) error {

	secretPath := secrets.ConstructPath(credInfo)
	LOGGER.Infof("Getting environment variable: %s", secretPath)
	secretPath = strings.ToLower(secretPath)

	if _, exists := os.LookupEnv(secretPath); exists {
		LOGGER.Infof("Environment variable %s already defined", secretPath)
		return nil
	} else {

		LOGGER.Infof("Retrieving environment variable %s from Hashicorp Vault", secretPath)

		result, err := vault.Read(secretPath)
		if err != nil {
			LOGGER.Errorf("Cannot get the specified environment variable: %s", err)
			return err
		}

		for key, val := range result {
			os.Setenv(key, val.(string))
		}
		os.Setenv(secretPath, "true")
	}
	return nil

}

// admin services

func (vault *Vault) GetSecretPhrase(participantId, accountName string) (string, error) {

	LOGGER.Infof("Retrieving killswitch info for account %s for %s", accountName, participantId)

	credential := secrets.CredentialInfo{
		Environment: os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION),
		Domain:      participantId,
		Service:     "killswitch",
		Variable:    "accounts",
	}

	secretPath := secrets.ConstructPath(credential)

	res, err := vault.Read(secretPath)
	if err != nil {
		LOGGER.Errorf("Cannot get the killswitch info for participant %s: %v", participantId, err)
		return "", err
	}

	if _, found := res[accountName]; !found {
		LOGGER.Errorf("Cannot get the killswitch info for account %s", accountName)
		return "", errors.New("Account not found")
	}

	return res[accountName].(string), nil

}

func (vault *Vault) DeleteSingleSecretEntry(credentialInfo secrets.CredentialInfo, target string) error {

	LOGGER.Infof("DeleteSingleSecretEntry: %s, entry name: %s", credentialInfo, target)
	secretPath := secrets.ConstructPath(credentialInfo)

	res, err := vault.Read(secretPath)
	if err != nil {
		LOGGER.Errorf("Cannot get the specified environment variable: %s", err)
		return err
	}

	if _, exists := res[target]; !exists {
		return errors.New("target key not found")
	}

	delete(res, target)

	err = vault.Override(res, secretPath)
	if err != nil {
		errMsg := "Error while updating secret in the deleteSingeSecretEntry process"
		LOGGER.Errorf(errMsg)
		return errors.New(errMsg)
	}
	LOGGER.Infof("DeleteSingleSecretEntry: Success!")
	return nil
}

func (vault *Vault) AppendSecret(credInfo secrets.CredentialInfo, input map[string]interface{}) error {

	secretPath := secrets.ConstructPath(credInfo)
	LOGGER.Infof("Appending secret to path %v", secretPath)
	err := vault.Append(input, secretPath)
	if err != nil {
		LOGGER.Errorf("Failed appending secrets to path %s: %s", secretPath, err)
		return err
	}

	return nil

}

// api service

func (vault *Vault) UpdateAccount(credInfo secrets.CredentialInfo, input map[string]interface{}) error {

	secretPath := secrets.ConstructPath(credInfo)
	secretPath = strings.ToLower(secretPath)
	LOGGER.Infof("Updating environment variable: %s", secretPath)

	err := vault.Update(input, secretPath)
	if err != nil {
		return err
	}
	for key, value := range input {
		LOGGER.Debugf("account: %v: %v \n", key, value)
		os.Setenv(credInfo.Variable+"_"+key, value.(string))
	}
	os.Setenv(credInfo.Variable, "true")
	return nil

}

func (vault *Vault) CreateAccount(credInfo secrets.CredentialInfo, input map[string]interface{}) error {

	secretPath := secrets.ConstructPath(credInfo)
	secretPath = strings.ToLower(secretPath)
	LOGGER.Infof("Creating environment variable: %s", secretPath)

	err := vault.Create(input, secretPath)
	if err != nil {
		return err
	}
	for key, value := range input {
		LOGGER.Debugf("account: %v: %v \n", key, value)
		os.Setenv(credInfo.Variable+"_"+key, value.(string))
	}
	os.Setenv(credInfo.Variable, "true")
	return nil

}
