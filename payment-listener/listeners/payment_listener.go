package listeners

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/gorilla/mux"
	"github.com/stellar/go/clients/horizon"
	"github.com/IBM/world-wire/gftn-models/model"
	pr_client "github.com/IBM/world-wire/participant-registry-client/pr-client"
	payment_utils "github.com/IBM/world-wire/payment-listener/utility"
	"github.com/IBM/world-wire/utility"
	comn "github.com/IBM/world-wire/utility/common"

	"github.com/IBM/world-wire/utility/database"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"github.com/IBM/world-wire/utility/global-environment/services/secrets"
	"github.com/IBM/world-wire/utility/global-environment/services/secrets/vault"
	"github.com/IBM/world-wire/utility/kafka"
	"github.com/IBM/world-wire/utility/response"
)

type PaymentListenerOperation struct {
	accountPaymentListeners map[string]StellarAccountPaymentListener
	PRClient                pr_client.RestPRServiceClient
	client                  *database.MongoClient
	Secrets                 secrets.Client
}

// var dynamoClient *payment_utils.DynamoClient
var kafkaOperations *kafka.KafkaOpreations

func CreatePaymentListenerOperation() PaymentListenerOperation {
	LOGGER.Infof("* Initializng payment listener operations")

	var err error
	op := PaymentListenerOperation{}
	op.accountPaymentListeners = make(map[string]StellarAccountPaymentListener)

	prClient, _ := pr_client.CreateRestPRServiceClient(os.Getenv(global_environment.ENV_KEY_PARTICIPANT_REGISTRY_URL))
	op.PRClient = prClient

	LOGGER.Infof("* Initializng Dynamo database client")

	op.client, err = database.InitializeIbmCloudConnection()
	if err != nil {
		utility.ExitOnErr(LOGGER, err, "Error establishing Mongo connection")
	}

	LOGGER.Infof("* Initializng Kafka producer")

	kafkaOperations, err = kafka.Initialize()
	if err != nil {
		utility.ExitOnErr(LOGGER, err, "Kafka initialized failed")
	}

	if strings.ToUpper(os.Getenv(global_environment.ENV_KEY_SECRET_STORAGE_LOCATION)) == comn.HASHICORP_VAULT_SECRET {
		op.Secrets, err = vault.InitializeVault()
		if err != nil {
			panic(err)
		}
	} else {
		panic("No valid secret storage location is specified")
	}

	return op

}

//CreatePaymentListeners - Set payment listeners for given array of operating accounts
func (op *PaymentListenerOperation) CreatePaymentListeners(accounts []model.Account) {

	httpClientWithTimeout := http.Client{
		Timeout: 60 * time.Second,
	}
	hc := horizon.Client{
		URL:  os.Getenv(global_environment.ENV_KEY_HORIZON_CLIENT_URL),
		HTTP: &httpClientWithTimeout,
	}
	for _, account := range accounts {
		accountListener, err := NewStellarAccountStellarAccountPaymentListener(&hc)
		if err != nil {
			LOGGER.Errorf("Not able to set up payment receiver")
		}
		accountListener.SetDistAccount(account)
		cursor, err := payment_utils.GetAccountCursor(op.client, account.Name)
		if err != nil || cursor.Cursor == "" {
			LOGGER.Errorf("Could not load last cursor from DB: %v", err)

			//We may need to send out email notification of service being down here.
			// If no last cursor saved set it to: `now`
			accountListener.cursor = horizon.Cursor("now")
			err = payment_utils.AddAccountCursor(op.client, account.Name, "now")
			if err != nil {
				LOGGER.Errorf(err.Error())
				return
			}
		} else {
			accountListener.cursor = horizon.Cursor(cursor.Cursor)
		}
		LOGGER.Infof("In CreatePaymentListeners Account: %v, Cursor: %v", account.Name, cursor)
		err = accountListener.Listen(accountListener.channel)
		if err != nil {
			LOGGER.Errorf("Error CreatePaymentListeners Account: %v, Cursor: %v", account.Name, cursor)
		}
		op.accountPaymentListeners[account.Name] = accountListener
	}
}

func (op *PaymentListenerOperation) ReStartListener(w http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)
	accountName := vars["account_name"]
	cursor := vars["cursor"]

	if accountName == "" || cursor == "" {
		err := errors.New("Missing Parameter: account_name and or cursor")
		response.NotifyWWError(w, request, http.StatusBadRequest, "API-1041", err)
		return
	}

	//Save the received cursor and restart listener
	domainId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	account, err := op.Secrets.GetAccount(domainId, accountName)
	msg := ""
	if err == nil && account.NodeAddress != "" {
		if accountListener, ok := op.accountPaymentListeners[accountName]; ok {

			accountListener.cursor = horizon.Cursor(cursor)
			err = payment_utils.UpdateAccountCursor(op.client, accountListener.account.Name, cursor)
			if err != nil {
				LOGGER.Error("Error updating cursor to Dynamo DB for account %v: %v", accountListener.account.Name, err.Error())
				return
			}

			LOGGER.Infof("In ReStartListener Account: %v, Cursor: %v", accountName, cursor)
			go func() {
				//time.Sleep(1*time.Second)
				accountListener.channel <- "restart"
			}()
			msg = "Restarted listener at cursor:" + cursor
			response.NotifySuccess(w, request, msg)
			return
		} else {
			distAccount := model.Account{}
			distAccount.Name = accountName
			distAccount.Address = &account.NodeAddress
			accounts := []model.Account{distAccount}
			op.CreatePaymentListeners(accounts)
			msg = "Started listener at cursor:" + cursor
			response.NotifySuccess(w, request, msg)
			return
		}
	}
	response.NotifyWWError(w, request, http.StatusConflict, "API-1124", errors.New(accountName))
}

// GetParticipantOperatingAccounts: Extracts all registered operating accounts for a participant from participant registry
// Alternatively this can also be extracted from secrets store/vault/Local storage directly
func (op *PaymentListenerOperation) GetParticipantOperatingAccounts() (accounts []model.Account, err error) {
	LOGGER.Debugf("getParticipantOperatingAccounts")

	participantObj, err := op.PRClient.GetParticipantForDomain(os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME))
	if err != nil {
		LOGGER.Errorf(" Error GetParticipantForDomain failed: %v", err)
		return accounts, err
	}

	for _, account := range participantObj.OperatingAccounts {
		accounts = append(accounts, *account)
	}

	return accounts, err
}
