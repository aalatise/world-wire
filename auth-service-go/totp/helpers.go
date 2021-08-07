package totp

import (
	"encoding/base32"
	"encoding/json"
	"github.com/op/go-logging"
	"github.com/pquerna/otp/totp"
	"github.com/IBM/world-wire/auth-service-go/utility/stringutil"
	"net/http"
	"strings"
)

var LOGGER = logging.MustGetLogger("totp-helper")

func Create(accountName string, user User, newTOTP bool) (*Response, *User, bool, bool) {
	if !newTOTP {
		if user.Registered {
			LOGGER.Warningf("Account %s registered already!", accountName)
			response := &Response{
				Success:    true,
				Registered: true,
				Msg:        accountName + " registered!",
				Data: QRCode{
					QRCodeURI:   "",
					AccountName: accountName,
				},
			}
			return response, nil, true, true
		}
	}

	key := stringutil.RandStringRunes(32, false)

	// encoded will be the secret key, base32 encoded
	encoded := base32.StdEncoding.EncodeToString([]byte(key))

	// Google authenticator doesn't like equal signs
	encodedForGoogle := strings.ReplaceAll(encoded, "/=/g", "")

	secretKey, generateErr := totp.Generate(totp.GenerateOpts{
		Issuer: "WorldWireOTP",
		AccountName: accountName,
		Secret: []byte(encodedForGoogle),
	})

	if generateErr != nil {
		LOGGER.Errorf("Failed to generate TOTP: %s", generateErr.Error())
		return nil, nil, false, false
	}

	// to create a URI for a qr code (change totp to hotp if using hotp)
	res := &Response{
		Success:    true,
		Registered: false,
		Msg:        "creating TOTP, please confirm",
		Data: QRCode{
			QRCodeURI:   secretKey.URL(),
			AccountName: accountName,
		},
	}

	updatedTOTP := &User{
		UID:        user.UID,
		Secret:     secretKey.Secret(),
		Registered: false,
	}

	return res, updatedTOTP, true, false
}

func Check(user User, token TokenBody) bool {
	key := user.Secret

	if !totp.Validate(token.Token, key) {
		return false
	}

	return true
}

func ResponseError(w http.ResponseWriter, statusCode int, err error) {
	res := &Response{}
	res.Success = false
	res.Registered = false
	res.Msg = err.Error()

	result, _ := json.Marshal(res)

	w.WriteHeader(statusCode)
	w.Write(result)
}

func ResponseSuccess(w http.ResponseWriter, res *Response, statusCode int, msg string) {
	if res == nil {
		res = &Response{}
		res.Success = true
		res.Registered = true
		res.Msg = msg
	}

	result, _ := json.Marshal(res)

	w.WriteHeader(statusCode)
	w.Write(result)
}