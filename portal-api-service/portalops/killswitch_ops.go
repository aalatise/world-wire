package portalops

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	"github.ibm.com/gftn/world-wire-services/utility/database"
	"github.ibm.com/gftn/world-wire-services/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type KillswitchOps struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateKillswitchOps(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) KillswitchOps {
	killswitchOps := KillswitchOps{dbClient, dbName, collName, dbTimeout}
	return killswitchOps
}

func (ops KillswitchOps) GetKillswitchRequest(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)

	participantId := vars["participantId"]
	accountAddress := vars["accountAddress"]

	result := bson.M{}

	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	err := coll.FindOne(ctx, bson.D{{"participantId", participantId}, {"accountAddress", accountAddress}}).Decode(&result)

	// Return after the first document needed does not exist
	if err == mongo.ErrNoDocuments {
		response.Respond(w, http.StatusOK, []byte("null"))
		return
	}

	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	logger.Infof("killswitch request found")

	bytes, err := json.Marshal(result)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops KillswitchOps) AddKillswitchRequest(w http.ResponseWriter, r *http.Request) {
	var killswitchReqModel model.KillswitchReq

	killswitchReq, err := resolveReqBody(r, &killswitchReqModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Insert data into db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.InsertOne(ctx, killswitchReq)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Killswitch request added")
}

func (ops KillswitchOps) UpdateKillswitchRequest(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)

	participantId := vars["participantId"]
	accountAddress := vars["accountAddress"]

	var killswitchReqModel model.KillswitchReq

	killswitchReq, err := resolveReqBody(r, &killswitchReqModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	// Update data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.ReplaceOne(ctx, bson.D{{"participantId", participantId}, {"accountAddress", accountAddress}}, killswitchReq)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Killswitch request updated")
}

func (ops KillswitchOps) RemoveKillswitchRequest(w http.ResponseWriter, r *http.Request) {
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

	response.NotifySuccess(w, r, "Killswitch request removed")
}
