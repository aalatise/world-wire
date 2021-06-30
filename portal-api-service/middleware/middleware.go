package middleware

import "go.mongodb.org/mongo-driver/mongo"

func getCollection(dbClient *mongo.Client, dbName string, collName string) *mongo.Collection {
	logger.Infof("\t* DB access: getting %s from %s", collName, dbName)
	coll := dbClient.Database(dbName).Collection(collName)
	return coll
}
