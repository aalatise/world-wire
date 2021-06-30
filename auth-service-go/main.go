package main

import (
	"context"
	"flag"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/op/go-logging"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/handler"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/sso"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"net/http"
	"os"
	"os/signal"
	"time"

	u "github.ibm.com/gftn/world-wire-services/utility"
	"github.ibm.com/gftn/world-wire-services/utility/logconfig"
	"github.ibm.com/gftn/world-wire-services/utility/response"
)

type App struct {
	authOp      handler.AuthOperations
	Router      *mux.Router
	HTTPHandler func(http.Handler) http.Handler
}

var LOGGER = logging.MustGetLogger("auth-service")

func (a *App) initializeHandlers() error {
	op, err := handler.CreateAuthServiceOperations()
	if err != nil {
		LOGGER.Errorf(err.Error())
		return err
	}

	a.authOp = op

	headersOk := handlers.AllowedHeaders([]string{"Access-Control-Allow-Headers", "Access-Control-Allow-Origin", "Origin", "Content-Type", "X-Auth-Token", "Authorization", "X-Fid", "X-Iid", "X-Verify-Token", "x-verify-code"})
	originsOk := handlers.AllowedOrigins([]string{os.Getenv("ALLOW_ORIGIN")})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
	maxAgeOk := handlers.MaxAge(86400)
	credOk := handlers.AllowCredentials()
	LOGGER.Infof("Setting up CORS ...")
	a.HTTPHandler = handlers.CORS(headersOk, originsOk, methodsOk, credOk, maxAgeOk)

	return nil
}

func (a *App) initializeRoutes() {
	a.Router = mux.NewRouter()
	standardMiddleware := alice.New(sso.RecoverPanic, sso.LogRequest, sso.SecureHeaders)
	secureMiddleware := alice.New(sso.Authenticate, sso.SecureHeaders)
	totpMiddleware := alice.New(sso.Authenticate, sso.SecureHeaders)//, a.authOp.CheckAccountNameMiddleWare, sso.SecureHeaders)
	portalMiddleware := alice.New(sso.Authenticate, a.authOp.CheckTOTPMiddleWareIBMIdUser, sso.SecureHeaders)
	jwtMiddleware := alice.New(a.authOp.AuthenticateUser, a.authOp.CheckPermissions, a.authOp.CheckTOTPMiddleWarePortalUser, sso.SecureHeaders)
	//jwtRefreshMiddleware := alice.New(sso.SecureHeaders)

	a.Router.Handle("/check", standardMiddleware.ThenFunc(a.authOp.ServiceCheck)).Methods(http.MethodGet)

	// Routes for SSO login, logout
	a.Router.Handle("/sso/login", standardMiddleware.ThenFunc(sso.HandleIBMIdLogin)).Methods(http.MethodGet)
	a.Router.Handle("/sso/callback", standardMiddleware.ThenFunc(sso.HandleIBMIdLoginCallback)).Methods(http.MethodGet)
	//a.Router.Handle("/sso/logout", standardMiddleware.ThenFunc(sso.Logout)).Methods(http.MethodGet)
	a.Router.Handle("/sso/token", secureMiddleware.ThenFunc(a.authOp.HandleSSOToken)).Methods(http.MethodGet)
	a.Router.Handle("/sso/portal-login-totp", portalMiddleware.ThenFunc(a.authOp.HandlePortalLoginTOTP)).Methods(http.MethodPost)

	//a.Router.Handle("/sso/failure", secureMiddleware.ThenFunc(sso.Token)).Methods(http.MethodGet)

	// Routes for TOTP create and confirm
	a.Router.Handle("/totp/{accountName}", totpMiddleware.ThenFunc(a.authOp.HandleTOTPCreate)).Methods(http.MethodGet)
	a.Router.Handle("/totp/{accountName}/confirm", totpMiddleware.ThenFunc(a.authOp.HandleTOTPConfirm)).Methods(http.MethodPost)

	// Route for generate ID Token
	a.Router.Handle("/idtoken/generate", standardMiddleware.ThenFunc(a.authOp.HandleGenerateIDToken)).Methods(http.MethodGet)

	// Routes for JWT
	a.Router.Handle("/jwt/generate", jwtMiddleware.ThenFunc(a.authOp.HandleJWTGenerate)).Methods(http.MethodPost)
	a.Router.Handle("/jwt/request", jwtMiddleware.ThenFunc(a.authOp.HandleJWTRequest)).Methods(http.MethodPost)
	a.Router.Handle("/jwt/verify", jwtMiddleware.ThenFunc(a.authOp.HandleJWTVerify)).Methods(http.MethodPost)
	a.Router.Handle("/jwt/refresh", standardMiddleware.ThenFunc(a.authOp.HandleJWTRefresh)).Methods(http.MethodPost)
	a.Router.Handle("/jwt/revoke", jwtMiddleware.ThenFunc(a.authOp.HandleJWTRevoke)).Methods(http.MethodPost)
	a.Router.Handle("/jwt/reject", jwtMiddleware.ThenFunc(a.authOp.HandleJWTRevoke)).Methods(http.MethodPost)
	a.Router.Handle("/jwt/approve", jwtMiddleware.ThenFunc(a.authOp.HandleJWTApprove)).Methods(http.MethodPost)

	// Routes for handling permission update
	a.Router.Handle("/permissions/participant", jwtMiddleware.ThenFunc(a.authOp.HandlePermissionParticipantUpdate)).Methods(http.MethodPost)
	a.Router.Handle("/permissions/super", jwtMiddleware.ThenFunc(a.authOp.HandlePermissionSuperUpdate)).Methods(http.MethodPost)

	a.Router.NotFoundHandler = http.HandlerFunc(sso.NotFound)
}

func ServiceCheck(w http.ResponseWriter, req *http.Request) {
	LOGGER.Infof("Performing service check")
	response.Respond(w, http.StatusOK, []byte(`{"status":"Alive"}`))
	return
}

func main() {
	serviceLogs := os.Getenv(global_environment.ENV_KEY_SERVICE_LOG_FILE)
	f, err := logconfig.SetupLogging(serviceLogs, LOGGER)
	if err != nil {
		u.ExitOnErr(LOGGER, err, "Unable to set up logging")
	}
	defer f.Close()

	APP := App{}

	err = APP.initializeHandlers()
	if err != nil {
		panic(err)
	}

	APP.initializeRoutes()

	servicePort := os.Getenv(global_environment.ENV_KEY_SERVICE_PORT)

	var httpHandler http.Handler
	httpHandler = APP.HTTPHandler(APP.Router)

	srv := &http.Server{
		Addr: ":" + servicePort,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      httpHandler, // Pass our instance of gorilla/mux in.
	}

	LOGGER.Infof("Auth service listen and serve on port: %v\n", servicePort)
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
