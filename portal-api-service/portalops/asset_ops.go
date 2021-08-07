package portalops

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/IBM/world-wire/gftn-models/model"
	"github.com/IBM/world-wire/utility/database"
	"github.com/IBM/world-wire/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AssetOps struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateAssetOps(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) AssetOps {
	assetOps := AssetOps{dbClient, dbName, collName, dbTimeout}
	return assetOps
}

func (ops AssetOps) GetAssetRequests(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)
	query := r.URL.Query()

	participantId := vars["participantId"]

	queryOffset := query.Get("offset")
	queryLimit := query.Get("limit")
	var offset int64 = 0
	var limit int64 = 0
	var err error

	if queryOffset != "" {
		offset, err = strconv.ParseInt(queryOffset, 10, 64)
		if err != nil {
			logger.Errorf(err.Error())
			response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
			return
		}
	}

	if queryLimit != "" {
		limit, err = strconv.ParseInt(queryLimit, 10, 64)
		if err != nil {
			logger.Errorf(err.Error())
			response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
			return
		}
	}

	results := []bson.M{}

	findOptions := options.Find()
	findOptions.SetSkip(offset).SetLimit(limit)

	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{"participantId": participantId}, findOptions)
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

	logger.Infof("%d asset request(s) found", len(results))
	bytes, err := json.Marshal(results)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops AssetOps) AddAssetRequest(w http.ResponseWriter, r *http.Request) {
	var assetReqModel model.AssetReq

	assetReq, err := resolveReqBody(r, &assetReqModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Insert data into db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.InsertOne(ctx, assetReq)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Asset request added")
}

func (ops AssetOps) UpdateAssetRequest(w http.ResponseWriter, r *http.Request) {

	var assetReqModel model.AssetReq

	assetReq, err := resolveReqBody(r, &assetReqModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	assetReqObj, ok := assetReq.(map[string]interface{})
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	participantId, ok := assetReqObj["participantId"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	assetType, ok := assetReqObj["asset_type"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	assetCode, ok := assetReqObj["asset_code"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	// Update data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.ReplaceOne(ctx, bson.D{{"participantId", participantId}, {"asset_type", assetType}, {"asset_code", assetCode}}, assetReq)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Asset request updated")
}

func (ops AssetOps) RemoveAssetRequest(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)

	approvalId := vars["approvalId"]

	// Remove data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err := coll.DeleteOne(ctx, bson.D{{"approvalIds", bson.A{approvalId}}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Asset request removed")
}
