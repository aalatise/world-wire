package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/IBM/world-wire/utility/database"
	"github.com/IBM/world-wire/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Authorization struct {
	dbClient           *database.MongoClient
	dbName             string
	userCollName       string
	permissionCollName string
	dbTimeout          time.Duration
}

func CreateAuthorization(dbClient *database.MongoClient, dbName string, userCollName string, permissionCollName string, dbTimeout time.Duration) Authorization {
	authorization := Authorization{dbClient, dbName, userCollName, permissionCollName, dbTimeout}
	return authorization
}

func (auth Authorization) AuthorizeSuperUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Middleware - authorizing super user")

		// Extract uid from request context
		uid, ok := r.Context().Value("uid").(string)
		if !ok {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1002", errors.New("Unauthorized"))
		}

		valid := auth.authorizeUser(uid, "super")

		if valid {
			next.ServeHTTP(w, r)
		} else {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1002", errors.New("Unauthorized"))
		}
	})
}

func (auth Authorization) AuthorizeParticipantUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Middleware - authorizing participant user")

		// Extract uid from request context
		uid, ok := r.Context().Value("uid").(string)
		if !ok {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1002", errors.New("Unauthorized"))
		}

		valid := auth.authorizeUser(uid, "participant")

		if valid {
			next.ServeHTTP(w, r)
		} else {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1002", errors.New("Unauthorized"))
		}
	})
}

func (auth Authorization) AuthorizeSuperOrParticipantUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Middleware - authorizing super or participant user")

		// Extract uid from request context
		uid, ok := r.Context().Value("uid").(string)
		if !ok {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1002", errors.New("Unauthorized"))
		}

		superValid := auth.authorizeUser(uid, "super")
		participantValid := auth.authorizeUser(uid, "participant")

		if superValid || participantValid {
			next.ServeHTTP(w, r)
		} else {
			response.NotifyWWError(w, r, http.StatusUnauthorized, "PORTAL-API-1002", errors.New("Unauthorized"))
		}
	})
}

func (auth Authorization) authorizeUser(id string, permission string) bool {

	if permission == "super" {

		user := bson.M{}

		userColl, ctx := auth.dbClient.GetSpecificCollection(auth.dbName, auth.userCollName)

		ctx, cancel := context.WithTimeout(ctx, time.Second*auth.dbTimeout)
		defer cancel()

		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			logger.Errorf(err.Error())
			return false
		}

		filter := bson.D{
			{"_id", objectId},
			{
				"$or", bson.A{
					bson.D{{"super_permission.roles.admin", true}},
					bson.D{{"super_permission.roles.manager", true}},
					bson.D{{"super_permission.roles.viewer", true}},
				}},
		}

		err = userColl.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			logger.Errorf(err.Error())
			return false
		}

		return true
	}

	if permission == "participant" {
		permission := bson.M{}

		permissionColl, ctx := auth.dbClient.GetSpecificCollection(auth.dbName, auth.permissionCollName)

		ctx, cancel := context.WithTimeout(ctx, time.Second*auth.dbTimeout)
		defer cancel()

		filter := bson.D{
			{"user_id", id},
			{
				"$or", bson.A{
					bson.D{{"roles.admin", true}},
					bson.D{{"roles.manager", true}},
					bson.D{{"roles.viewer", true}},
				}},
		}

		err := permissionColl.FindOne(ctx, filter).Decode(&permission)
		if err != nil {
			logger.Errorf(err.Error())
			return false
		}

		return true
	}

	return false
}

func (auth Authorization) CheckTOTPPasscode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// passcode := r.Header.Get("x-verify-code")
		// fmt.Println(passcode)
		next.ServeHTTP(w, r)
	})
}
