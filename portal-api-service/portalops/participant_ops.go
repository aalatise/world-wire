package portalops

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/IBM/world-wire/gftn-models/model"
	"github.com/IBM/world-wire/utility/database"
	"github.com/IBM/world-wire/utility/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ParticipantOps struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateParticipantOps(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) ParticipantOps {
	participantOps := ParticipantOps{dbClient, dbName, collName, dbTimeout}
	return participantOps
}

func (ops ParticipantOps) GetParticipantApproval(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from request
	vars := mux.Vars(r)
	approvalId := vars["approvalId"]

	result := bson.M{}

	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(approvalId)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	err = coll.FindOne(ctx, bson.M{"_id": objectId}).Decode(&result)

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

	logger.Infof("Participant approval found")
	bytes, err := json.Marshal(result)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}

func (ops ParticipantOps) UpdateParticipantApproval(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	approvalId := vars["approvalId"]

	var approvalUpdateModel model.ApprovalUpdate

	approvalUpdate, err := resolveReqBody(r, &approvalUpdateModel)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", err)
		return
	}

	approval, ok := approvalUpdate.(map[string]interface{})
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	uidApprove, ok := approval["uid_approve"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	status, ok := approval["status"].(string)
	if !ok {
		response.NotifyWWError(w, r, http.StatusBadRequest, "PORTAL-API-1000", errors.New("Type checking failed"))
		return
	}

	objectId, err := primitive.ObjectIDFromHex(approvalId)
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	// Update data from db
	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	_, err = coll.UpdateOne(ctx, bson.D{{"_id", objectId}}, bson.D{{"$set", bson.D{{"uid_approve", uidApprove}, {"status", status}}}})
	if err != nil {
		logger.Errorf(err.Error())
		response.NotifyWWError(w, r, http.StatusInternalServerError, "PORTAL-API-1000", err)
		return
	}

	response.NotifySuccess(w, r, "Participant approval updated")
}
