package client

import "github.com/IBM/world-wire/gftn-models/model"

type AdministrationServiceClient interface {
	GetTxnDetails(txnDetailsRequest model.FItoFITransactionRequest) ([]byte, int, error)
}
