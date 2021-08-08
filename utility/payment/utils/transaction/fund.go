package transaction

import (
	"errors"
	"github.com/IBM/world-wire/utility/nodeconfig"
	"github.com/IBM/world-wire/utility/secrets/vault"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	gasserviceclient "github.com/IBM/world-wire/gas-service-client"
	"github.com/IBM/world-wire/utility/payment/client"
	"github.com/IBM/world-wire/utility/payment/constant"
	"github.com/IBM/world-wire/utility/payment/utils/sendmodel"
	"github.com/IBM/world-wire/utility/payment/utils/signing"
	"github.com/IBM/world-wire/utility/common"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"github.com/stellar/go/xdr"
)

type CreateFundingOpereations struct {
	gasServiceURL string
	admClient     *client.RestAdministrationServiceClient
	homeDomain    string
	serviceName   string
	prServiceURL  string
	signHandler   signing.CreateSignOperations
	GasClient     gasserviceclient.Client
	secrets   nodeconfig.Client
}

func InitiateFundingOperations(pr, domain string) (op CreateFundingOpereations) {
	op.gasServiceURL = os.Getenv(global_environment.ENV_KEY_GAS_SVC_URL)
	op.homeDomain = domain
	//op.serviceName = os.Getenv(global_environment.ENV_KEY_SERVICE_NAME)
	op.admClient, _ = client.CreateRestAdministrationServiceClient()
	op.signHandler = signing.InitiateSignOperations(pr)
	op.prServiceURL = pr

	op.GasClient = gasserviceclient.Client{
		HTTP: &http.Client{Timeout: time.Second * 80},
		URL:  op.gasServiceURL,
	}
	var err error
	if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == common.HASHICORP_VAULT_SECRET {
		op.secrets, err = vault.InitializeVault()
		if err != nil {
			panic(err)
		}
	} else {
		panic("No valid secret storage location is specified")
	}

	return op
}

func (op *CreateFundingOpereations) FundAndSubmitPaymentTransaction(rfiAccount, instructionId, xmlMsgType, rfiSettlementAccountName string, dbData sendmodel.SignData, memoHash xdr.Memo) (int, string, string) {
	var sendingAccount, receivingAccount, settlementAccountName, ofiAccount string

	switch xmlMsgType {
	case constant.IBWF001:
		// if OFI receive a ibwf001 message, the transaction sender will be OFI and receiver will be RFI and will use OFI settlement account to sign the transaction
		receivingAccount = rfiAccount
		settlementAccountName = dbData.SettlementAccountName

		domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
		account, getErr := op.secrets.GetAccount(domainId, dbData.SettlementAccountName)
		if getErr != nil {
			LOGGER.Error("Failed to get OFI account address from AWS secret manager")
			return constant.STATUS_CODE_INTERNAL_ERROR, "", ""
		}
		ofiAccount = account.NodeAddress
		sendingAccount = ofiAccount
	case constant.PACS004:
		// if RFI receive a pacs004 message, the transaction sender will be RFI and receiver will be OFI and will use RFI settlement account to sign the transaction
		sendingAccount = rfiAccount
		settlementAccountName = rfiSettlementAccountName
		account := client.GetParticipantAccount(op.prServiceURL, dbData.OFIId, dbData.SettlementAccountName)
		if account == nil {
			LOGGER.Error("Failed to get OFI account address from PR")
			return constant.STATUS_CODE_INTERNAL_ERROR, "", ""
		}
		receivingAccount = *account
	case constant.PACS002:
		domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
		account, getErr := op.secrets.GetAccount(domainId, dbData.SettlementAccountName)
		if getErr != nil {
			LOGGER.Error("Failed to get OFI account address from AWS secret manager")
			return constant.STATUS_CODE_INTERNAL_ERROR, "", ""
		}
		sendingAccount = account.NodeAddress
		receivingAccount = rfiSettlementAccountName
		settlementAccountName = dbData.SettlementAccountName
	}

	LOGGER.Infof("Get IBM account and sequence number from the gas service")
	ibmAccount, seqNum, gasErr := op.GasClient.GetAccountAndSequence()
	if gasErr != nil || ibmAccount == "" {
		if ibmAccount == "" {
			LOGGER.Errorf("Unable to retrieve gas account from gas service")
		} else {
			LOGGER.Errorf("IBM account: %s, Seq: %d", ibmAccount, seqNum)
		}
		if gasErr != nil {
			LOGGER.Errorf("Failed to get IBM account and tx sequence: %s", gasErr.Error())
		}
		return constant.STATUS_CODE_INTERNAL_ERROR, "", ""
	}

	LOGGER.Infof("Create Stellar transaction")
	signedTx, txErr := op.createStellarTransaction(ibmAccount, sendingAccount, receivingAccount, settlementAccountName, dbData, seqNum, memoHash)
	if txErr != nil {
		LOGGER.Errorf("Failed to create Stellar Transaction: %s", txErr.Error())
		return constant.STATUS_CODE_INTERNAL_ERROR, "", ""
	}

	var retryInterval time.Duration
	var retryTimes float64

	// default value of retry time interval is 5 seconds
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL); !exists {
		retryInterval = time.Duration(5)
	} else {
		temp, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL))
		retryInterval = time.Duration(temp)
	}

	// default value of retry times is 4
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES); !exists {
		retryTimes = 4
	} else {
		retryTimes, _ = strconv.ParseFloat(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES), 64)
	}

	var txHash string
	LOGGER.Infof("Submit Stellar transaction")
	err := common.Retry(retryTimes, retryInterval*time.Second, func() error {
		var submitErr error
		txHash, submitErr = op.submitToStellar(signedTx)
		if submitErr != nil {
			return submitErr
		}
		if txHash == "" {
			return errors.New("Failed submitting tx to gas-service")
		}
		return nil
	})
	if err != nil {
		LOGGER.Errorf("Failed to submit Transaction to Stellar: %s", err.Error())
		return constant.STATUS_CODE_INTERNAL_ERROR, "", ""
	}

	LOGGER.Infof("Successfully submit transaction to Stellar network. Instrucion ID: %v, TX Hash: %v", instructionId, txHash)

	return constant.STATUS_CODE_TX_SEND_TO_STELLAR, txHash, ofiAccount
}

//func (op *CreateFundingOpereations) RecordTimeLogsToKafka(action, task string, timeStamp time.Time) {
//	logs := fmt.Sprintf("%s-[%s]:[%s][%s][Time:%s]", action, op.homeDomain, op.serviceName, task, time.Since(timeStamp).String())
//	postgres.CreateServiceLogInDB(op.serviceName, logs)
//}
