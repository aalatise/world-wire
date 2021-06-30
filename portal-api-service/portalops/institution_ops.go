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

type InstitutionOps struct {
	dbClient            *database.MongoClient
	dbName              string
	institutionCollName string
	permissionCollName  string
	dbTimeout           time.Duration
}

func CreateInstitutionOps(dbClient *database.MongoClient, dbName string, institutionCollName string, permissionCollName string, dbTimeout time.Duration) InstitutionOps {
	institutionOps := InstitutionOps{dbClient, dbName, institutionCollName, permissionCollName, dbTimeout}
	return institutionOps
}

func (ops InstitutionOps) GetInstitutions(w http.ResponseWriter, r *http.Request) {
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

	institutions := []bson.M{}

	findOptions := options.Find()
	findOptions.SetSkip(offset).SetLimit(limit)

	institutionColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.institutionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	cursor, err := institutionColl.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	err = cursor.All(ctx, &institutions)
	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	logger.Infof("%d institution(s) found", len(institutions))

	// Include participant permissions of each institution
	permissionColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.permissionCollName)

	for _, institution := range institutions {

		permissions := []bson.M{}

		institutionInfo, ok := institution["info"].(bson.M)
		if !ok {
			response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
			return
		}

		institutionId := institutionInfo["institutionId"]

		cursor, err := permissionColl.Find(ctx, bson.D{{"institution_id", institutionId}})
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

		institution["participant_permissions"] = permissions
	}

	bytes, err := json.Marshal(institutions)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops InstitutionOps) GetInstitution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	institutionIdOrSlug := vars["institutionIdOrSlug"]

	institution := bson.M{}

	institutionColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.institutionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	err := institutionColl.FindOne(ctx, bson.D{
		{"$or",
			bson.A{
				bson.D{{"info.institutionId", institutionIdOrSlug}},
				bson.D{{"info.slug", institutionIdOrSlug}}},
		}}).Decode(&institution)

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

	logger.Infof("Institution found")

	institutionInfo, ok := institution["info"].(bson.M)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	institutionId := institutionInfo["institutionId"]

	permissions := []bson.M{}

	permissionColl, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.permissionCollName)

	cursor, err := permissionColl.Find(ctx, bson.D{{"institution_id", institutionId}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	err = cursor.All(ctx, &permissions)
	if err != nil {
		logger.Errorf("Error parsing data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	// Include participant permissions for an institution
	if len(permissions) > 0 {
		institution["participant_permissions"] = permissions
	}

	bytes, err := json.Marshal(institution)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops InstitutionOps) AddInstitution(w http.ResponseWriter, r *http.Request) {
	var institutionModel model.Institution

	institution, err := resolveReqBody(r, &institutionModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Insert data into db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.institutionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	insertOneResult, err := coll.InsertOne(ctx, institution)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	insertedId := insertOneResult.InsertedID.(primitive.ObjectID)
	_, err = coll.UpdateOne(ctx, bson.D{{"_id", insertedId}}, bson.D{{"$set", bson.D{{"info.institutionId", insertedId.Hex()}}}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Institution added")
}

func (ops InstitutionOps) UpdateInstitution(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)

	institutionId := vars["institutionId"]

	var institutionModel model.Institution

	institution, err := resolveReqBody(r, &institutionModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(institutionId)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	// Insert data into db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.institutionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.ReplaceOne(ctx, bson.D{{"_id", objectId}}, institution)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Institution updated")
}

func (ops InstitutionOps) RemoveInstitution(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)

	institutionId := vars["institutionId"]

	objectId, err := primitive.ObjectIDFromHex(institutionId)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	// Remove data into db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.institutionCollName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.DeleteOne(ctx, bson.D{{"_id", objectId}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Institution removed")
}
