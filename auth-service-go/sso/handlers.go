package sso

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/sso/pkg/oauth"
	"github.ibm.com/gftn/world-wire-services/auth-service-go/utility/stringutil"
	"net/http"
	"strings"
)

func Home(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("Rendering home page")
}

func HandleIBMIdLogin(w http.ResponseWriter, r *http.Request) {
	referURL := "/sso/token"//r.Header.Get("referer")
	oauthStateString := stringutil.GenerateUUID()
	oauthStateString = strings.TrimSuffix(oauthStateString, "\n")

	oauth.OAuthStateMap[oauthStateString] = struct {
		State       string
		OriginalURL string
	}{State: oauthStateString, OriginalURL: referURL}
	url := oauth.IBMIdOAuthConfig.AuthCodeURL(oauthStateString)
	LOGGER.Debugf("URL: %+v", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleIBMIdLoginCallback(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("IBM RESPONSE")
	var err error

	defer func() {
		if err != nil {
			LOGGER.Errorf(err.Error())
			w.Write([]byte("Login failed"))
		}
	}()

	state := r.FormValue("state")
	if state != oauth.OAuthStateMap[state].State {
		err = errors.New("Invalid state")
		return
	}

	token, err := oauth.IBMIdOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		LOGGER.Errorf(err.Error())
		return
	}

	jwtToken, _, err := new(jwt.Parser).ParseUnverified(token.Extra("id_token").(string), jwt.MapClaims{})
	if err != nil {
		LOGGER.Errorf(err.Error())
		return
	}

	if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
		LOGGER.Debugf("Email: %+v", claims["preferred_username"])
		initUserSession(w, r, claims["preferred_username"].(string))
	} else {
		err = errors.New("Parsing of token failed")
		return
	}

	http.Redirect(w, r, oauth.OAuthStateMap[state].OriginalURL, http.StatusMovedPermanently)
}

func Token(w http.ResponseWriter, r *http.Request) {
	Render(w, r, "login/login.page.tmpl", nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	Render(w, r, "login/login.page.tmpl", nil)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	name := "user-sessions"
	session, err := SessionStore.Get(r, name)
	if err != nil {
		LOGGER.Errorf(err.Error())
	}
	session.Values["userId"] = nil
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		LOGGER.Errorf(err.Error())
	}

	w.Write([]byte("Logout successfully"))
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	var data struct {

	}
	Render(w, r, "404.page.tmpl", &data)
}

func initUserSession(w http.ResponseWriter, r *http.Request, email string) error {
	name := "user-sessions"
	session, err := SessionStore.Get(r, name)
	if err != nil {
		return err
	}
	session.Values["userId"] = email
	session.Save(r, w)
	LOGGER.Debugf("email stored to session: %v", email)
	return nil
}