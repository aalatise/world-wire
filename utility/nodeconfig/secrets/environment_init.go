package secrets

import (
	"github.com/IBM/world-wire/utility/nodeconfig/secrets/vault"
	"os"
	"strings"

	"github.com/IBM/world-wire/utility/common"

	global_environment "github.com/IBM/world-wire/utility/global-environment"
)

func InitEnv() {
	if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == common.HASHICORP_VAULT_SECRET {
		client, err := vault.InitializeVault()
		if err != nil {
			panic(err)
		}
		client.InitEnv()
	} else {
		panic("No valid secret storage location is specified")
	}
}
