package handler

import (
	"github.com/golang-jwt/jwt"
	"github.com/IBM/world-wire/auth-service-go/idtoken"
	"github.com/IBM/world-wire/auth-service-go/permission"
	"github.com/IBM/world-wire/auth-service-go/totp"
	"github.com/IBM/world-wire/auth-service-go/utility/stringutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

func (op *AuthOperations) HandleGenerateIDToken(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("===== Generate ID Token for portal user")

	uid := r.Header.Get("UID")
	email := r.Header.Get("email")

	LOGGER.Debugf("UID: %s, Email: %s", uid, email)

	var user permission.User
	id, _ := primitive.ObjectIDFromHex(uid)
	userCollection, userCtx := op.session.GetSpecificCollection(PortalDBName, UserCollection)
	err := userCollection.FindOne(userCtx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		LOGGER.Errorf("Error during user query: %s", err.Error())
		totp.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	// Generate JWT ID Token
	jti := stringutil.RandStringRunes(26, false)
	standardClaim := jwt.StandardClaims{
		Audience:  email,
		ExpiresAt: time.Now().Add(time.Minute * 60).Unix(),
		Id:        jti,
		IssuedAt:  time.Now().Unix(),
		Issuer:    "ww-auth",
		NotBefore: time.Now().Unix(),
		Subject:   uid,
	}

	idToken := idtoken.ClaimsIDToken{
		StandardClaims: standardClaim,
		UID:            uid,
		Email:          email,
	}

	tokenSecret := stringutil.RandStringRunes(64, false)
	encodedToken, err := idtoken.CreateAndSign(idToken, tokenSecret)
	if err != nil {
		LOGGER.Errorf("Unable to create the ID token: %s", err.Error())
		idtoken.ResponseError(w, http.StatusInternalServerError, err)
		return
	}

	jwtSecure := idtoken.TokenSecure{
		Secret: tokenSecret,
		JTI:    jti,
		Number: 0,
		UID:    uid,
	}

	LOGGER.Debugf("Generate/Update ID token record for user: %s", uid)
	secureCollection, secureCtx := op.session.GetSpecificCollection(AuthDBName, IDTokenSecureCollection)
	update := secureCollection.FindOneAndUpdate(secureCtx, bson.M{"uid": uid}, bson.M{"$set": &jwtSecure})
	if update.Err() != nil {
		LOGGER.Warningf("User ID token record not found, create a new one: %+v", update.Err())
		_, err = secureCollection.InsertOne(secureCtx, jwtSecure)
		if err != nil {
			LOGGER.Errorf("Insert id token secure failed:  %s", err.Error())
			idtoken.ResponseError(w, http.StatusInternalServerError, err)
			return
		}
	}

	LOGGER.Debugf("Successfully generate/update ID token")
	idtoken.ResponseSuccess(w, encodedToken)
	return
}

