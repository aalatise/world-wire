package utility

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/globalsign/mgo/bson"
	"github.com/op/go-logging"
	"github.com/IBM/world-wire/gftn-models/model"
	"github.com/IBM/world-wire/utility/database"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"go.mongodb.org/mongo-driver/mongo"
)

var LOGGER = logging.MustGetLogger("database-client")

func AddAccountCursor(client *database.MongoClient, accountName, cursor string) error {
	LOGGER.Infof("Adding cursor %v to %v account", cursor, accountName)

	coll, ctx := client.GetCollection()

	item := CursorData{
		Account:       accountName,
		Cursor:        cursor,
		ParticipantId: os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME),
	}
	input, err := json.Marshal(item)
	if err != nil {
		LOGGER.Error(err)
		return err
	}
	var bdoc interface{}
	err = bson.UnmarshalJSON(input, &bdoc)
	if err != nil {
		LOGGER.Error(err)
		return err
	}
	_, err = coll.InsertOne(ctx, &bdoc)
	if err != nil {
		LOGGER.Error(err)
		return err
	}

	// id := res.InsertedID
	LOGGER.Info("Cursor successfully added!")
	return nil
}

func GetAccountCursor(client *database.MongoClient, accountName string) (*model.Cursor, error) {

	LOGGER.Infof("Getting account cursor for account %v", accountName)
	collection, ctx := client.GetCollection()

	var result CursorData
	err := collection.FindOne(ctx, bson.M{"participant_id": os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME), "account_name": accountName}).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			LOGGER.Errorf("Can't find matching documents %v", err.Error())
		}
		return nil, err
	}
	LOGGER.Infof("found document %+v", result)

	var cursor model.Cursor
	cursor.Cursor = result.Cursor

	return &cursor, nil
}

func UpdateAccountCursor(client *database.MongoClient, accountName, cursor string) error {

	collection, ctx := client.GetCollection()
	// find the document for which the participant_id field matches participant_id and set the cursor to "cursor"
	// specify the Upsert option to insert a new document if a document matching the filter isn't found
	filter := bson.M{"participant_id": os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME), "account_name": accountName}
	update := bson.M{"$set": bson.M{"account_cursor": cursor}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
	}

	if result.MatchedCount == 0 {
		return errors.New("Cannot find a matching document")
	}

	return nil
}
