package secrets

import "github.com/IBM/world-wire/utility/nodeconfig"

type Client interface {
	Read(string) (map[string]interface{}, error)
	Create(map[string]interface{}, string) error
	Append(map[string]interface{}, string) error
	Update(map[string]interface{}, string) error
	Override(map[string]interface{}, string) error
	Delete(string) error

	// admin
	// domainId, accountName
	GetAccount(string, string) (nodeconfig.Account, error)
	GetSecretPhrase(string, string) (string, error)
	DeleteSingleSecretEntry(CredentialInfo, string) error
	AppendSecret(CredentialInfo, map[string]interface{}) error

	//api
	StoreAccount(string, nodeconfig.Account, string) error
	InitEnv()
}
