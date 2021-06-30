package cryptoservice

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	b "github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
)

type MockClient struct {
	HTTP *http.Client
}

func (mcsc *MockClient) RequestSigning(txeBase64 string, requestBase64 string, signedRequestBase64 string, accountName string, participant model.Participant) (string, error) {
	txeBase64R := strings.NewReader(txeBase64)
	txeByteR := base64.NewDecoder(base64.StdEncoding, txeBase64R)
	var txe xdr.TransactionEnvelope
	xdr.Unmarshal(txeByteR, &txe)
	txeb := &b.TransactionEnvelopeBuilder{E: &txe}
	txeb.Init()
	txeb.MutateTX(b.TestNetwork)
	if *participant.ID == "hk.one.payments.worldwire.io" {
		nodeseed := "SC5HQEXEYBYFQ5ZRJ6IVPWW24EBKZBAYM565JBSXQDI4RGDSYBR72EZE"
		sig := b.Sign{Seed: nodeseed}
		err := txeb.Mutate(sig)
		if err != nil {
			LOGGER.Error(err)
			return "", err
		}
		txeB64, _ := txeb.Base64()
		return txeB64, nil
	}
	if *participant.ID == "ie.one.payments.worldwire.io" {
		nodeseed := "SBNOGNMWZVHZNM2L5NK3BZFYI7SR5QWO4YIFYTETSKSU6X3K7RJD7DEM"
		sig := b.Sign{Seed: nodeseed}
		err := txeb.Mutate(sig)
		if err != nil {
			LOGGER.Error(err)
			return "", err
		}
		txeB64, _ := txeb.Base64()
		return txeB64, nil
	}
	return "", errors.New("error requesting crypto service to sign XDR")
}
