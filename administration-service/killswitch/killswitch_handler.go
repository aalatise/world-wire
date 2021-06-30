package killswitch

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.ibm.com/gftn/world-wire-services/utility/global-environment/services/secrets"

	"github.com/gorilla/mux"
	b "github.com/stellar/go/build"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
	crypto_client "github.ibm.com/gftn/world-wire-services/crypto-service-client/crypto-client"
	gasserviceclient "github.ibm.com/gftn/world-wire-services/gas-service-client"
	ast "github.ibm.com/gftn/world-wire-services/utility/asset"
	secret_manager "github.ibm.com/gftn/world-wire-services/utility/aws/golang/secret-manager"
	"github.ibm.com/gftn/world-wire-services/utility/aws/golang/utility"
	"github.ibm.com/gftn/world-wire-services/utility/common"
	util "github.ibm.com/gftn/world-wire-services/utility/common"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/global-environment/services/secrets/vault"
	"github.ibm.com/gftn/world-wire-services/utility/participant"
	"github.ibm.com/gftn/world-wire-services/utility/response"
)

type KillSwitch struct {
	secrets secrets.Client
}

type Operations struct {
	GasServiceClient    gasserviceclient.GasServiceClient
	CryptoServiceClient crypto_client.CryptoServiceClient
}

/*
CreateKillSwitch is a Context initializing function
*/
func CreateKillSwitch() (KillSwitch, error) {
	ks := KillSwitch{}
	var err error
	if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == common.HASHICORP_VAULT_SECRET {
		ks.secrets, err = vault.InitializeVault()
		if err != nil {
			panic(err)
		}
	} else {
		panic("No valid secret storage location is specified")
	}

	return ks, nil
}

/*
SuspendAccount function would be called by passing a Stellar Address in the request parameter
and the account passed will be suspended by removing SHA256 signer from the signer list,
making master key's weigh to 0. At the same time, this function would modify the Threshold
value to [1,1,1]. So that IBM account which is in the signer list with weight 1 could do
all the operations.
*/
func (ks KillSwitch) SuspendAccount(w http.ResponseWriter, req *http.Request) {
	LOGGER.Info("Kill Switch has activated a Suspend Account function")
	vars := mux.Vars(req)
	accountName := vars["account_name"]
	participantId := vars["participant_id"]

	if len(accountName) == 0 || !ValidateAccount(accountName) {
		LOGGER.Error("Stellar account is invalid or Empty.")
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0009", errors.New("stellar account is invalid or Empty"))
		return
	}

	if len(participantId) == 0 {
		LOGGER.Error("Participant ID is Empty.")
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0009", errors.New("Participant ID is Empty"))
		return
	}

	LOGGER.Debug("Account to be suspended: ", accountName)

	horizonClient := util.GetHorizonClient(os.Getenv(global_environment.ENV_KEY_HORIZON_CLIENT_URL))
	stellarNetwork := util.GetStellarNetwork(os.Getenv(global_environment.ENV_KEY_STELLAR_NETWORK))

	// retrieve killswitch string from vault
	secretString, err := ks.secrets.GetSecretPhrase(participantId, accountName)
	if err != nil {
		errorMsg := "Encounter error when retrieving secret phrase from AWS"
		LOGGER.Error(errorMsg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0018", errors.New(errorMsg))
		return
	}

	op := Operations{}
	gasServiceClient := gasserviceclient.Client{
		HTTP: &http.Client{Timeout: time.Second * 20},
		URL:  os.Getenv(global_environment.ENV_KEY_GAS_SVC_URL),
	}
	op.GasServiceClient = &gasServiceClient
	gasAccount, sequenceNum, err := op.GasServiceClient.GetAccountAndSequence()
	if gasAccount == "" || sequenceNum == 0 {
		errMsg := "IBM Gas account failed to load"
		LOGGER.Errorf(errMsg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0021", errors.New(errMsg))
		return
	}

	// retrieve IBM token account
	domainId := os.Getenv(global_environment.ENV_KEY_IBM_TOKEN_DOMAIN_ID)
	wwAdminAccount, err := ks.secrets.GetAccount(domainId, common.MASTER_ACCOUNT)
	if err != nil {
		LOGGER.Debugf("IBM account: %v", wwAdminAccount.NodeAddress)
		LOGGER.Errorf("Error: %v", err.Error())
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	tx, err := b.Transaction(
		b.SourceAccount{AddressOrSeed: gasAccount},
		b.AutoSequence{SequenceProvider: &horizonClient},
		stellarNetwork,
		b.SetOptions(
			b.SourceAccount{AddressOrSeed: accountName},
			b.MasterWeight(0),
			b.SetThresholds(1, 1, 1),
			b.RemoveSigner(generateSHA256Hash(secretString)),
		),
	)

	if err != nil {
		LOGGER.Errorf("Error creating SetOptions stellar transaction: %v", err)
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	txe, err := tx.Sign(wwAdminAccount.NodeSeed)
	if err != nil {
		LOGGER.Errorf("Error signing SetOptions stellar transaction")
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	signaturearr := make([]xdr.DecoratedSignature, len(txe.E.Signatures)+1)

	for i, element := range txe.E.Signatures {
		signaturearr[i] = element
	}

	ds0 := xdr.DecoratedSignature{
		Hint:      xdr.SignatureHint(getHint([]byte(secretString))),
		Signature: xdr.Signature([]byte(secretString)),
	}

	txe.E.Signatures = append(txe.E.Signatures, ds0)

	txeB64, err := txe.Base64()

	if err != nil {
		LOGGER.Errorf("Error getting SetOptions transaction xdr: %v", err)
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}
	LOGGER.Infof("TxeB64: %v", txeB64)

	LOGGER.Debug("txeB64: ", txeB64)

	//Submit on Gas service
	_, _, err = op.GasServiceClient.SubmitTxe(txeB64)
	if err != nil {
		err = ast.DecodeStellarError(err)
		errMsg := "CreateAccount Failed. Error Communicating with Stellar." + err.Error()
		LOGGER.Error(errMsg)
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	LOGGER.Infof("Account has been suspended successfully.")

	credential := secrets.CredentialInfo{
		Environment: os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION),
		Domain:      participantId,
		Service:     "killswitch",
		Variable:    "accounts",
	}
	err = ks.secrets.DeleteSingleSecretEntry(credential, accountName)
	if err != nil {
		errMsg := "Encounter error while deleting secret phrase of account after getting suspend"
		LOGGER.Errorf(errMsg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0018", errors.New(errMsg))
		return
	}

	response.NotifySuccess(w, req, "Account has been suspended successfully.")
}

/*
ReactivateAccount function would be called by passing a suspended Stellar Address in the request parameter
and the account passed will be reactivated by adding SHA256 signer to the signer list,
making master key's weigh to 2. At the same time, this function would modify the Threshold
value to [1,2,3]. So that participant could do regular operations.
*/
func (ks KillSwitch) ReactivateAccount(w http.ResponseWriter, req *http.Request) {
	LOGGER.Info("Kill Switch has activated a Reactivate Account function")

	vars := mux.Vars(req)
	accountName := vars["account_name"]
	participantId := vars["participant_id"]

	if len(accountName) == 0 || !ValidateAccount(accountName) {
		LOGGER.Error("Stellar account is invalid or Empty.")
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0009", errors.New("stellar account is invalid or Empty"))
		return
	}

	if len(participantId) == 0 {
		LOGGER.Error("Participant ID is Empty.")
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0009", errors.New("Participant ID is Empty"))
		return
	}

	LOGGER.Debug("Account to be activated: ", accountName)
	horizonClient := util.GetHorizonClient(os.Getenv(global_environment.ENV_KEY_HORIZON_CLIENT_URL))
	stellarNetwork := util.GetStellarNetwork(os.Getenv(global_environment.ENV_KEY_STELLAR_NETWORK))

	secretString := participant.GetSecretPhrase()

	op := Operations{}
	gasServiceClient := gasserviceclient.Client{
		HTTP: &http.Client{Timeout: time.Second * 20},
		URL:  os.Getenv(global_environment.ENV_KEY_GAS_SVC_URL),
	}
	op.GasServiceClient = &gasServiceClient
	gasAccount, sequenceNum, err := op.GasServiceClient.GetAccountAndSequence()
	if gasAccount == "" || sequenceNum == 0 {
		errMsg := "IBM Gas account failed to load"
		LOGGER.Errorf(errMsg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0021", errors.New(errMsg))
		return
	}

	// retrieve IBM token account
	domainId := os.Getenv(global_environment.ENV_KEY_IBM_TOKEN_DOMAIN_ID)
	wwAdminAccount, err := ks.secrets.GetAccount(domainId, common.MASTER_ACCOUNT)
	if err != nil {
		LOGGER.Debugf("IBM account: %v", wwAdminAccount.NodeAddress)
		LOGGER.Errorf("Error: %v", err.Error())
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	tx, err := b.Transaction(
		b.SourceAccount{AddressOrSeed: gasAccount},
		b.AutoSequence{SequenceProvider: &horizonClient},
		stellarNetwork,
		b.SetOptions(
			b.SourceAccount{accountName},
			b.MasterWeight(2),
			b.AddSigner(wwAdminAccount.NodeAddress, 1),
			b.AddSigner(generateSHA256Hash(secretString), 2),
			b.SetThresholds(1, 2, 3),
		),
	)

	if err != nil {
		LOGGER.Errorf("Error creating SetOptions stellar transaction")
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	txe, err := tx.Sign(wwAdminAccount.NodeSeed)
	if err != nil {
		LOGGER.Errorf("Error signing SetOptions stellar transaction")
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		LOGGER.Errorf("Error getting SetOptions transaction xdr: %v", err)
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}
	LOGGER.Infof("TxeB64: %v", txeB64)

	//Submit on Gas service
	_, _, err = op.GasServiceClient.SubmitTxe(txeB64)
	if err != nil {
		err = ast.DecodeStellarError(err)
		errMsg := "CreateAccount Failed. Error Communicating with Stellar." + err.Error()
		LOGGER.Error(errMsg)
		response.NotifyWWError(w, req, http.StatusInternalServerError, "ADMIN-0008", err)
		return
	}

	LOGGER.Infof("Account has been activated successfully.")

	credential := secrets.CredentialInfo{
		Environment: os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION),
		Domain:      participantId,
		Service:     "killswitch",
		Variable:    "accounts",
	}

	secretInput := make(map[string]interface{})
	secretInput[accountName] = secretString

	err = ks.secrets.AppendSecret(credential, secretInput)
	if err != nil {
		errMsg := "Encounter error while accessing AWS secret manager"
		LOGGER.Errorf(errMsg)
		response.NotifyWWError(w, req, http.StatusBadRequest, "ADMIN-0018", errors.New(errMsg))
		return
	}
	response.NotifySuccess(w, req, "Account has been activated successfully.")
}

/*
 This function would accept one string value as argument and generate SHA256 value
 key with Stellar's HashX encoding. This would be used as signer of an account.
*/
func generateSHA256Hash(key string) string {
	hasher := sha256.New()
	hasher.Write([]byte(key))

	actual, err := strkey.Encode(strkey.VersionByteHashX, hasher.Sum(nil))
	if err != nil {
		LOGGER.Fatal(err)
		return ""
	}
	return actual
}

// get this value from aws
func getSecretPhrase(accountName, participantId string) (string, error) {

	if os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION) == "" {
		errorMsg := "Please set environment variables: ENV_VERSION & HOME_DOMAIN_NAME correctly"
		LOGGER.Errorf(errorMsg)
		return "", errors.New(errorMsg)
	}

	credential := utility.CredentialInfo{
		Environment: os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION),
		Domain:      participantId,
		Service:     "killswitch",
		Variable:    "accounts",
	}

	return secret_manager.GetSingleSecretEntry(credential, accountName)
}

/*
This function would take 'preimage' value as argument and get the
last four bytes of the preimage to use as hint of 'DecoratedSignature'
*/
func getHint(publickey []byte) (r [4]byte) {
	hasher := sha256.New()
	hasher.Write(publickey)

	bytekey := hasher.Sum(nil)
	hint := bytekey[len(bytekey)-4:]

	copy(r[:], hint)

	return
}

/*
ValidateAccount is a utility function to validate the argument passed as parameter
is a valid type of Stellar key according to the Stellar's key encoding.
*/
func ValidateAccount(account string) bool {
	versionByte, err := strkey.Version(account)

	if err != nil {
		LOGGER.Errorf("Stellar Account validation failed: %v", err)
		return false
	}
	LOGGER.Infof("Version Byte: ", versionByte)
	errorKey := func(vByte strkey.VersionByte) error {
		if vByte == strkey.VersionByteAccountID {
			return nil
		}
		if vByte == strkey.VersionByteSeed {
			return nil
		}
		if vByte == strkey.VersionByteHashTx {
			return nil
		}
		if vByte == strkey.VersionByteHashX {
			return nil
		}
		return strkey.ErrInvalidVersionByte
	}(versionByte)

	if errorKey != nil {
		LOGGER.Errorf("Stellar Account validation failed: %v", err)
		return false
	}

	return true

}
