package cryptoservice

import "github.ibm.com/gftn/world-wire-services/gftn-models/model"

type InterfaceClient interface {
	RequestSigning(txeBase64 string, requestBase64 string, signedRequestBase64 string, accountName string, participant model.Participant) (string, error)
}
