package portalops

import (
	"net/http"

	"github.ibm.com/gftn/world-wire-services/utility/response"
)

func StatusCheck(w http.ResponseWriter, r *http.Request) {
	response.NotifySuccess(w, r, "OK")
}
