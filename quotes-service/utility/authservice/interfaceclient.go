package authservice

type InterfaceClient interface {
	VerifyTokenAndEndpoint(jwt string, endpoint string) (bool, error)
}
