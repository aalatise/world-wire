package handler

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/permission"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/totp"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
)

func (op *AuthOperations) HandleTOTPCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Get the user account name which will be the email address
	email := vars["accountName"]

	LOGGER.Debugf("Account name is: %s", email)

	if email == "" {
		LOGGER.Errorf("Account name is empty")
		totp.ResponseError(w, http.StatusInternalServerError, errors.New("account name is empty"))
		return
	}

	var user permission.User
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	err := collection.FindOne(ctx, bson.M{"profile": bson.M{"email": strings.ToLower(email)}}).Decode(&user)
	if err != nil {
		LOGGER.Errorf("Error during user query: %s", err.Error())
		totp.ResponseError(w, http.StatusInternalServerError, err)
		return
	}
	LOGGER.Debugf("%+v", user)

	newTOTP := false
	var tUser totp.User
	totpCollection, totpCtx := op.session.GetSpecificCollection(PortalDBName, TOTPCollection)
	err = totpCollection.FindOne(totpCtx, bson.M{"uid": user.UID.Hex()}).Decode(&tUser)
	if err != nil {
		LOGGER.Warningf("Error during totp query: %s", err.Error())
		newTOTP = true
		tUser.UID = user.UID.Hex()
	}
	LOGGER.Debugf("%+v", tUser)

	res, totpOject, success, registered := totp.Create(email, tUser, newTOTP)
	if success && registered {
		LOGGER.Warning("TOTP for user was already created and registered")
		totp.ResponseSuccess(w, nil, http.StatusCreated, "TOTP for user was already created and registered")
		return
	}

	if success && !registered  {
		if newTOTP {
			_, err = totpCollection.InsertOne(totpCtx, totpOject)
			if err != nil {
				LOGGER.Errorf("Insert TOTP was not successful: %s", err.Error())
				totp.ResponseError(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			updateResult, err := totpCollection.UpdateOne(totpCtx, bson.M{"uid": user.UID.Hex()}, bson.M{"$set": &totpOject})
			if err != nil || updateResult.MatchedCount < 1 {
				LOGGER.Errorf("Update TOTP was not successful: %+v", err)
				totp.ResponseError(w, http.StatusInternalServerError, err)
				return
			}
		}

		LOGGER.Debugf("TOTP successfully created")
		totp.ResponseSuccess(w, res, http.StatusOK, "TOTP create success")
		return
	}

	totp.ResponseError(w, http.StatusForbidden, errors.New("can't create TOTP"))
	return
}

func (op *AuthOperations) HandleTOTPConfirm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["accountName"]
	var token totp.TokenBody

	LOGGER.Debugf("Account name is: %s", email)
	if email == "" {
		LOGGER.Errorf("Account name is empty")
		totp.ResponseError(w, http.StatusInternalServerError, errors.New("account name is empty"))
		return
	}

	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		LOGGER.Errorf("Error while validating token body :  %v", err)
		totp.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	var user permission.User
	userCollection, userCtx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	err = userCollection.FindOne(userCtx, bson.M{"profile": bson.M{"email": strings.ToLower(email)}}).Decode(&user)
	if err != nil {
		LOGGER.Errorf("Error during user query: %s", err.Error())
		totp.ResponseError(w, http.StatusInternalServerError, err)
		return
	}
	uid := user.UID.Hex()

	var tUser totp.User
	collection, ctx := op.session.GetSpecificCollection(PortalDBName, TOTPCollection)
	err = collection.FindOne(ctx, bson.M{"uid": uid}).Decode(&tUser)
	if err != nil {
		LOGGER.Errorf("Error during totp query: %s", err.Error())
		totp.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	if !totp.Check(tUser, token) {
		LOGGER.Errorf("Error verify TOTP token")
		totp.ResponseError(w, http.StatusInternalServerError, errors.New("TOTP verification failed"))
		return
	}

	// valid token
	tUser.Registered = true

	updateResult, err := collection.UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": &tUser})
	if err != nil || updateResult.MatchedCount < 1 {
		LOGGER.Errorf("Update participant was not successful  %+v", err)
		totp.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	LOGGER.Debugf("OTP verification success")
	totp.ResponseSuccess(w, nil, http.StatusOK, "OTP verification success")
	return
	//http.Redirect(w, r, "/idtoken/generate/{accountName}", http.StatusPermanentRedirect)
}

func (op *AuthOperations) CheckAccountNameMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LOGGER.Debugf("----- Middleware: check TOTP account name")
		vars := mux.Vars(r)
		// Get the user account name which will be the email address
		accountName := vars["accountName"]
		email := r.Header.Get("email")

		if email == "" {
			value, found := op.c.Get(r.RemoteAddr)
			if !found {
				LOGGER.Errorf("unauthorized, missing email")
				totp.ResponseError(w, http.StatusUnauthorized, errors.New("unauthorized, missing email"))
				return
			}
			email = value.(string)
			op.c.Delete(r.RemoteAddr)
		}

		LOGGER.Debugf("TOTP Middleware check: accountName=%s, email=%s", accountName, email)

		if accountName == email {
			next.ServeHTTP(w, r)
		}

		LOGGER.Errorf("unauthorized")
		totp.ResponseError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	})
}