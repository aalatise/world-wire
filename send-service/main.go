package main

import (
	"context"
	"flag"
	"github.com/gorilla/handlers"
	middlewares "github.com/IBM/world-wire/auth-service-go/handler"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/lestrrat-go/libxml2/xsd"
	"github.com/IBM/world-wire/utility/response"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/urfave/negroni"
	"github.com/IBM/world-wire/send-service/handler"

	u "github.com/IBM/world-wire/utility"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"github.com/IBM/world-wire/utility/global-environment/services"
	"github.com/IBM/world-wire/utility/logconfig"
	"github.com/IBM/world-wire/utility/message"
	middleware_checks "github.com/IBM/world-wire/utility/middleware"
	"github.com/IBM/world-wire/utility/payment/constant"
	message_handler "github.com/IBM/world-wire/utility/payment/message-handler"
	_ "net/http/pprof"
)

type App struct {
	Router      *mux.Router
	sendHandler message_handler.PaymentOperations
	mwHandler   *middleware_checks.MiddlewareHandler
	HTTPHandler func(http.Handler) http.Handler
}

var LOGGER = logging.MustGetLogger("send-service")

func (a *App) initializeHandlers() (message_handler.PaymentOperations, error) {
	a.HTTPHandler = nil
	if os.Getenv(global_environment.ENV_KEY_ORIGIN_ALLOWED) == "true" {
		headersOk := handlers.AllowedHeaders([]string{"Access-Control-Allow-Headers", "Origin", "Content-Type", "X-Auth-Token", "Authorization", "X-Fid", "X-Iid", "X-Pid", "X-Permission", "X-Request"})
		originsOk := handlers.AllowedOrigins([]string{"*"})
		methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
		LOGGER.Infof("Setting up CORS")
		a.HTTPHandler = handlers.CORS(
			headersOk, originsOk, methodsOk)
	}

	LOGGER.Infof("Initializing Kafka producer...")
	a.mwHandler = middleware_checks.CreateMiddlewareHandler()
	sendHandler, err := message_handler.InitiatePaymentOperations()

	if err != nil {
		LOGGER.Error(err.Error())
		return message_handler.PaymentOperations{}, err
	}
	LOGGER.Infof("Initializing Kafka consumer...")
	initConsumerErr := sendHandler.KafkaActor.InitPaymentConsumer("G1", handler.KafkaRouter)
	if initConsumerErr != nil {
		LOGGER.Errorf("Initialize Kafka consumer failed: %s", initConsumerErr.Error())
		return message_handler.PaymentOperations{}, initConsumerErr
	}

	return sendHandler, nil
}

func (a *App) initializeRoutes() {
	serviceVersion := os.Getenv(global_environment.ENV_KEY_SERVICE_VERSION)

	a.Router = mux.NewRouter()
	/*
		Service check Endpoints
	*/
	LOGGER.Infof("\t* External API: Service Check")
	a.Router.Handle("/"+serviceVersion+"/client/service_check", negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
		negroni.WrapFunc(ServiceCheck),
	)).Methods("GET")

	LOGGER.Infof("\t* External API: Send payment request")
	a.Router.Handle("/"+serviceVersion+"/client/transactions/send", negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
		negroni.WrapFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.Router(w, r, a.sendHandler, constant.REQUEST)
		}),
	)).Methods("POST")

	LOGGER.Infof("\t* External API: Send response")
	a.Router.Handle("/"+serviceVersion+"/client/transactions/reply", negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
		negroni.WrapFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.Router(w, r, a.sendHandler, constant.RESPONSE)
		}),
	)).Methods("POST")

	LOGGER.Infof("\t* External API: DA redemption")
	a.Router.Handle("/"+serviceVersion+"/client/transactions/redeem", negroni.New(
		negroni.HandlerFunc(middlewares.ParticipantAuthorization),
		negroni.HandlerFunc(a.mwHandler.ParticipantStatusCheck),
		negroni.WrapFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.Router(w, r, a.sendHandler, constant.REDEEM)
		}),
	)).Methods("POST")
}

func ReleaseXSD(xsds []*xsd.Schema) {
	for _, i := range xsds {
		if i != nil {
			i.Free()
		}
	}
}

func ServiceCheck(w http.ResponseWriter, req *http.Request) {
	LOGGER.Infof("Performing service check")
	response.Respond(w, http.StatusOK, []byte(`{"status":"Alive"}`))
	return
}

func main() {
	services.VariableCheck()
	services.InitEnv()
	serviceLogs := os.Getenv(global_environment.ENV_KEY_SERVICE_LOG_FILE)
	f, err := logconfig.SetupLogging(serviceLogs, LOGGER)
	if err != nil {
		u.ExitOnErr(LOGGER, err, "Unable to set up logging")
	}
	defer f.Close()

	APP := App{}

	APP.sendHandler, err = APP.initializeHandlers()
	if err != nil {
		panic(err)
	}

	APP.initializeRoutes()

	APP.Router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	defer ReleaseXSD(APP.sendHandler.XsdSchemas)

	servicePort := os.Getenv(global_environment.ENV_KEY_SERVICE_PORT)
	errorCodes := os.Getenv(global_environment.ENV_KEY_SERVICE_ERROR_CODES_FILE)

	err = message.LoadErrorConfig(errorCodes)
	u.ExitOnErr(LOGGER, err, "Unable to set up error message config")

	var handler http.Handler = APP.Router
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

	LOGGER.Infof("Send service listen and serve on port: %v\n", servicePort)
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
