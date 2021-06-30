package portalops

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	"github.ibm.com/gftn/world-wire-services/utility/database"
	"github.ibm.com/gftn/world-wire-services/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserOps struct {
	dbClient            *database.MongoClient
	dbName              string
	userCollName        string
	institutionCollName string
	permissionCollName  string
	dbTimeout           time.Duration
}

func CreateUserOps(dbClient *database.MongoClient, dbName string, userCollName string, institutionCollName string, permissionCollName string, dbTimeout time.Duration) UserOps {
	userOps := UserOps{dbClient, dbName, userCollName, institutionCollName, permissionCollName, dbTimeout}
	return userOps
}

func (ops UserOps) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	// Find a user
	user := bson.M{}

	userColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.userCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	err = userColl.FindOne(ctx, bson.M{"_id": objectId}).Decode(&user)

	// Return after the first document needed does not exist
	if err == mongo.ErrNoDocuments {
		response.Respond(w, http.StatusOK, []byte("null"))
		return
	}

	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	logger.Infof("User found")

	// Find permissions for a user
	permissions := []bson.M{}

	permissionColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.permissionCollName)

	cursor, err := permissionColl.Find(ctx, bson.M{"user_id": userId})

	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	err = cursor.All(ctx, &permissions)
	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	logger.Infof("%d permission(s) found", len(permissions))

	// Insert permissions into a user
	if len(permissions) > 0 {
		user["participant_permissions"] = permissions
	}

	bytes, err := json.Marshal(user)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops UserOps) GetSuperUsers(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	query := r.URL.Query()

	queryOffset := query.Get("offset")
	queryLimit := query.Get("limit")
	var offset int64 = 0
	var limit int64 = 0
	var err error

	if queryOffset != "" {
		offset, err = strconv.ParseInt(queryOffset, 10, 64)
		if err != nil {
			logger.Errorf(err.Error())
			response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
			return
		}
	}

	if queryLimit != "" {
		limit, err = strconv.ParseInt(queryLimit, 10, 64)
		if err != nil {
			logger.Errorf(err.Error())
			response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
			return
		}
	}

	results := []bson.M{}

	findOptions := options.Find()
	findOptions.SetSkip(offset).SetLimit(limit)

	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.userCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	filter := bson.D{{
		"$or", bson.A{
			bson.D{{"super_permission.roles.admin", true}},
			bson.D{{"super_permission.roles.manager", true}},
			bson.D{{"super_permission.roles.viewer", true}},
		}}}

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	err = cursor.All(ctx, &results)
	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	logger.Infof("%d super user(s) found", len(results))
	bytes, err := json.Marshal(results)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops UserOps) UpdateSuperUser(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	var userRoleModel model.UserRole

	userRole, err := resolveReqBody(r, &userRoleModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	user, ok := userRole.(map[string]interface{})
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	userId, ok := user["userId"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	role, ok := user["role"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Update data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.userCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	rolesUnsetUpdate := bson.D{{"$set", bson.D{{"super_permission.roles", bson.D{{"admin", false}, {"manager", false}, {"viewer", false}}}}}}
	rolesSetUpdate := bson.D{{"$set", bson.D{{"super_permission.roles." + role, true}}}}

	writeModels := []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetFilter(bson.D{{"_id", objectId}}).SetUpdate(rolesUnsetUpdate).SetUpsert(false),
		mongo.NewUpdateOneModel().SetFilter(bson.D{{"_id", objectId}}).SetUpdate(rolesSetUpdate).SetUpsert(false),
	}

	// Set the bulkWrite to execute the update in order
	bulkWriteOptions := options.BulkWrite().SetOrdered(true)
	_, err = coll.BulkWrite(ctx, writeModels, bulkWriteOptions)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Super user updated")
}

func (ops UserOps) RemoveSuperUser(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)
	userId := vars["userId"]

	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	// Update data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.userCollName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.UpdateOne(ctx, bson.D{{"_id", objectId}}, bson.D{{"$set", bson.D{{"super_permission.roles", bson.D{{"admin", false}, {"manager", false}, {"viewer", false}}}}}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Super user removed")
}

func (ops UserOps) GetParticipantUsers(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)
	institutionId := vars["institutionId"]

	query := r.URL.Query()
	queryOffset := query.Get("offset")
	queryLimit := query.Get("limit")
	var offset int64 = 0
	var limit int64 = 0
	var err error

	if queryOffset != "" {
		offset, err = strconv.ParseInt(queryOffset, 10, 64)
		if err != nil {
			logger.Errorf(err.Error())
			response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
			return
		}
	}

	if queryLimit != "" {
		limit, err = strconv.ParseInt(queryLimit, 10, 64)
		if err != nil {
			logger.Errorf(err.Error())
			response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
			return
		}
	}

	findOptions := options.Find()
	findOptions.SetSkip(offset).SetLimit(limit)

	// Find permissions for an institution
	permissions := []bson.M{}

	permissionColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.permissionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	cursor, err := permissionColl.Find(ctx, bson.M{"institution_id": institutionId}, findOptions)
	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	err = cursor.All(ctx, &permissions)
	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	logger.Infof("%d permission(s) found", len(permissions))

	// Find all the users and insert their corresponding participant permissions
	users := []bson.M{}

	userColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.userCollName)

	for _, permission := range permissions {

		user := bson.M{}

		userId, ok := permission["user_id"].(string)
		if !ok {
			response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
			return
		}

		objectId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			logger.Errorf(err.Error())
			response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
			return
		}

		err = userColl.FindOne(ctx, bson.M{"_id": objectId}).Decode(&user)
		if err != nil {
			logger.Errorf("Error parsing data")
			response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
			return
		}

		user["participant_permissions"] = []bson.M{permission}

		users = append(users, user)
	}

	logger.Infof("%d user(s) found", len(users))
	bytes, err := json.Marshal(users)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops UserOps) UpdateParticipantUser(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)
	userId := vars["userId"]
	institutionId := vars["institutionId"]

	var userRoleModel model.UserRole

	userRole, err := resolveReqBody(r, &userRoleModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	user, ok := userRole.(map[string]interface{})
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	role, ok := user["role"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	// Update or else upsert participant permission in db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.permissionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	rolesUnsetUpdate := bson.D{{"$set", bson.D{{"roles", bson.D{{"admin", false}, {"manager", false}, {"viewer", false}}}}}}
	rolesSetUpdate := bson.D{{"$set", bson.D{{"roles." + role, true}}}}

	writeModels := []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetFilter(bson.D{{"institution_id", institutionId}, {"user_id", userId}}).SetUpdate(rolesUnsetUpdate).SetUpsert(true),
		mongo.NewUpdateOneModel().SetFilter(bson.D{{"institution_id", institutionId}, {"user_id", userId}}).SetUpdate(rolesSetUpdate).SetUpsert(true),
	}

	// Set the bulkWrite to execute the update in order
	bulkWriteOptions := options.BulkWrite().SetOrdered(true)
	_, err = coll.BulkWrite(ctx, writeModels, bulkWriteOptions)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Participant user updated")
}

func (ops UserOps) RemoveParticipantUser(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)
	userId := vars["userId"]
	institutionId := vars["institutionId"]

	// Remove participant permission from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.permissionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err := coll.DeleteOne(ctx, bson.D{{"institution_id", institutionId}, {"user_id", userId}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Participant user removed")
}
