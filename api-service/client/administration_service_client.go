package client

import "github.ibm.com/gftn/world-wire-services/gftn-models/model"

type AdministrationServiceClient interface {
	GetTxnDetails(txnDetailsRequest model.FItoFITransactionRequest) ([]byte, int, error)
}
