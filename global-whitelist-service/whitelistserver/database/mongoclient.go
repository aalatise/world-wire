package database

import (
	"fmt"

	"github.ibm.com/gftn/world-wire-services/utility/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbClient struct {
	//svc *dynamodb.DynamoDB
	svc *database.MongoClient

	// ProfileFile    string
	// ProfileName    string
}
type Item struct {
	Participant string   `json:"_id" bson:"_id"`
	Whitelist   []string `json:"whitelist" bson:"whitelist"`
}

func (dc *DbClient) CreateConnection() error {

	LOGGER.Infof("\t* CreateParticipantRegistryOperations connecting Mongo DB ")

	client, err := database.InitializeIbmCloudConnection()
	if err != nil {
		LOGGER.Errorf("Mongo Atlas DB connection failed! %s", err)
		panic("Mongo Atlas DB connection failed! " + err.Error())
	}
	dc.svc = client
	LOGGER.Infof("\t* CreateParticipantRegistryOperations DB is set")

	return nil
}

func (dc *DbClient) DeleteWhitelistParticipant(participantID, wlParticipant string) error {

	collection, ctx := dc.svc.GetCollection()
	var updatedDocument bson.M

	err := collection.FindOneAndUpdate(ctx, bson.M{"_id": participantID}, bson.M{"$pull": bson.M{"whitelist": wlParticipant}}).Decode(&updatedDocument)
	if err != nil {
		LOGGER.Debugf("Error during DeleteWhitelistParticipant query")
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			LOGGER.Errorf("the filter did not match any documents in the collection")
		}
		return err
	}

	LOGGER.Infof("Delete success, result: %v", updatedDocument)
	return nil

}

func (dc *DbClient) AddWhitelistParticipant(participant, wlparticipant string) error {
	collection, ctx := dc.svc.GetCollection()

	// specify the Upsert option to insert a new document if a document matching the filter isn't found
	opts := options.Update().SetUpsert(true)

	result, err := collection.UpdateOne(ctx, bson.M{"_id": participant}, bson.M{"$addToSet": bson.M{"whitelist": wlparticipant}}, opts)
	if err != nil {
		LOGGER.Errorf("Error during AddWhitelistParticipant query: %v", err)
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			LOGGER.Errorf("the filter did not match any documents in the collection")
		}
		return err
	}

	LOGGER.Infof("Add success, result: %v", result)
	return nil
}

func (dc *DbClient) GetWhiteListParicipants(participantID string) ([]string, error) {

	collection, ctx := dc.svc.GetCollection()

	var result Item
	fmt.Println(participantID)
	err := collection.FindOne(ctx, bson.M{"_id": participantID}).Decode(&result)
	if err != nil {
		LOGGER.Errorf("Error during GetWhiteListParicipants query: %v", err)

		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			LOGGER.Errorf("the filter did not match any documents in the collection")
		}
		return nil, err
	}

	var whitelist []string

	for _, str := range result.Whitelist {
		whitelist = append(whitelist, str)
	}
	return whitelist, nil
}
