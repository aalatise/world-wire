package crypto_client

import (
	"github.com/IBM/world-wire/gftn-models/model"
	"github.com/IBM/world-wire/utility/nodeconfig"
)

type CryptoServiceClient interface {
	CreateAccount(accountName string) (account nodeconfig.Account, err error, statusCode int, errorCode string)
	SignPayload(accountName string, payload []byte) (signedPayload []byte, err error, statusCode int, errorCode string)
	SignXdr(accountName string, idUnsigned []byte, idSigned []byte, transactionUnsigned []byte) (transactionSigned []byte,
		err error, statusCode int, errorCode string)
	ParticipantSignXdr(accountName string, transactionUnsigned []byte) (transactionSigned []byte,
		err error, statusCode int, errorCode string)
	AddIBMSign(transactionUnsigned []byte) (transactionSigned []byte,
		err error, statusCode int, errorCode string)
	GetIBMAccount() (account model.Account, err error, statusCode int, errorCode string)
}
