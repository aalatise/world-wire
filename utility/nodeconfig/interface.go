package nodeconfig

import "github.com/IBM/world-wire/utility/nodeconfig/secrets"

type Client interface {
	Read(string) (map[string]interface{}, error)
	Create(map[string]interface{}, string) error
	Append(map[string]interface{}, string) error
	Update(map[string]interface{}, string) error
	Override(map[string]interface{}, string) error
	Delete(string) error

	// admin
	// domainId, accountName
	GetAccount(string, string) (Account, error)
	GetSecretPhrase(string, string) (string, error)
	DeleteSingleSecretEntry(secrets.CredentialInfo, string) error
	AppendSecret(secrets.CredentialInfo, map[string]interface{}) error

	//api
	StoreAccount(string, Account, string) error
	InitEnv()
}
