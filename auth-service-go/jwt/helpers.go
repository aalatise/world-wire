package jwt

import (
	"errors"
	"fmt"
	jwt "github.com/golang-jwt/jwt"
	"github.com/op/go-logging"
	"net/http"
	"strconv"
	"time"
)

var LOGGER = logging.MustGetLogger("jwt-helper")

func CreateClaims(token Info, count int, iid, keyID string) IJWTTokenClaim {
	info := jwt.StandardClaims{
		Audience:  token.Aud,
		Id:        token.JTI,
		Subject:   iid,
	}

	claims := IJWTTokenClaim{
		info,
		token.Acc,
		token.Ver,
		token.IPs,
		token.Env,
		token.Enp,
		count,
	}

	return claims
}

func CreateAndSign(claims IJWTTokenClaim, hmacSampleSecret, keyID string) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
 	tokenHeader := token.Header
	newHeader := make(map[string]interface{})
 	for key, value := range tokenHeader {
 		newHeader[key] = value
	}
	newHeader["kid"] = keyID
	token.Header = newHeader

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(hmacSampleSecret))

	return tokenString, err
}

func Verify(tokenString, secret string) (*IJWTTokenClaim, bool) {
	t, err := jwt.ParseWithClaims(tokenString, &IJWTTokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		LOGGER.Errorf("Token verify failed: %s", err.Error())
		return nil, false
	}

	if claim, ok := t.Claims.(*IJWTTokenClaim); ok {
		LOGGER.Debugf("%+v", claim)
		return claim, true
	}

	return nil, false
}

func Parse(tokenString string) (*IJWTTokenClaim, string, error) {
	t, _ := jwt.ParseWithClaims(tokenString, &IJWTTokenClaim{}, nil)
	keyID := t.Header["kid"]

	if t == nil {
		LOGGER.Errorf("Token is malformed")
		return nil, "", errors.New("token is malformed")
	}

	if claim, ok := t.Claims.(*IJWTTokenClaim); ok {
		return claim, "", nil
	}

	return nil, keyID.(string), errors.New("failed to convert claim type")
}

func ResponseError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
}

func ResponseSuccess(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func InstitutionOwnsParticipantId(participantId string, institution Institution) bool {
	// check if the selected institutionId contains the selected participantId
	// ie: prevents a user from generating a token for a
	// participantId for which they don't have access rights
	// NOTE 1: for this endpoint to issue a token it assumes that the stack details
	// exist in database under /participants/{participantId}/nodes/{env}/...
	// NOTE 2: middleware prevents a user that is not associate with this institution
	// from calling this endpoint.

	for _, i := range institution.Nodes {
		if i.ParticipantId != "" && i.ParticipantId == participantId {
			return true
		}
	}

	return false
}

func VerifyWWTokenCustom(decodedToken IJWTTokenClaim, nFromDb int, jtiFromDb, compareIncomingIp, compareEndpoint, compareAccount string) (bool, string) {
	msg := "failed: "

	// check expiration date
	isNotExpired := false
	if decodedToken.VerifyExpiresAt(time.Now().Unix(), true) {
		isNotExpired = true
	} else {
		msg = msg + "isNotExpired:now=" + strconv.Itoa(int(time.Now().Unix())) + ",exp=" + strconv.Itoa(int(decodedToken.ExpiresAt)) + "; "
	}

	// not before date
	isNotBefore := false
	if decodedToken.VerifyNotBefore(time.Now().Unix(), true) {
		isNotBefore = true
	} else {
		msg = msg + "decodedToken:now=" + strconv.Itoa(int(time.Now().Unix())) + ",nbf=" + strconv.Itoa(int(decodedToken.NotBefore)) + "; "
	}

	// check if token has account in array
	hasAccount := false
	// only check account if provided
	if compareAccount != "" {
		if findSomething(decodedToken.Account, compareAccount) {
			hasAccount = true
		} else {
			msg = msg + "compareAccount:comparedTo=" + compareAccount + "; "
		}
	} else {
		hasAccount = true
	}

	// check if token has IP in array
	hasIP := false
	// only check ip if provided
	if compareIncomingIp != "" {
		if findSomething(decodedToken.IPs, compareIncomingIp) {
			hasIP = true
		} else {
			msg = msg + "compareIncomingIp:incomming=" + compareIncomingIp + "; "
		}
	} else {
		hasIP = true
	}

	// check if token has Endpoint in array
	hasEndpoint := false
	// only check endpoint if provided
	if compareEndpoint != "" {
		if findSomething(decodedToken.Endpoints, compareEndpoint) {
			hasEndpoint = true
		} else {
			msg = msg + "compareEndpoint=" + compareEndpoint + "; "
		}
	} else {
		hasEndpoint = true
	}

	// check if decoded jti matches the jti stored in the db
	matchingJti := false
	if decodedToken.Id == jtiFromDb {
		matchingJti = true
	} else {
		msg = msg + "matchingJti" + "; "
	}

	// check if token is on count
	isOnCount := false
	// check if the count in db is same as the one in the db
	if decodedToken.Number == nFromDb {
		isOnCount = true
	} else {
		msg = msg + "isOnCount" + "; "
	}

	// send response
	if isNotExpired && isNotBefore && hasAccount && hasIP && hasEndpoint && isOnCount && matchingJti {
		return true, "clear"
	}

	return false, msg
}

func findSomething(input []string, target string) bool {
	for _, ele := range input {
		if ele == target {
			return true
		}
	}

	return false
}