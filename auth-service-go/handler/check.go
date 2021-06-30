package handler

import "net/http"

func (op *AuthOperations) MiddlewareCheck(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("Hello, World Wire")
	LOGGER.Debugf("Congratulations, you passed the middleware check")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Congratulations, you passed the middleware check"))
}
