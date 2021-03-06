package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/go-openapi/strfmt"
	au "github.ibm.com/gftn/world-wire-services/anchor-service/anchor-util"
	apiutil "github.ibm.com/gftn/world-wire-services/api-service/utility"
	middlewares "github.ibm.com/gftn/world-wire-services/auth-service-go/middleware"
	gasserviceclient "github.ibm.com/gftn/world-wire-services/gas-service-client"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	pr_client "github.ibm.com/gftn/world-wire-services/participant-registry-client/pr-client"
	"github.ibm.com/gftn/world-wire-services/utility/common"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/global-environment/services/secrets"
	"github.ibm.com/gftn/world-wire-services/utility/global-environment/services/secrets/vault"
	participant_checks "github.ibm.com/gftn/world-wire-services/utility/participant"
	"github.ibm.com/gftn/world-wire-services/utility/response"
)

type TrustHandler struct {
	prClient         pr_client.RestPRServiceClient
	VaultSession     secrets.Client
	GasServiceClient gasserviceclient.GasServiceClient
}

func CreateTrustHandler() (TrustHandler, error) {
	th := TrustHandler{}
	var err error
	if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == common.HASHICORP_VAULT_SECRET {
		th.VaultSession, err = vault.InitializeVault()
		if err != nil {
			panic(err)
		}
	} else {
		panic("No valid secret storage location is specified")
	}

	prClient, err := pr_client.CreateRestPRServiceClient(os.Getenv(global_environment.ENV_KEY_PARTICIPANT_REGISTRY_URL))
	if err != nil {
		LOGGER.Errorf(" Error getParticipantForDomain CreateRestPRServiceClient failed  %v", err)
		return th, err
	}
	th.prClient = prClient

	gasServiceClient := gasserviceclient.Client{
		HTTP: &http.Client{Timeout: time.Second * 20},
		URL:  os.Getenv(global_environment.ENV_KEY_GAS_SVC_URL),
	}
	th.GasServiceClient = &gasServiceClient

	return th, nil
}

func (th TrustHandler) AllowTrust(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

	LOGGER.Infof("Handling Allow Trust request")

	vars := mux.Vars(req)
	anchorID := vars["anchor_id"]

	if anchorID == "" {
		LOGGER.Warningf("anchor_id is required")
		response.NotifyWWError(w, req, http.StatusBadRequest, "ANCHOR-0012", nil)
		return
	}

	//Check JWT token
	if os.Getenv(global_environment.ENV_KEY_ENABLE_JWT) != "false" {
		participantID, err := middlewares.GetIdentity(req)
		//Check if requesting anchor id is same as participant id in the token
		if participantID != anchorID {
			response.NotifyWWError(w, req, http.StatusUnauthorized, "ANCHOR-0067",
				err)
			return
		}
	}

	var trustRequest model.Trust
	err := json.NewDecoder(req.Body).Decode(&trustRequest)
	if err != nil {
		LOGGER.Debugf("Error  %v", err.Error())
		response.NotifyWWError(w, req, http.StatusNotFound, "ANCHOR-0013", nil)
		return
	}
	err = trustRequest.Validate(strfmt.Default)
	if err != nil {
		LOGGER.Debugf("Error  %v", err.Error())
		response.NotifyWWError(w, req, http.StatusNotFound, "ANCHOR-0014", nil)
		return
	}

	permission := *trustRequest.Permission
	assetCode := *trustRequest.AssetCode

	err = model.IsValidDACode(assetCode)
	if err != nil {
		LOGGER.Debug("AllowTrust:", err)
		response.NotifyWWError(w, req, http.StatusBadRequest, "ANCHOR-0015", err)
		return
	}
	//**************************

	participant, err := au.GetParticipantForDomain(*trustRequest.ParticipantID)
	if err != nil {
		LOGGER.Warningf("Participant Domain does not exist")
		response.NotifyWWError(w, req, http.StatusNotFound, "ANCHOR-0009", nil)
		return
	}

	// Check if participant is active
	LOGGER.Info("Check participant active")
	err = participant_checks.CheckStatusActive(participant)
	if err != nil {
		msg := err.Error()
		LOGGER.Error(msg)
		response.NotifyFailure(w, req, http.StatusBadRequest, msg)
		return
	}

	participantAddress := au.GetAccountAddressForParticipant(participant, *trustRequest.AccountName)
	if participantAddress == "" {
		LOGGER.Warningf("Invalid Account: %v", *trustRequest.AccountName)
		response.NotifyWWError(w, req, http.StatusNotFound, "ANCHOR-0010", nil)
		return
	}

	anchor, err := au.GetParticipantForDomain(anchorID)
	if err != nil {
		LOGGER.Warningf("Anchor Domain does not exist")
		response.NotifyWWError(w, req, http.StatusNotFound, "ANCHOR-0020", nil)
		return
	}

	anchorAddress := anchor.IssuingAccount
	if anchorAddress == "" {
		LOGGER.Warningf("Invalid Anchor Account: %v", *trustRequest.AccountName)
		response.NotifyWWError(w, req, http.StatusNotFound, "ANCHOR-0010", nil)
		return
	}
	errMsg := ""
	if permission == "allow" {
		err, errMsg = au.AllowTrustForDigitalAsset(th.GasServiceClient, participantAddress, anchorAddress, assetCode,
			true, th.VaultSession)
		if err != nil {
			LOGGER.Warningf("Allow Trust Failed: %v", errMsg)
			response.NotifyWWError(w, req, http.StatusConflict, "ANCHOR-0021", errors.New(errMsg))
			return
		}
		response.NotifySuccess(w, req, "Allow Trust is successful")
		return
	} else if permission == "revoke" {
		err, errMsg = au.AllowTrustForDigitalAsset(th.GasServiceClient, participantAddress, anchorAddress, assetCode,
			false, th.VaultSession)
		if err != nil {
			LOGGER.Warningf("Revoke Trust Failed: %v", errMsg)
			response.NotifyWWError(w, req, http.StatusConflict, "ANCHOR-0021", errors.New(errMsg))
			return
		}
		response.NotifySuccess(w, req, "Revoke Trust is successful")

		return
	}

	response.NotifyWWError(w, req, http.StatusBadRequest, "ANCHOR-0011", nil)
	return

}

func (th TrustHandler) GetIssuedAssets(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

	LOGGER.Infof("Handling get issued assets")
	vars := mux.Vars(req)
	anchorID := vars["anchor_id"]

	if anchorID == "" {
		LOGGER.Warningf("anchor_id is required")
		response.NotifyWWError(w, req, http.StatusBadRequest, "ANCHOR-0012", nil)
		return
	}

	//Check JWT token
	if os.Getenv(global_environment.ENV_KEY_ENABLE_JWT) != "false" {
		participantID, err := middlewares.GetIdentity(req)
		//Check if requesting anchor id is same as participant id in the token
		if participantID != anchorID {
			response.NotifyWWError(w, req, http.StatusUnauthorized, "ANCHOR-0067",
				err)
			return
		}
	}
	var assets []*model.Asset

	//Get IBM token account from nc, vault or AWS secret mngr
	domainId := os.Getenv(global_environment.ENV_KEY_IBM_TOKEN_DOMAIN_ID)
	wwAdminAccount, err := th.VaultSession.GetAccount(domainId, common.MASTER_ACCOUNT)
	if err != nil {
		response.NotifyWWError(w, req, http.StatusUnauthorized, "ANCHOR-0078",
			err)
		return
	}

	LOGGER.Infof("IBM Token Account: %s", wwAdminAccount)

	if strings.TrimSpace(wwAdminAccount.NodeAddress) == "" {
		response.NotifyWWError(w, req, http.StatusUnauthorized, "ANCHOR-0078",
			err)
		return
	}

	//Get trusted assets by IBM account
	wwAssets, err := apiutil.GetAssets(wwAdminAccount.NodeAddress, th.prClient)
	if err != nil {
		response.NotifyWWError(w, req, http.StatusUnauthorized, "ANCHOR-0079",
			err)
		return
	}
	LOGGER.Debug("%v", wwAssets)

	if len(wwAssets) > 0 {
		for _, ast := range wwAssets {
			if ast.IssuerID != "" && ast.IssuerID == anchorID {
				assets = append(assets, ast)
				continue
			} else {
				continue
			}
		}
	}

	if assets == nil || len(assets) == 0 {
		newError := errors.New(anchorID)
		LOGGER.Debug("No Issued assets found for: %v ", anchorID)
		response.NotifyWWError(w, req, http.StatusNotFound, "ANCHOR-0080", newError)
		return
	}
	assetBytes, _ := json.Marshal(assets)
	response.Respond(w, http.StatusOK, assetBytes)

}
