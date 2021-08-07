package portalops

import (
	"context"
	"encoding/json"
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

type AccountOps struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateAccountOps(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) AccountOps {
	accountOps := AccountOps{dbClient, dbName, collName, dbTimeout}
	return accountOps
}

func (ops AccountOps) GetAccountRequests(w http.ResponseWriter, r *http.Request) {
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

	logger.Infof("%d account request(s) found", len(results))
	bytes, err := json.Marshal(results)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops AccountOps) AddAccountRequest(w http.ResponseWriter, r *http.Request) {
	var accountReqModel model.AccountReq

	accountReq, err := resolveReqBody(r, &accountReqModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Insert data into db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)
	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()
	_, err = coll.InsertOne(ctx, accountReq)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Account request added")
}

func (ops AccountOps) UpdateAccountRequest(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)

	approvalId := vars["approvalId"]

	var accountReqModel model.AccountReq

	accountReq, err := resolveReqBody(r, &accountReqModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Update data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.ReplaceOne(ctx, bson.D{{"approvalIds", bson.A{approvalId}}}, accountReq)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Account request updated")
}

func (ops AccountOps) RemoveAccountRequest(w http.ResponseWriter, r *http.Request) {
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

	response.NotifySuccess(w, r, "Account request removed")
}
