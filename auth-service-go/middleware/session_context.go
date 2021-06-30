package middleware

import (
	"errors"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/jwt"
	"net/http"
	"time"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	gcontext "github.com/gorilla/context"
)

//Set context to pass some session data to the handler function call.
// Using gorilla mux context here to share context between middleware and handler function
//Reference: https://www.nicolasmerouze.com/share-values-between-middlewares-context-golang/
//and https://stackoverflow.com/questions/30506831/negroni-and-gorilla-context-clearhandler

// SessionContext : object to store token session
type SessionContext struct {
	ParticipantID string
	TimeTill      int64
	Account       []string
}

type key string

const ContextKey key = "sessionToken"

// GetSessionContext : Return the session context from jwt token, without the dependency of func JwtAuthorization.
func GetSessionContext(r *http.Request) (SessionContext, error) {
	token, _ := request.ParseFromRequest(r, request.OAuth2Extractor, func(token *jwt_go.Token) (interface{}, error) {
		return nil, nil
	}, request.WithClaims(&jwt.IJWTTokenClaim{}))
	if token == nil {
		return SessionContext{}, errors.New("Jwt Token Required")
	}
	claim, _ := token.Claims.(*jwt.IJWTTokenClaim)
	sessionContext, err := ParseContext(r, claim)
	if err != nil {
		return SessionContext{}, err
	}

	return sessionContext, nil
}

// GetIdentity : Return the participant ID of the caller/user
// this function assume the authentication and authorization check has already been performed and varified against participantID
func GetIdentity(req *http.Request) (string, error) {

	// check if all x-iid, x-pid ,x-fid exist. if yes then authentication and authorization checking shall already been done according to assumption
	iid := req.Header.Get("x-iid")
	pid := req.Header.Get("x-pid")
	fid := req.Header.Get("x-fid")
	if iid != "" && fid != "" && pid != "" {
		return pid, nil
	}
	// if one of the header x-iid, x-pid ,x-fid does not exist, fall back grabbing the identity from jwt.
	LOGGER.Info("Header x-iid, x-pid, x-fid not all filled, grab participant ID from jwt...")
	sessionContext, err := GetSessionContext(req)
	if err != nil {
		LOGGER.Error("Failed to get jwt...")
		LOGGER.Error(err)
		return "", err
	}
	identity := sessionContext.ParticipantID
	LOGGER.Info("Caller Identity: ", identity)
	return identity, nil
}

// GetTimeTill :  Returns JWT token time till
// this function assume the authentication and authorization check has already been performed and varified against participantID
func GetTimeTill(req *http.Request) (int64, error) {

	// check if all x-iid, x-pid ,x-fid exist. if yes then authentication and authorization checking shall already been done according to assumption
	iid := req.Header.Get("x-iid")
	pid := req.Header.Get("x-pid")
	fid := req.Header.Get("x-fid")
	if iid != "" && fid != "" && pid != "" {
		//don't check this for fid based tokens
		return 1, nil
	}
	// if one of the header x-iid, x-pid ,x-fid does not exist, fall back grabbing the time from jwt.
	LOGGER.Info("Header x-iid, x-pid, x-fid not all filled, grab value from jwt...")
	sessionContext, err := GetSessionContext(req)
	if err != nil {
		LOGGER.Error("Failed to get jwt...")
		LOGGER.Error(err)
		return 0, err
	}
	timeTill := sessionContext.TimeTill
	LOGGER.Info("timeTill: ", timeTill)
	return timeTill, nil
}

// ParseContext : Parse jwt token
func ParseContext(r *http.Request, claims *jwt.IJWTTokenClaim) (SessionContext, error) {

	var ctx SessionContext

	//Set participant id
	aud := claims.Audience
	if aud != "" {
		ctx.ParticipantID = aud
	} else {
		return ctx, errors.New("jwt claims key aud not found")
	}
	//Set Time
	timeTill := claims.ExpiresAt
	if timeTill > 0 {
		//Calculate remaining time, jwt token uses unix time
		ctx.TimeTill = timeTill - time.Now().Unix()
	} else {
		return ctx, errors.New("jwt claims key exp not found")
	}

	//Set account name list
	accountToken := claims.Account
	if len(accountToken) > 0 {
		ctx.Account = accountToken
	} else {
		return ctx, errors.New("jwt claims key acc not found")
	}
	return ctx, nil
}

// ClearContext : clear context
func ClearContext(r *http.Request) {
	gcontext.Clear(r)
}

// HasAccount : has correct account access in the jwt token
func HasAccount(accountName string, req *http.Request) bool {

	// check if all x-iid, x-pid ,x-fid exist. if yes then authentication and authorization checking shall already been done according to assumption
	iid := req.Header.Get("x-iid")
	pid := req.Header.Get("x-pid")
	fid := req.Header.Get("x-fid")
	if iid != "" && fid != "" && pid != "" {
		// for fid or portal only tokens, don't check account level access, its taken care by maker checker flow
		return true
	}
	// if one of the header x-iid, x-pid ,x-fid does not exist, fall back grabbing the identity from jwt.
	LOGGER.Info("Header x-iid, x-pid, x-fid not all filled, grab participant ID from jwt...")

	sessionContext, err := GetSessionContext(req)
	if err != nil {
		LOGGER.Error("Failed to get jwt...")
		LOGGER.Error(err)
		return false
	}
	for _, v := range sessionContext.Account {
		if v == accountName {
			return true
		}
	}
	return false
}
