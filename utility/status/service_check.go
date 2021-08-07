package status

import (
	"net/http"

	"github.com/IBM/world-wire/utility/response"
)

type ServiceCheck struct {
}

func CreateServiceCheck() (ServiceCheck, error) {
	sc := ServiceCheck{}
	return sc, nil
}

func (ServiceCheck) ServiceCheck(w http.ResponseWriter, req *http.Request) {
	LOGGER.Infof("Performing service check")
	//Service check sends message okay
	response.NotifySuccess(w, req, "OK")
}
