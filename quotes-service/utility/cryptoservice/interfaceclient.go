package cryptoservice

import "github.com/IBM/world-wire/gftn-models/model"

type InterfaceClient interface {
	RequestSigning(txeBase64 string, requestBase64 string, signedRequestBase64 string, accountName string, participant model.Participant) (string, error)
}
