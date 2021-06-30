package utility

type CredentialInfo struct {
	Environment string
	Domain      string
	Service     string
	Variable    string
}

type ParameterContent struct {
	Value       string
	Description string
}

type SecretContent struct {
	Entry       []SecretEntry
	Description string
	FilePath    string
	RawJson     []byte
}

type SecretEntry struct {
	Key   string
	Value string
}
