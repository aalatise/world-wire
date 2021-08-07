package portalops

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/IBM/world-wire/utility/database"
	"github.com/IBM/world-wire/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TxOps struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateTxOps(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) TxOps {
	txOps := TxOps{dbClient, dbName, collName, dbTimeout}
	return txOps
}

func (ops TxOps) GetTransactions(w http.ResponseWriter, r *http.Request) {
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

	cursor, err := coll.Find(ctx, bson.D{
		{
			"$or", bson.A{
				bson.D{{"RFIID", participantId}},
				bson.D{{"OFIID", participantId}},
				bson.D{{"fitoficctnonpiidata.transactiondetails.ofi_id", participantId}},
				bson.D{{"fitoficctnonpiidata.transactiondetails.rfi_id", participantId}}}}}, findOptions)
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

	logger.Infof("%d transaction(s) found", len(results))
	bytes, err := json.Marshal(results)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}
