package participant

import (
	"github.com/IBM/world-wire/automation-service/constant"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (op DeploymentOperations) createPartsicipantWhitelistEntry(participantID string) error {
	LOGGER.Infof("Updating Participant(%v) status in MongoDB", participantID)

	collection, ctx := op.Session.GetSpecificCollection(constant.GFTNMongoDBName, constant.WhitelistMongoDBCollectionName)
	// First we find if there is an existing entry (possibly from a failed deployment previously)
	participantExists, err := op.checkWhiteListEntryExists(participantID)
	if err != nil {
		LOGGER.Errorf("Error checking whitelist entry for participant with ID:%s\nError: %v", participantID, err)
		return err
	}

	if !participantExists {
		whitelistEntry := WhitelistEntry{
			ID:        participantID,
			Whitelist: []string{},
		}
		_, err := collection.InsertOne(ctx, whitelistEntry)
		if err != nil {
			LOGGER.Errorf("Error inserting entry for participant with ID:%s\nError: %v", participantID, err)
			return err
		}
		LOGGER.Infof("Successfully created whitelist entry for participant with ID: %s", participantID)
		return nil
	}

	LOGGER.Infof("Whitelist entry already exists for participant with ID: %s", participantID)
	return nil
}

func (op DeploymentOperations) checkWhiteListEntryExists(participantID string) (bool, error) {
	collection, ctx := op.Session.GetSpecificCollection(constant.GFTNMongoDBName, constant.WhitelistMongoDBCollectionName)
	filter := bson.M{
		"_id": participantID,
	}
	var result WhitelistEntry
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			LOGGER.Infof("Whitelist entry for participant with ID:%s not found", participantID)
			return false, nil
		}
		LOGGER.Errorf("Error finding whitelist entry for participant with ID:%s\nError: %v", participantID, err)
		return false, err
	}
	return true, nil
}

// WhitelistEntry - Whitelist entry for MongoDB
type WhitelistEntry struct {
	ID        string   `bson:"_id"`
	Whitelist []string `bson:"whitelist"`
}
