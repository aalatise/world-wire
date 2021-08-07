package vault

import (
	"encoding/json"
	"errors"
	"github.com/IBM/world-wire/utility/nodeconfig/secrets"
	"os"

	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"github.com/IBM/world-wire/utility/nodeconfig"
)

func (vault *Vault) StoreAccount(accountName string, account nodeconfig.Account, randString string) error {

	// creating account
	LOGGER.Infof("Storing account %s", accountName)
	envVersion := os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION)
	participantId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	secretPath := secrets.ConstructPath(secrets.CredentialInfo{
		Environment: envVersion,
		Domain:      participantId,
		Service:     "account",
		Variable:    accountName,
	})

	var input map[string]interface{}

	b, _ := json.Marshal(account)
	json.Unmarshal(b, &input)

	err := vault.Create(input, secretPath)
	if err != nil {
		LOGGER.Errorf("Ancounter error while creating account: %v", err)
		return err
	}

	LOGGER.Infof("Account created!")

	// store killswitch secret string
	LOGGER.Infof("Now storing killswitch string for the account")

	killSwitchPath := secrets.ConstructPath(secrets.CredentialInfo{
		Environment: envVersion,
		Domain:      participantId,
		Service:     "killswitch",
		Variable:    "accounts",
	})

	killswtichInput := make(map[string]interface{})
	var updateErr error
	killswtichInput[account.NodeAddress] = randString
	_, err = vault.Read(killSwitchPath)
	if err != nil {
		updateErr = vault.Create(killswtichInput, killSwitchPath)
	} else {
		updateErr = vault.Append(killswtichInput, killSwitchPath)
	}

	if updateErr != nil {
		return errors.New("Encounter error while creating killswitch secrets")
	}

	return nil
}

func (vault *Vault) GetAccount(participantId, accountName string) (nodeconfig.Account, error) {

	LOGGER.Infof("Reading account %s for %s", accountName, participantId)
	envVersion := os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION)
	secretPath := secrets.ConstructPath(secrets.CredentialInfo{
		Environment: envVersion,
		Domain:      participantId,
		Service:     "account",
		Variable:    accountName,
	})
	result, err := vault.Read(secretPath)
	if err != nil {
		LOGGER.Errorf("Cannot get the account %s for participant %s: %v", accountName, participantId, err)
		return nodeconfig.Account{}, err
	}

	var account nodeconfig.Account
	b, _ := json.Marshal(result)
	json.Unmarshal(b, &account)
	return account, nil

}
