package portalops

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v9"
)

func getCollection(dbClient *mongo.Client, dbName string, collName string) *mongo.Collection {
	logger.Infof("\t* DB access: getting %s from %s", collName, dbName)
	coll := dbClient.Database(dbName).Collection(collName)
	return coll
}

func ConnectToDb(dbUser string, dbPwd string, mongoId string, dbTimeout time.Duration) (*mongo.Client, error) {
	logger.Infof("\t* Connecting to database")
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://dbUser:dbUserPassword@cluster0.a1vpb.mongodb.net/portal-db?retryWrites=true&w=majority",
	))

	if err != nil {
		return nil, err
	}

	return client, nil

}

func resolveReqBody(r *http.Request, dataModel interface{}) (interface{}, error) {
	var resolved interface{}
	// Decode request body
	bodyDecoder := json.NewDecoder(r.Body)
	err := bodyDecoder.Decode(&dataModel)
	if err != nil {
		return nil, err
	}

	// Validate data from request body
	validator := validator.New()
	err = validator.Struct(dataModel)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(dataModel)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &resolved)
	if err != nil {
		return nil, err
	}

	return resolved, nil
}
