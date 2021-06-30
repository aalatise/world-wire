package vault

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRead(t *testing.T) {
	var vault *Vault
	var err error
	Convey("init connection", t, func() {
		vault, err = InitializeVault()
		So(err, ShouldBeNil)
	})

	Convey("create secret", t, func() {
		path, _ := os.Executable()
		b, _ := ioutil.ReadFile(path + "data/cred.json")

		input := make(map[string]interface{})

		json.Unmarshal(b, &input)
		err = vault.Create(input, "sandbox/test")
		So(err, ShouldBeNil)
	})

	Convey("Appending secret", t, func() {
		input := make(map[string]interface{})
		input["ping2"] = "pong"
		err = vault.Create(input, "sandbox/test")
		So(err, ShouldBeNil)

	})

	Convey("Updating secret", t, func() {
		input := make(map[string]interface{})
		input["ping2"] = "pong2"
		err = vault.Update(input, "sandbox/test")
		So(err, ShouldBeNil)
	})

	Convey("Verifying value", t, func() {
		r, _ := vault.Read("initialize")
		So(r["ping2"], ShouldEqual, "pong2")
	})

	Convey("Deleting secret", t, func() {

		err = vault.Delete("sandbox/test")
		So(err, ShouldBeNil)
	})

}
