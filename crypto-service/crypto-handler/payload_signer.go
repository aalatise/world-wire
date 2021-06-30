package crypto_handler

import (
	"encoding/json"
	"net/http"
	"os"

	"github.ibm.com/gftn/world-wire-services/utility/common"
	"github.ibm.com/gftn/world-wire-services/utility/xmldsig"

	"github.com/go-openapi/strfmt"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/response"
)

func (op *CryptoOperations) SignXML(w http.ResponseWriter, req *http.Request) {
	var signRequest model.RequestPayload
	err := json.NewDecoder(req.Body).Decode(&signRequest)

	if err != nil {
		LOGGER.Errorf("Error occured while decoding the JSON request: %+v", err)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	err = signRequest.Validate(strfmt.Default)
	if err != nil {
		msg := "Unable to validate sign payload request: " + err.Error()
		LOGGER.Errorf(msg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	//Retrieve the public key handle and private key handle from environment variable
	//Environment variable will be injected from AWS during the startup
	domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	account, err := op.secrets.GetAccount(domainId, *signRequest.AccountName)
	if err != nil {
		LOGGER.Errorf("Error occured retrieving the account: %+v", err)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	//LOGGER.Infof("Request to sign: %s", signRequest.Payload.String())

	signedXML, err := op.HSMInstance.SignXML(signRequest.Payload.String(), account.PrivateKeyLabel, account.NodeAddress, false)
	if err != nil {
		LOGGER.Errorf("Error occured while siging the payload: %+v", err)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}
	//LOGGER.Infof("Signed XML:%s", signedXML)
	xmldsig.VerifySignature(signedXML)

	//XML has to be sent as base64 encoded
	//encodedSignedXML := base64.StdEncoding.EncodeToString([]byte(signedXML))
	signedPayload := model.PayloadWithSignature{}
	dataArray := strfmt.Base64{}
	dataArray = []byte(signedXML)
	signedPayload.PayloadWithSignature = &dataArray
	responseData, marshalErr := json.Marshal(signedPayload)
	if marshalErr != nil {
		LOGGER.Errorf("Error: %v", marshalErr.Error())
		response.NotifyWWError(w, req, http.StatusNotFound, "CRYPTO-0001", err)
		return
	}
	response.Respond(w, http.StatusOK, responseData)
	return
}

func (op *CryptoOperations) SignXMLUsingStellar(w http.ResponseWriter, req *http.Request) {
	var signRequest model.RequestPayload
	err := json.NewDecoder(req.Body).Decode(&signRequest)

	if err != nil {
		LOGGER.Errorf("Error occured while decoding the JSON request: %+v", err)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	err = signRequest.Validate(strfmt.Default)
	if err != nil {
		msg := "Unable to validate sign payload request: " + err.Error()
		LOGGER.Errorf(msg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	//Retrieve the public key handle and private key handle from environment variable
	//Environment variable will be injected from AWS during the startup
	domainId := os.Getenv(global_environment.ENV_KEY_IBM_TOKEN_DOMAIN_ID)
	account, err := op.secrets.GetAccount(domainId, common.MASTER_ACCOUNT)
	if err != nil {
		LOGGER.Errorf("Error occured retrieving the account: %+v", err)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	signedXML, err := op.HSMInstance.SignXML(signRequest.Payload.String(), account.NodeSeed, account.NodeAddress, true)
	if err != nil {
		LOGGER.Errorf("Error occured while siging the payload: %+v", err)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	xmldsig.VerifySignature(signedXML)

	//XML has to be sent as base64 encoded
	//encodedSignedXML := base64.StdEncoding.EncodeToString([]byte(signedXML))
	signedPayload := model.PayloadWithSignature{}
	dataArray := strfmt.Base64{}
	dataArray = []byte(signedXML)
	signedPayload.PayloadWithSignature = &dataArray
	responseData, marshalErr := json.Marshal(signedPayload)
	if marshalErr != nil {
		LOGGER.Errorf("Error: %v", marshalErr.Error())
		response.NotifyWWError(w, req, http.StatusNotFound, "CRYPTO-0001", err)
		return
	}
	response.Respond(w, http.StatusOK, responseData)
	return
}

func (op *CryptoOperations) SignPayload(w http.ResponseWriter, req *http.Request) {

	var payload model.RequestPayload
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		LOGGER.Errorf("Error: %v", err.Error())
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}
	err = payload.Validate(strfmt.Default)
	if err != nil {
		msg := "Unable to validate sign payload request: " + err.Error()
		LOGGER.Errorf(msg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0001", err)
		return
	}

	domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	account, err := op.secrets.GetAccount(domainId, *payload.AccountName)
	if err != nil {
		LOGGER.Errorf("Error: %v", err.Error())
		response.NotifyWWError(w, req, http.StatusFailedDependency, "CRYPTO-0005", err)
		return
	}
	signedPayload := model.Signature{}
	singedData, err := op.HSMInstance.GenericSignPayload([]byte(*payload.Payload), account)
	if err != nil {
		LOGGER.Errorf("Error: %v", err.Error())
		response.NotifyWWError(w, req, http.StatusFailedDependency, "CRYPTO-0006", err)
		return
	}

	dataArray := strfmt.Base64{}
	dataArray = singedData[:]
	signedPayload.TransactionSigned = &dataArray
	responseData, marshalErr := json.Marshal(signedPayload)
	if marshalErr != nil {
		LOGGER.Errorf("Error: %v", marshalErr.Error())
		response.NotifyWWError(w, req, http.StatusNotFound, "CRYPTO-0006", err)
		return
	}

	response.Respond(w, http.StatusOK, responseData)
	return
}
