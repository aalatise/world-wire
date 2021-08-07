package sso

import (
	"bytes"
	"fmt"
	"github.com/IBM/world-wire/auth-service-go/sso/pkg/template"
	"net/http"
	"runtime/debug"
)

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	LOGGER.Debug(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func Render(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := template.TemplateCache[name]
	if !ok {
		ServerError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer)
	err := ts.Execute(buf, data)

	if err != nil {
		ServerError(w, err)
		return
	}

	buf.WriteTo(w)
}