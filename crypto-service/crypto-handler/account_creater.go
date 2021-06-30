package crypto_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/response"
)

//add next
func (op *CryptoOperations) CreateAccount(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	accountName := vars["account_name"]

	if accountName == "" {
		err := errors.New("Missing required parameter: account_name")
		LOGGER.Errorf("Error: %v", err.Error())
		response.NotifyWWError(w, req, http.StatusBadRequest, "CRYPTO-0004", err)
		return
	}

	domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	account, err := op.secrets.GetAccount(domainId, accountName)
	if err != nil {
		LOGGER.Errorf("account: %v, %v, %v", account.NodeAddress, account.PrivateKeyLabel, account.PublicKeyLabel)
	}
	if account.NodeAddress != "" {
		//account already exists
		LOGGER.Errorf("Account already exists: %s", account.NodeAddress)
		response.NotifyWWError(w, req, http.StatusAlreadyReported, "CYRPTO-0004", errors.New("account already exists: "+account.NodeAddress))
		return
	}

	accountHSM, err := op.HSMInstance.GenericGenerateAccount()
	if err != nil {
		LOGGER.Errorf("Error: %v", err.Error())
		response.NotifyWWError(w, req, http.StatusFailedDependency, "CYRPTO-0004", err)
		return
	}

	LOGGER.Debugf("Account Generated: %+v", accountHSM)

	responseData, marshalErr := json.Marshal(accountHSM)
	if marshalErr != nil {
		response.NotifyWWError(w, req, http.StatusNotFound, "CYRPTO-0004", err)
		return
	}
	response.Respond(w, http.StatusCreated, responseData)
	return
}
