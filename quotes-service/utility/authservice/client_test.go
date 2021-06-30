package authservice

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
)

func TestRequestSigning(t *testing.T) {
	part := model.Participant{}
	URL := "http://localhost:8888"
	part.URLCallback = &URL

	csc := Client{
		HTTP: &http.Client{Timeout: time.Second * 10},
	}
	Convey("Successful get caller identity", t, func() {
		// So(err, ShouldBeNil)
		// So(signedXdr, ShouldNotBeNil)

	})
}
