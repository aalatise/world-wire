package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.ibm.com/gftn/world-wire-services/utility/database"
	"github.ibm.com/gftn/world-wire-services/utility/response"
	"go.mongodb.org/mongo-driver/bson"
)

type Authentication struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateAuthentication(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) Authentication {
	authentication := Authentication{dbClient, dbName, collName, dbTimeout}
	return authentication
}

func (auth Authentication) getSecretKey(id string) (string, error) {
	result := bson.M{}

	coll, ctx := auth.dbClient.GetSpecificCollection(auth.dbName, auth.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*auth.dbTimeout)
	defer cancel()

	err := coll.FindOne(ctx, bson.M{"jti": id}).Decode(&result)
	if err != nil {
		return "", err
	}

	key, ok := result["secret"].(string)
	if !ok {
		return "", errors.New("Cannot find a key")
	}

	return key, nil
}

func (auth Authentication) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Middleware - authenticating request")

		token, err := request.OAuth2Extractor.ExtractToken(r)
		if err != nil {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1001", errors.New("Invalid Token"))
			return
		}

		decodedToken, valid := auth.verifyToken(token)
		if valid {
			claims, ok := decodedToken.Claims.(*UserClaims)
			if !ok {
				logger.Infof("Unexpected claims")
				response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1001", errors.New("Unauthenticated request"))
				return
			}
			ctx := context.WithValue(r.Context(), "uid", claims.Uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1001", errors.New("Unauthenticated request"))
		}
	})
}

func (auth Authentication) verifyToken(token string) (*jwt.Token, bool) {

	// Check if the token is valid
	decodedToken, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {

		// Validating the signing method (algorithm)
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			logger.Infof("Unexpected signing algorithm")
			return nil, errors.New("Unexpected signing method")
		}

		if method.Alg() != "HS256" {
			logger.Infof("Unexpected signing algorithm")
			return nil, errors.New("Unexpected signing method")
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			logger.Infof("Unexpected claims")
			return nil, errors.New("Unexpected claims")
		}

		key, err := auth.getSecretKey(claims.Id)
		if err != nil {
			return nil, err
		}

		return []byte(key), nil
	})

	if err != nil {
		logger.Infof(err.Error())
		return nil, false
	}

	logger.Infof("Authenticated successfully!")
	return decodedToken, true
}
