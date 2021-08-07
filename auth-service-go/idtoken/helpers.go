package idtoken

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/op/go-logging"
	"net/http"
)

var LOGGER = logging.MustGetLogger("idToken")

func CreateAndSign(claims ClaimsIDToken, hmacSampleSecret string) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(hmacSampleSecret))

	return tokenString, err
}

func Parse(tokenString string) (*ClaimsIDToken, error) {
	t, _ := jwt.ParseWithClaims(tokenString, &ClaimsIDToken{}, nil)

	if t == nil {
		LOGGER.Errorf("Token is malformed")
		return nil, errors.New("token is malformed")
	}

	if claim, ok := t.Claims.(*ClaimsIDToken); ok {
		return claim, nil
	}

	return nil, errors.New("failed to convert claim type")
}

func ResponseError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
	return
}

func ResponseSuccess(w http.ResponseWriter, msg string) {
	response := struct {
		Token string `json:"token"`
	}{Token: msg}

	res, _ := json.Marshal(response)

	w.WriteHeader(http.StatusOK)
	w.Write(res)
	return
}