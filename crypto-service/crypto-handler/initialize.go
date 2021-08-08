package crypto_handler

import (
	"errors"
	"os"
	"strings"

	"github.com/op/go-logging"
	"github.com/IBM/world-wire/crypto-service/environment"
	"github.com/IBM/world-wire/crypto-service/utility/common"
	"github.com/IBM/world-wire/crypto-service/utility/constant"
	comn "github.com/IBM/world-wire/utility/common"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"github.com/IBM/world-wire/utility/nodeconfig/secrets"
	"github.com/IBM/world-wire/utility/nodeconfig/secrets/vault"
)

var LOGGER = logging.MustGetLogger("crypto-handler")

type CryptoOperations struct {
	HSMInstance common.HsmObject
	secrets     secrets.Client
}

//Global handler variable used for clean up session in the end
var CYPTO_OPERATIONS = CryptoOperations{}

func CreateCryptoOperations() (op CryptoOperations, err error) {

	if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == comn.HASHICORP_VAULT_SECRET {
		op.secrets, err = vault.InitializeVault()
		if err != nil {
			panic(err)
		}
		op.secrets.InitEnv()
	} else {
		panic("No valid secret storage location is specified")
	}

	if strings.ToUpper(os.Getenv(environment.ENV_KEY_ACCOUNT_SOURCE)) == constant.ACCOUNT_FROM_HSM {
		LOGGER.Infof("Using HSM as account source")
		if os.Getenv(environment.ENV_KEY_PKCS11_PIN) == "" || os.Getenv(environment.ENV_KEY_PKCS11_SLOT) == "" {
			LOGGER.Errorf("Error reading PKCS11_PIN && PKCS11_SLOT environment settings")
			return op, errors.New("Error reading PKCS11_PIN && PKCS11_SLOT environment settings")
		}
		if op.HSMInstance.C == nil {
			LOGGER.Infof("Initializing HSM client")
			op.HSMInstance.C, op.HSMInstance.Slot, op.HSMInstance.Session = common.InitiateHSM()
		} else {
			LOGGER.Infof("HSM client already initialized. Skipped")
		}
	} else if strings.ToUpper(os.Getenv(environment.ENV_KEY_ACCOUNT_SOURCE)) == constant.ACCOUNT_FROM_STELLAR {
		LOGGER.Infof("Using HSM as account source")
	} else {
		LOGGER.Errorf("Error reading account source environment settings")
	}
	CYPTO_OPERATIONS = op
	return op, nil
}
