package gasservice

import (
	"net/http"
)

type MockClient struct {
	HTTP *http.Client
}

func (mgs *MockClient) GetAccountAndSequence() (string, uint64, error) {
	return "GBYYNSO5QYTZD6YFY63CGHGTFPUPQZHJOSKNIRBDZM6MPC3QP7OPIQ5E", 3736260770267154, nil
}

func (mgs *MockClient) SubmitTxe(txeOfiRfiSignedB64 string) (string, uint64, error) {
	return "", 0, nil
}
