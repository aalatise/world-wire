package portalops

import (
	"net/http"

	"github.com/IBM/world-wire/utility/response"
)

func StatusCheck(w http.ResponseWriter, r *http.Request) {
	response.NotifySuccess(w, r, "OK")
}
