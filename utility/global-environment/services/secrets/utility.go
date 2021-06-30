package secrets

type CredentialInfo struct {
	Environment string
	Domain      string
	Service     string
	Variable    string
}

func ConstructPath(credInfo CredentialInfo) string {
	return credInfo.Environment + "/" + credInfo.Domain + "/" + credInfo.Service + "/" + credInfo.Variable
}
