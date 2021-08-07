package handler

import (
	"net/http"

	middlewares "github.com/IBM/world-wire/auth-service-go/handler"
)

type MiddleWare struct {
}

func (mw *MiddleWare) VerifyTokenAndEndpoints(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	LOGGER.Info("checking jwt validaty")
	middlewares.ParticipantAuthorization(w, r, next)
}
