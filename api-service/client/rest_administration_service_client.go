package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/go-resty/resty"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
)

type RestAdministrationServiceClient struct {
	AdministrationServiceURL string
}

func CreateRestAdministrationServiceClient() (RestAdministrationServiceClient, error) {

	client := RestAdministrationServiceClient{}
	url := os.Getenv(global_environment.ENV_KEY_ADMIN_SVC_URL)
	if url == "" {
		LOGGER.Errorf("You MUST set the %v environment variable to point to Administration Service URL", global_environment.ENV_KEY_ADMIN_SVC_URL)
		os.Exit(1)
	}
	client.AdministrationServiceURL = url
	return client, nil

}

func (client RestAdministrationServiceClient) GetTxnDetails(txnDetailsRequest model.FItoFITransactionRequest) ([]byte, int, error) {

	url := client.AdministrationServiceURL + "/internal/transaction"
	LOGGER.Infof("Doing internal administration service call:  %v", url)
	bodyBytes, err := json.Marshal(&txnDetailsRequest)
	response, err := resty.R().SetBody(bodyBytes).SetHeader("Content-type", "application/json").Post(url)
	if err != nil {
		LOGGER.Errorf("Error with calling administration service API:  %v", err.Error())
		return []byte{}, http.StatusNotFound, err
	}

	if response.StatusCode() != http.StatusOK {
		LOGGER.Warningf("Internal Administration Service returned a non-200 status (%v)", response.StatusCode())
		return []byte{}, response.StatusCode(), errors.New(string(response.Body()[:]))
	}

	responseBody := response.Body()

	return responseBody, http.StatusOK, nil

}
