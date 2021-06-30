package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/patrickmn/go-cache"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/environment"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/permission"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/sso"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

func (op *AuthOperations) HandlePortalLoginTOTP(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("===== Portal Login TOTP")

	email := r.Header.Get("email")
	LOGGER.Debugf("Email: %s", email)

	if email == "" {
		LOGGER.Errorf("Email is empty")
		sso.ServerError(w, errors.New("email is empty"))
		return
	}

	// Get user info from database
	var UID string
	var user permission.User
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	err := collection.FindOne(ctx, bson.M{"profile": bson.M{"email": strings.ToLower(email)}}).Decode(&user)
	if err != nil {
		// no user with this email exists, so create a new user
		LOGGER.Warning("No user with this email exists, so create a new user")
		uid := primitive.NewObjectIDFromTimestamp(time.Now())
		newUser := permission.User{
			UID:     uid,
			Profile: permission.Profile{Email: strings.ToLower(email)},
		}

		// create new user data
		insertedID, err := collection.InsertOne(ctx, newUser)
		if err != nil {
			LOGGER.Errorf("Insert new user data failed:  %+v", err)
			sso.ServerError(w, err)
			return
		}
		id := insertedID.InsertedID.(primitive.ObjectID)
		UID = id.Hex()
	} else {
		UID = user.UID.Hex()
	}

	LOGGER.Debugf("uid: %s", UID)
	r.Header.Set("uid", UID)

	op.HandleGenerateIDToken(w, r)
	return
}

func (op *AuthOperations) HandleSSOToken(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("===== SSO Token")
	// Get the user account name from the header
	email := r.Header.Get("email")
	LOGGER.Debugf("Email: %s", email)

	if email == "" {
		LOGGER.Errorf("Email is empty")
		sso.ServerError(w, errors.New("email is empty"))
		return
	}

	// Get user info from database
	var uid string
	var user permission.User
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	err := collection.FindOne(ctx, bson.M{"profile": bson.M{"email": strings.ToLower(email)}}).Decode(&user)
	if err != nil {
		// no user with this email exists, so create a new user
		LOGGER.Warning("No user with this email exists, so create a new user")
		objectID := primitive.NewObjectIDFromTimestamp(time.Now())
		newUser := permission.User{
			UID:     objectID,
			Profile: permission.Profile{Email: strings.ToLower(email)},
		}

		// create new user data
		_, err := collection.InsertOne(ctx, newUser)
		if err != nil {
			LOGGER.Errorf("Insert new user data failed:  %+v", err)
			sso.ServerError(w, err)
			return
		}

		uid = objectID.Hex()
	} else {
		uid = user.UID.Hex()
	}

	LOGGER.Debugf("uid: %s", uid)

	// store the email into the cache
	op.c.Set(string(time.Now().Unix()), email, cache.DefaultExpiration)

	// 1: check if user has registered 2FA username
	var tUser totp.User
	totpCollection, totpCtx := op.session.GetSpecificCollection(PortalDBName, TOTPCollection)
	err = totpCollection.FindOne(totpCtx, bson.M{"uid": user.UID.Hex()}).Decode(&tUser)
	if err != nil || !tUser.Registered {
		LOGGER.Warningf("TOTP not found redirect for registration:  %+v", err)
		registerData := struct {
			Email string `json:"email"`
		}{Email: email}
		data, _ := json.Marshal(registerData)
		registerDataEncoded := sso.RedirectData{
			Data: base64.StdEncoding.EncodeToString(data),
		}
		redirectTraffic(w, r, "2fa/register", registerDataEncoded)
		return
	}

	LOGGER.Debugf("TOTP record found, redirect for verification")
	redirectTraffic(w, r, "2fa/verify", sso.RedirectData{Data:""})
	return
}

func redirectTraffic(w http.ResponseWriter, r *http.Request, redirect string, queryObj sso.RedirectData) {
	LOGGER.Debugf("===== redirect traffic to %s", redirect)

	siteRoot := os.Getenv(environment.ENV_KEY_PORTAL_DOMAIN)
	LOGGER.Debugf("site root: %s", siteRoot)
	if queryObj.Data == "" {
		//w.WriteHeader(http.StatusOK)
		http.Redirect(w, r, siteRoot + "/" + redirect, http.StatusTemporaryRedirect)
		return
	}

	val := reflect.Indirect(reflect.ValueOf(queryObj))
	queryString := ""
	for n := 0; n < val.Type().NumField(); n++ {
		str := strings.ToLower(val.Type().Field(n).Name) + "=" + queryObj.Data
		queryString += str
		queryString += "&"
	}

	LOGGER.Debugf(queryString)
	//w.WriteHeader(http.StatusOK)
	http.Redirect(w, r, siteRoot + "/" + redirect + "/?" + queryString, http.StatusTemporaryRedirect)
	return
}