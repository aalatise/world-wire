package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	authtesting "github.ibm.com/gftn/world-wire-services/utility/testing"
)

func TestAuthForExternalEndpoint(t *testing.T) {
	a := App{}
	a.initRoutes()
	Convey("Testing authorization for external endpoints...", t, func() {
		authtesting.InitAuthTesting()
		err := a.Router.Walk(authtesting.AuthWalker)
		So(err, ShouldBeNil)
		err = a.InternalRouter.Walk(authtesting.AuthWalker)
		So(err, ShouldBeNil)
	})
}
