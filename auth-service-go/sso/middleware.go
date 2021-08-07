package sso

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/op/go-logging"
	"github.com/IBM/world-wire/auth-service-go/environment"
	"net/http"
	"os"
	"strings"
)

var SessionStore = sessions.NewCookieStore([]byte("very secret"))
var LOGGER = logging.MustGetLogger("sso")

func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv(environment.ENV_KEY_PORTAL_DOMAIN))
		w.Header().Set("Access-Control-Allow-Headers", "*")
		next.ServeHTTP(w, r)
	})
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LOGGER.Debugf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event // of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a // panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500 // Internal Server response.
				ServerError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LOGGER.Debug("----- Middleware: Authenticating user")
		name := "user-sessions"
		//session, err := firestore.SessionStore.Get(r, name)

		session, err := SessionStore.Get(r, name)
		if err != nil {
			// Could not get the session. Log an error and continue, saving a new
			// session.
			LOGGER.Debugf("store.Get: %v", err)
		}

		LOGGER.Debugf("UserId: %v", session.Values["userId"])

		LOGGER.Debugf("request uri: %s", r.URL.RequestURI())
		if r.URL.RequestURI() == "/sso/portal-login-totp" || strings.Contains(r.URL.RequestURI(), "/totp/") {
			LOGGER.Debugf("handle /sso/portal-login-totp or /totp/ endpoints")
			next.ServeHTTP(w, r)
			return
		} else if session.Values["userId"] == nil {
			http.Redirect(w, r, os.Getenv("LOGIN_URL"), http.StatusMovedPermanently)
		} else {
			LOGGER.Debugf("next server")
			r.Header.Set("email", session.Values["userId"].(string))
			next.ServeHTTP(w, r)
			return
		}
	})
}
