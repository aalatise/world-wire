package services

import (
	"os"
	"strings"

	"github.ibm.com/gftn/world-wire-services/utility/common"

	secret_manager "github.ibm.com/gftn/world-wire-services/utility/aws/golang/secret-manager"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/global-environment/services/secrets/vault"
)

func InitEnv() {
	if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == common.AWS_SECRET {
		secret_manager.InitEnv()
	} else if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == common.HASHICORP_VAULT_SECRET {
		client, err := vault.InitializeVault()
		if err != nil {
			panic(err)
		}
		client.InitEnv()
	} else {
		panic("No valid secret storage location is specified")
	}
}
