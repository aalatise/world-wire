package portalops

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	"github.ibm.com/gftn/world-wire-services/utility/database"
	"github.ibm.com/gftn/world-wire-services/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlocklistOps struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateBlocklistOps(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) BlocklistOps {
	blocklistOps := BlocklistOps{dbClient, dbName, collName, dbTimeout}
	return blocklistOps
}

func (ops BlocklistOps) GetBlocklistRequests(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	query := r.URL.Query()

	blocklistType := query.Get("type")

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

	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{"type": blocklistType}, findOptions)
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

	logger.Infof("%d blocklist request(s) found", len(results))
	bytes, err := json.Marshal(results)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops BlocklistOps) AddBlocklistRequest(w http.ResponseWriter, r *http.Request) {
	var blocklistReqModel model.BlocklistReq

	blocklistReq, err := resolveReqBody(r, &blocklistReqModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Insert data into db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.InsertOne(ctx, blocklistReq)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Blocklist request added")
}

func (ops BlocklistOps) RemoveBlocklistRequest(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)

	approvalId := vars["approvalId"]

	// Remove data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ops.dbTimeout)
	defer cancel()

	deleteResult, err := coll.DeleteOne(ctx, bson.D{{"approvalId", approvalId}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	logger.Infof("%d blocklist request(s) removed", deleteResult.DeletedCount)

	response.NotifySuccess(w, r, "Blocklist request removed")
}
