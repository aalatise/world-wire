package authservice

import (
	"net/http"
)

type MockClient struct {
	HTTP *http.Client
}

func (asc *MockClient) VerifyTokenAndEndpoint(jwt string, endpoint string) (bool, error) {
	return true, nil
}
