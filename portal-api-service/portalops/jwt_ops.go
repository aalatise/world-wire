package portalops

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/IBM/world-wire/utility/database"
	"github.com/IBM/world-wire/utility/response"
	"go.mongodb.org/mongo-driver/bson"
)

type JwtOps struct {
	dbClient  *database.MongoClient
	dbName    string
	collName  string
	dbTimeout time.Duration
}

func CreateJwtOps(dbClient *database.MongoClient, dbName string, collName string, dbTimeout time.Duration) JwtOps {
	jwtOps := JwtOps{dbClient, dbName, collName, dbTimeout}
	return jwtOps
}

func (ops JwtOps) GetJwtInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	institutionId := vars["institutionId"]
	results := []bson.M{}

	coll, ctx := ops.dbClient.GetSpecificCollection(ops.dbName, ops.collName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*ops.dbTimeout)
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{"institution": institutionId})
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

	logger.Infof("%d JWT(s) found", len(results))
	bytes, err := json.Marshal(results)
	if err != nil {
		logger.Errorf("Error marshalling data")
		response.NotifyWWError(w, r, http.StatusNotFound, "PORTAL-API-1000", err)
		return
	}

	response.Respond(w, http.StatusOK, bytes)
}
