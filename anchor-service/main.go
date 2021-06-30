package main

import (
	"context"
	"flag"
	middlewares "github.ibm.com/gftn/world-wire-services/auth-service-go/handler"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	httpHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	b "github.com/stellar/go/build"
	"github.com/urfave/negroni"
	"github.ibm.com/gftn/world-wire-services/anchor-service/handlers"
	"github.ibm.com/gftn/world-wire-services/anchor-service/kafka"
	"github.ibm.com/gftn/world-wire-services/api-service/participants"
	"github.ibm.com/gftn/world-wire-services/utility"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/global-environment/services"
	"github.ibm.com/gftn/world-wire-services/utility/logconfig"
	"github.ibm.com/gftn/world-wire-services/utility/message"
	middleware_checks "github.ibm.com/gftn/world-wire-services/utility/middleware"
	message_handler "github.ibm.com/gftn/world-wire-services/utility/payment/message-handler"
	"github.ibm.com/gftn/world-wire-services/utility/status"
)

type App struct {
	Router                     *mux.Router
	HTTPHandler                func(http.Handler) http.Handler
	mwHandler                  *middleware_checks.MiddlewareHandler
	serviceCheck               status.ServiceCheck
	discoverParticipantHandler handlers.DiscoverHandler
	trustHandler               handlers.TrustHandler
	fundHandler                handlers.FundHandler
	onBoardingHandler          handlers.OnBoardingHandler
	participantOps             participants.ParticipantOperations
	sendHandler                *message_handler.PaymentOperations
}

var APP App

var serviceVersion = ""

func (a *App) Initialize() *message_handler.PaymentOperations {
	services.VariableCheck()
	services.InitEnv()

	a.HTTPHandler = nil
	if os.Getenv(global_environment.ENV_KEY_ORIGIN_ALLOWED) == "true" {
		headersOk := httpHandlers.AllowedHeaders([]string{"Access-Control-Allow-Headers", "Origin", "Content-Type", "X-Auth-Token", "Authorization", "X-Fid", "X-Iid", "X-Pid", "X-Permission", "X-Request"})
		originsOk := httpHandlers.AllowedOrigins([]string{"*"})
		methodsOk := httpHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
		LOGGER.Infof("Setting up CORS")
		a.HTTPHandler = httpHandlers.CORS(
			headersOk, originsOk, methodsOk)
	}
	serviceVersion = os.Getenv(global_environment.ENV_KEY_SERVICE_VERSION)

	//Set Defaults for stellar network
	b.DefaultNetwork.Passphrase = os.Getenv(global_environment.ENV_KEY_STELLAR_NETWORK)

	LOGGER.Infof("Initializing Kafka producer")
	sendHandler, err := message_handler.InitiatePaymentOperations()
	if err != nil {
		LOGGER.Error(err.Error())
		return nil
	}

	LOGGER.Infof("Initializing Kafka consumer")
	initConsumerErr := sendHandler.KafkaActor.InitPaymentConsumer("G1", kafka.KafkaRouter)
	if initConsumerErr != nil {
		LOGGER.Errorf("Initialize Kafka consumer failed: %s", initConsumerErr.Error())
		return nil
	}

	errorCodes := os.Getenv(global_environment.ENV_KEY_SERVICE_ERROR_CODES_FILE)
	err = message.LoadErrorConfig(errorCodes)
	utility.ExitOnErr(LOGGER, err, "Unable to set up error message config")

	// Create middleware handler
	a.mwHandler = middleware_checks.CreateMiddlewareHandler()

	LOGGER.Infof("Setting up service status check")
	a.serviceCheck, err = status.CreateServiceCheck()
	utility.ExitOnErr(LOGGER, err, "Unable to set up Service Check API")

	LOGGER.Infof("Setting up Anchor Service API")
	a.discoverParticipantHandler, err = handlers.CreateDiscoverHandler()
	utility.ExitOnErr(LOGGER, err, "Unable to set up DiscoverHandler API")
	a.trustHandler, err = handlers.CreateTrustHandler()
	a.fundHandler, err = handlers.CreateFundHandler()
	utility.ExitOnErr(LOGGER, err, "Unable to set up DiscoverHandler API")

	a.onBoardingHandler, err = handlers.CreateOnBoardingHandler()
	utility.ExitOnErr(LOGGER, err, "Unable to set up onBoardingHandler API")

	// Participant operations
	LOGGER.Infof("Setting up Participant Ops API")
	a.participantOps, err = participants.CreateParticipantOperations()
	utility.ExitOnErr(LOGGER, err, "Unable to set up Participant Registry API")

	return &sendHandler

}

var LOGGER = logging.MustGetLogger("anchor-service")

func (a *App) initializeRoutes() {

	a.Router = mux.NewRouter()

	url := "/" + serviceVersion + "/anchor/service_check"
	LOGGER.Infof("\t* Internal API:  Service Check")
	a.Router.Handle(url, negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.WrapFunc(a.serviceCheck.ServiceCheck),
	)).Methods("GET")

	url = "/" + serviceVersion + "/anchor/address"
	LOGGER.Infof("Anchor Service Discover URL: %v", url)
	a.Router.Handle(url, negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.discoverParticipantHandler.DiscoverParticipant),
	)).Methods("GET")

	url = "/" + serviceVersion + "/anchor/assets/issued/{anchor_id}"
	LOGGER.Infof("Anchor Service Issued asset URL: %v", url)
	a.Router.Handle(url, negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.trustHandler.GetIssuedAssets),
	)).Methods(http.MethodGet)

	LOGGER.Infof("\t* External API:  execute Allow Trust Operation")
	a.Router.Handle("/"+serviceVersion+"/anchor/trust/{anchor_id}",
		negroni.New(
			negroni.HandlerFunc(middlewares.ParticipantAuthorization),
			negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
			negroni.HandlerFunc(a.trustHandler.AllowTrust),
		)).Methods("POST")

	// handler for anchor fund request
	LOGGER.Infof("\t* External API:  execute Fund Request Operation")
	a.Router.Handle("/"+serviceVersion+"/anchor/fundings/instruction",
		negroni.New(
			negroni.HandlerFunc(middlewares.ParticipantAuthorization),
			negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
			negroni.HandlerFunc(a.fundHandler.FundRequest),
		)).Methods("POST")

	LOGGER.Infof("\t* External API:  execute signed Fund Request Operation")
	a.Router.Handle("/"+serviceVersion+"/anchor/fundings/send",
		negroni.New(
			negroni.HandlerFunc(middlewares.ParticipantAuthorization),
			negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
			negroni.HandlerFunc(a.fundHandler.SignedFundRequest),
		)).Methods("POST")

	LOGGER.Infof("\t* External API:  get participants on WW using query")
	a.Router.Handle("/"+serviceVersion+"/anchor"+"/participants", negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
		negroni.WrapFunc(a.participantOps.GetParticipantByQuery),
	)).Methods(http.MethodGet)

	LOGGER.Infof("\t* External API:  Response of client asset redemption request")
	a.Router.Handle("/"+serviceVersion+"/anchor/assets/redeem", negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
		negroni.WrapFunc(func(w http.ResponseWriter, r *http.Request) {
			kafka.Router(w, r, *a.sendHandler)
		}),
	)).Methods(http.MethodPost)

	LOGGER.Infof("\t* External API:  get participants on WW using query")
	a.Router.Handle("/"+serviceVersion+"/anchor"+"/participants/{participant_id}", negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
		negroni.WrapFunc(a.participantOps.GetParticipantByDomain),
	)).Methods(http.MethodGet)

	LOGGER.Infof("\t* Admin API: Register an anchor on WW")

	a.Router.Handle("/"+serviceVersion+"/admin/anchor/{anchor_domain}/register", negroni.New(
		negroni.HandlerFunc(middlewares.SuperAuthorization),
		negroni.WrapFunc(a.onBoardingHandler.RegisterAnchor),
	)).Methods("POST")

	LOGGER.Infof("\t* Admin API: get issued assets by an anchor on WW")
	url = "/" + serviceVersion + "/admin/anchor/assets/issued/{anchor_id}"
	LOGGER.Infof("Anchor Service Issued asset URL: %v", url)
	a.Router.Handle(url, negroni.New(
		negroni.HandlerFunc(middlewares.SuperAuthorization),
		negroni.HandlerFunc(a.trustHandler.GetIssuedAssets),
	)).Methods(http.MethodGet)

	LOGGER.Infof("\t* Admin API: Onboard Anchor Asset on WW")
	a.Router.Handle("/"+serviceVersion+"/admin/anchor/{anchor_domain}/onboard/assets", negroni.New(
		negroni.HandlerFunc(middlewares.SuperAuthorization),
		negroni.WrapFunc(a.onBoardingHandler.OnBoardAsset),
	)).Methods("POST")

}

func main() {
	APP = App{}
	APP.sendHandler = APP.Initialize()
	serviceLogs := os.Getenv(global_environment.ENV_KEY_SERVICE_LOG_FILE)
	f, err := logconfig.SetupLogging(serviceLogs, LOGGER)
	if err != nil {
		utility.ExitOnErr(LOGGER, err, "Unable to set up logging")
	}
	defer f.Close()

	APP.initializeRoutes()

	servicePort := os.Getenv(global_environment.ENV_KEY_SERVICE_PORT)

	var handler http.Handler
	handler = APP.Router
	//if CORS is set
	if APP.HTTPHandler != nil {
		handler = APP.HTTPHandler(APP.Router)
	}

	writeTimeout, _ := strconv.ParseInt(os.Getenv(global_environment.ENV_KEY_WRITE_TIMEOUT), 10, 64)
	readTimeout, _ := strconv.ParseInt(os.Getenv(global_environment.ENV_KEY_READ_TIMEOUT), 10, 64)
	idleTimeout, _ := strconv.ParseInt(os.Getenv(global_environment.ENV_KEY_IDLE_TIMEOUT), 10, 64)

	if writeTimeout == 0 || readTimeout == 0 || idleTimeout == 0 {
		panic("Service timeout should not be zero, please check if the environment variables WRITE_TIMEOUT, READ_TIMEOUT, IDLE_TIMEOUT are being set correctly")
	}

	srv := &http.Server{
		Addr: ":" + servicePort,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * time.Duration(writeTimeout),
		ReadTimeout:  time.Second * time.Duration(readTimeout),
		IdleTimeout:  time.Second * time.Duration(idleTimeout),
		//TLSConfig:    &cfg,
		Handler: handler, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		LOGGER.Error(srv.ListenAndServe().Error())
	}()

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s")
	flag.Parse()
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)

	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	LOGGER.Errorf("shutting down")
	os.Exit(0)

}
