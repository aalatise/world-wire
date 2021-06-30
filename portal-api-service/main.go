package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.ibm.com/gftn/world-wire-services/portal-api-service/environment"
	"github.ibm.com/gftn/world-wire-services/portal-api-service/middleware"
	"github.ibm.com/gftn/world-wire-services/portal-api-service/portalops"
	"github.ibm.com/gftn/world-wire-services/utility"
	"github.ibm.com/gftn/world-wire-services/utility/database"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/global-environment/services"
	"github.ibm.com/gftn/world-wire-services/utility/logconfig"
	"github.ibm.com/gftn/world-wire-services/utility/message"
	"github.ibm.com/gftn/world-wire-services/utility/response"
)

type app struct {
	Router         *mux.Router
	ServiceVersion string
	AccountOps     portalops.AccountOps
	BlocklistOps   portalops.BlocklistOps
	WhitelistOps   portalops.WhitelistOps
	InstitutionOps portalops.InstitutionOps
	TrustOps       portalops.TrustOps
	AssetOps       portalops.AssetOps
	KillswitchOps  portalops.KillswitchOps
	SuperOps       portalops.SuperOps
	ParticipantOps portalops.ParticipantOps
	UserOps        portalops.UserOps
	JwtOps         portalops.JwtOps
	ExchangeOps    portalops.TxOps
	TransferOps    portalops.TxOps
	Authentication middleware.Authentication
	Authorization  middleware.Authorization
}

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("portal-api-service")
}

func (app *app) initRoutes() {

	app.Router = mux.NewRouter()

	healthSubrouter := app.Router.PathPrefix("/" + app.ServiceVersion + "/portal").Subrouter()
	portalSubrouter := app.Router.PathPrefix("/" + app.ServiceVersion + "/portal").Subrouter()

	superSubrouter := portalSubrouter.PathPrefix("").Subrouter()
	participantSubrouter := portalSubrouter.PathPrefix("").Subrouter()
	SPSubrouter := portalSubrouter.PathPrefix("").Subrouter()

	logger.Infof("\t* Portal API:  Check service")
	healthSubrouter.HandleFunc("/service_check", portalops.StatusCheck).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Get institutions")
	superSubrouter.HandleFunc("/institutions", app.InstitutionOps.GetInstitutions).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Get institution")
	SPSubrouter.HandleFunc("/institutions/{institutionIdOrSlug}", app.InstitutionOps.GetInstitution).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Add institution")
	superSubrouter.HandleFunc("/institutions", app.InstitutionOps.AddInstitution).Methods(http.MethodPost)

	logger.Infof("\t* Portal API:  Update institution")
	SPSubrouter.HandleFunc("/institutions/{institutionId}", app.InstitutionOps.UpdateInstitution).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Remove institution")
	SPSubrouter.HandleFunc("/institutions/{institutionId}", app.InstitutionOps.RemoveInstitution).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get blocklist requests")
	superSubrouter.HandleFunc("/blocklist_requests", app.BlocklistOps.GetBlocklistRequests).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Add blocklist request")
	superSubrouter.HandleFunc("/blocklist_requests", app.BlocklistOps.AddBlocklistRequest).Methods(http.MethodPost)

	logger.Infof("\t* Portal API:  Remove blocklist request")
	superSubrouter.HandleFunc("/blocklist_requests/{approvalId}", app.BlocklistOps.RemoveBlocklistRequest).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get super approval")
	superSubrouter.HandleFunc("/super_approvals/{approvalId}", app.SuperOps.GetSuperApproval).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Update super approval")
	superSubrouter.HandleFunc("/super_approvals/{approvalId}", app.SuperOps.UpdateSuperApproval).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Get participant approval")
	participantSubrouter.HandleFunc("/participant_approvals/{approvalId}", app.ParticipantOps.GetParticipantApproval).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Update participant approval")
	participantSubrouter.HandleFunc("/participant_approvals/{approvalId}", app.ParticipantOps.UpdateParticipantApproval).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Get user")
	SPSubrouter.HandleFunc("/users/{userId}", app.UserOps.GetUser).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Get super users")
	superSubrouter.HandleFunc("/super/users", app.UserOps.GetSuperUsers).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Update super user")
	superSubrouter.HandleFunc("/super/users", app.UserOps.UpdateSuperUser).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Remove super user")
	superSubrouter.HandleFunc("/super/users/{userId}", app.UserOps.RemoveSuperUser).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get participant users")
	SPSubrouter.HandleFunc("/{institutionId}/users", app.UserOps.GetParticipantUsers).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Update participant user")
	SPSubrouter.HandleFunc("/{institutionId}/users/{userId}", app.UserOps.UpdateParticipantUser).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Remove participant user")
	SPSubrouter.HandleFunc("/{institutionId}/users/{userId}", app.UserOps.RemoveParticipantUser).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get account requests")
	SPSubrouter.HandleFunc("/{participantId}/account_requests", app.AccountOps.GetAccountRequests).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Add account request")
	SPSubrouter.HandleFunc("/account_requests", app.AccountOps.AddAccountRequest).Methods(http.MethodPost)

	logger.Infof("\t* Portal API:  Update account request")
	SPSubrouter.HandleFunc("/account_requests/{approvalId}", app.AccountOps.UpdateAccountRequest).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Remove account request")
	SPSubrouter.HandleFunc("/account_requests/{approvalId}", app.AccountOps.RemoveAccountRequest).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get whitelist requests")
	participantSubrouter.HandleFunc("/{participantId}/whitelist_requests", app.WhitelistOps.GetWhitelistRequests).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Add whitelist request")
	participantSubrouter.HandleFunc("/whitelist_requests", app.WhitelistOps.AddWhitelistRequest).Methods(http.MethodPost)

	logger.Infof("\t* Portal API:  Update whitelist request")
	participantSubrouter.HandleFunc("/whitelist_requests/{whitelistedId}", app.WhitelistOps.UpdateWhitelistRequest).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Remove whitelist request")
	participantSubrouter.HandleFunc("/whitelist_requests/{whitelistedId}", app.WhitelistOps.RemoveWhitelistRequest).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get asset requests")
	participantSubrouter.HandleFunc("/{participantId}/asset_requests", app.AssetOps.GetAssetRequests).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Add asset request")
	participantSubrouter.HandleFunc("/asset_requests", app.AssetOps.AddAssetRequest).Methods(http.MethodPost)

	logger.Infof("\t* Portal API:  Update asset request")
	participantSubrouter.HandleFunc("/asset_requests", app.AssetOps.UpdateAssetRequest).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Remove asset request")
	participantSubrouter.HandleFunc("/asset_requests/{approvalId}", app.AssetOps.RemoveAssetRequest).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get trust requests")
	participantSubrouter.HandleFunc("/{participantId}/trust_requests/{requestField}", app.TrustOps.GetTrustRequests).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Add trust request")
	participantSubrouter.HandleFunc("/trust_requests", app.TrustOps.AddTrustRequest).Methods(http.MethodPost)

	logger.Infof("\t* Portal API:  Update trust request")
	participantSubrouter.HandleFunc("/trust_requests/{trustRequestId}", app.TrustOps.UpdateTrustRequest).Methods(http.MethodPatch)

	logger.Infof("\t* Portal API:  Remove trust request")
	participantSubrouter.HandleFunc("/trust_requests/{trustRequestId}", app.TrustOps.RemoveTrustRequest).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get killswitch request")
	participantSubrouter.HandleFunc("/{participantId}/killswitch_requests/{accountAddress}", app.KillswitchOps.GetKillswitchRequest).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Add killswitch request")
	participantSubrouter.HandleFunc("/killswitch_requests", app.KillswitchOps.AddKillswitchRequest).Methods(http.MethodPost)

	logger.Infof("\t* Portal API:  Update killswitch request")
	participantSubrouter.HandleFunc("/{participantId}/killswitch_requests/{accountAddress}", app.KillswitchOps.UpdateKillswitchRequest).Methods(http.MethodPut)

	logger.Infof("\t* Portal API:  Remove killswitch request")
	participantSubrouter.HandleFunc("/killswitch_requests/{approvalId}", app.KillswitchOps.RemoveKillswitchRequest).Methods(http.MethodDelete)

	logger.Infof("\t* Portal API:  Get exchange quotes")
	participantSubrouter.HandleFunc("/{participantId}/exchange/transactions", app.ExchangeOps.GetTransactions).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Get tranfer payments")
	participantSubrouter.HandleFunc("/{participantId}/transfer/transactions", app.TransferOps.GetTransactions).Methods(http.MethodGet)

	logger.Infof("\t* Portal API:  Get JWT info")
	participantSubrouter.HandleFunc("/{institutionId}/jwt_info", app.JwtOps.GetJwtInfo).Methods(http.MethodGet)

	portalSubrouter.NotFoundHandler = http.HandlerFunc(response.NotFound)

	// Add middleware for authentication
	portalSubrouter.Use(app.Authentication.AuthenticateUser)
	// Add middleware for authorization of either super or participant
	SPSubrouter.Use(app.Authorization.AuthorizeSuperOrParticipantUser, app.Authorization.CheckTOTPPasscode)
	superSubrouter.Use(app.Authorization.AuthorizeSuperUser, app.Authorization.CheckTOTPPasscode)
	participantSubrouter.Use(app.Authorization.AuthorizeParticipantUser, app.Authorization.CheckTOTPPasscode)
}

func main() {

	// Set up environment variables
	services.VariableCheck()
	services.InitEnv()

	app := app{}
	var err error

	// Retrieve environment variables
	app.ServiceVersion = os.Getenv(global_environment.ENV_KEY_SERVICE_VERSION)
	portalDbName := os.Getenv(environment.ENV_KEY_PORTAL_DB_NAME)
	authDbName := os.Getenv(environment.ENV_KEY_AUTH_DB_NAME)
	txDbName := os.Getenv(environment.ENV_KEY_TX_DB_NAME)
	institutionCollName := os.Getenv(environment.ENV_KEY_INSTITUTION_DB_TABLE)
	permissionCollName := os.Getenv(environment.ENV_KEY_PERMISSION_DB_TABLE)
	blocklistCollName := os.Getenv(environment.ENV_KEY_BLOCKLIST_REQ_DB_TABLE)
	whitelistCollName := os.Getenv(environment.ENV_KEY_WHITELIST_REQ_DB_TABLE)
	accountCollName := os.Getenv(environment.ENV_KEY_ACCOUNT_REQ_DB_TABLE)
	trustCollName := os.Getenv(environment.ENV_KEY_TRUST_REQ_DB_TABLE)
	assetCollName := os.Getenv(environment.ENV_KEY_ASSET_REQ_DB_TABLE)
	killswitchCollName := os.Getenv(environment.ENV_KEY_KILLSWITCH_REQ_DB_TABLE)
	superCollName := os.Getenv(environment.ENV_KEY_SUPER_APPROVAL_DB_TABLE)
	participantCollName := os.Getenv(environment.ENV_KEY_PARTICIPANT_APPROVAL_DB_TABLE)
	userCollName := os.Getenv(environment.ENV_KEY_USER_DB_TABLE)
	jwtCollName := os.Getenv(environment.ENV_KEY_JWT_INFO_DB_TABLE)
	jwtSecureCollName := os.Getenv(environment.ENV_KEY_JWT_SECURE_DB_TABLE)
	transferCollName := os.Getenv(environment.ENV_KEY_PAYMENT_DB_TABLE)
	exchangeCollName := os.Getenv(environment.ENV_KEY_QUOTE_DB_TABLE)
	dbTimeoutInt, err := strconv.ParseInt(os.Getenv(environment.ENV_KEY_DB_TIMEOUT), 10, 64)
	dbTimeout := time.Duration(dbTimeoutInt)
	utility.ExitOnErr(logger, err, "Error converting char to int")

	// Establish connection with database
	dbClient, err := database.InitializeIbmCloudConnection()
	utility.ExitOnErr(logger, err, "Unable to connect to DB")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), dbTimeout*time.Second)
		defer cancel()
		err = dbClient.Session.Disconnect(ctx)
		utility.ExitOnErr(logger, err, "Unable to disconnect to DB")
	}()

	// Instantiate multiple operations
	app.AccountOps = portalops.CreateAccountOps(dbClient, portalDbName, accountCollName, dbTimeout)
	app.BlocklistOps = portalops.CreateBlocklistOps(dbClient, portalDbName, blocklistCollName, dbTimeout)
	app.WhitelistOps = portalops.CreateWhitelistOps(dbClient, portalDbName, whitelistCollName, dbTimeout)
	app.InstitutionOps = portalops.CreateInstitutionOps(dbClient, portalDbName, institutionCollName, permissionCollName, dbTimeout)
	app.TrustOps = portalops.CreateTrustOps(dbClient, portalDbName, trustCollName, dbTimeout)
	app.AssetOps = portalops.CreateAssetOps(dbClient, portalDbName, assetCollName, dbTimeout)
	app.KillswitchOps = portalops.CreateKillswitchOps(dbClient, portalDbName, killswitchCollName, dbTimeout)
	app.SuperOps = portalops.CreateSuperOps(dbClient, portalDbName, superCollName, dbTimeout)
	app.ParticipantOps = portalops.CreateParticipantOps(dbClient, portalDbName, participantCollName, dbTimeout)
	app.UserOps = portalops.CreateUserOps(dbClient, portalDbName, userCollName, institutionCollName, permissionCollName, dbTimeout)
	app.JwtOps = portalops.CreateJwtOps(dbClient, authDbName, jwtCollName, dbTimeout)
	app.TransferOps = portalops.CreateTxOps(dbClient, txDbName, transferCollName, dbTimeout)
	app.ExchangeOps = portalops.CreateTxOps(dbClient, txDbName, exchangeCollName, dbTimeout)
	app.Authentication = middleware.CreateAuthentication(dbClient, authDbName, jwtSecureCollName, dbTimeout)
	app.Authorization = middleware.CreateAuthorization(dbClient, portalDbName, userCollName, permissionCollName, dbTimeout)

	// Load error code definition file
	errorCodes := os.Getenv(global_environment.ENV_KEY_SERVICE_ERROR_CODES_FILE)
	err = message.LoadErrorConfig(errorCodes)
	utility.ExitOnErr(logger, err, "Unable to set up error message config")

	// Set up log file
	logFilePath := os.Getenv(global_environment.ENV_KEY_SERVICE_LOG_FILE)
	logFile, err := logconfig.SetupLogging(logFilePath, logger)
	utility.ExitOnErr(logger, err, "Unable to set up logging")
	defer logFile.Close()

	servicePort := os.Getenv(global_environment.ENV_KEY_SERVICE_PORT)

	writeTimeout, _ := strconv.ParseInt(os.Getenv(global_environment.ENV_KEY_WRITE_TIMEOUT), 10, 64)
	readTimeout, _ := strconv.ParseInt(os.Getenv(global_environment.ENV_KEY_READ_TIMEOUT), 10, 64)
	idleTimeout, _ := strconv.ParseInt(os.Getenv(global_environment.ENV_KEY_IDLE_TIMEOUT), 10, 64)

	if writeTimeout == 0 || readTimeout == 0 || idleTimeout == 0 {
		panic("Service timeout should not be zero, please check if the environment variables WRITE_TIMEOUT, READ_TIMEOUT, IDLE_TIMEOUT are being set correctly")
	}

	app.initRoutes()

	var httpHandler http.Handler = app.Router

	if os.Getenv(global_environment.ENV_KEY_ORIGIN_ALLOWED) == "true" {
		headersOk := handlers.AllowedHeaders([]string{"Access-Control-Allow-Headers", "Origin", "Content-Type", "X-Auth-Token", "Authorization", "X-Fid", "X-Iid", "X-Pid", "X-Verify-Code"})
		originsOk := handlers.AllowedOrigins([]string{"*"})
		methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "PUT", "OPTIONS", "DELETE"})
		logger.Infof("Setting up CORS")
		httpHandler = handlers.CORS(
			headersOk,
			originsOk,
			methodsOk,
		)(app.Router)
	}

	// Instantiate server
	server := &http.Server{
		Addr: ":" + servicePort,
		// Set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * time.Duration(writeTimeout),
		ReadTimeout:  time.Second * time.Duration(readTimeout),
		IdleTimeout:  time.Second * time.Duration(idleTimeout),
		Handler:      httpHandler,
	}
	logger.Infof("Listening on: %s", servicePort)
	err = server.ListenAndServe()
	logger.Error(err.Error())

	os.Exit(0)

}
