package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//struct to hold driver instance, db name and collection name
type MongoDBConnect struct {
	mongoSession *mgo.Session
	database     string
}

/*
 *Create object to establish session to MongoDB for pool of socket connections
 */
func InitializeConnection(addrs []string, timeout int, authDatabase string, username string, password string, workDatabase string) (conn MongoDBConnect, err error) {
	conn = MongoDBConnect{}

	//creating DialInfo object to establish a session to MongoDB
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    addrs,
		Timeout:  time.Duration(timeout) * time.Second,
		Database: authDatabase,
		Username: username,
		Password: password,
	}

	conn.database = workDatabase

	//Creating a session object which creates a pool of socket connections to MongoDB
	conn.mongoSession, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		LOGGER.Error("Error while creating MongoDB socket connections pool: ", err)
		return
	}

	LOGGER.Infof("\t* Dialing MongoDB session for socket connections pool for the Database: [%s] in Mongo Host: [%s]", conn.database, addrs)
	conn.mongoSession.SetMode(mgo.Eventual, true)

	LOGGER.Infof("\t* MongoDB socket connections pool for [%s] initialized successfully.", conn.database)
	return
}

/*
 * Request a socket connection from the session and retrieve collection to process query.
 */
func (mongoDBConnection MongoDBConnect) GetSocketConn() (session *mgo.Session) {
	session = mongoDBConnection.mongoSession.Copy()
	return
}

/*
 * Get the collection object from the session and collection name provided
 */
func (mongoDBConnection MongoDBConnect) GetCollection(session *mgo.Session, collectionName string) (collection *mgo.Collection) {
	collection = session.DB(mongoDBConnection.database).C(collectionName)
	return
}

/*
 * Close the session when the goroutine exits
 */
func (mongoDBConnect MongoDBConnect) CloseSession() {
	defer mongoDBConnect.mongoSession.Close()
}

/*
 * Connect to Mongo Atlas
 */

func InitializeAtlasConnection(username string, password string, id string) (*mongo.Client, error) {

	LOGGER.Infof("\t* Establishing Mongo DB connection...")
	urlEncodedPassword := url.QueryEscape(password)
	envVersion := os.Getenv(global_environment.ENV_KEY_ENVIRONMENT_VERSION)
	ConnectionURI := "mongodb+srv://" + username + ":" + urlEncodedPassword + "@" + envVersion + "-" + id + ".mongodb.net/test?retryWrites=true"
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientAtlas, err := mongo.Connect(ctx, options.Client().ApplyURI(ConnectionURI))
	if err != nil {
		return &mongo.Client{}, err
	}

	// Check the connection
	err = clientAtlas.Ping(ctx, readpref.Primary())
	if err != nil {
		return &mongo.Client{}, err
	}

	LOGGER.Infof("\t* Mongo Atlas DB connection established!")
	return clientAtlas, nil
}

/*
 * Connect to IBM Cloud Mongo
 */

type MongoClient struct {
	Session *mongo.Client
}

func InitializeIbmCloudConnection() (*MongoClient, error) {
	LOGGER.Infof("\t* Establishing Mongo DB connection...")

	username := os.Getenv(global_environment.ENV_KEY_DB_USER)
	password := os.Getenv(global_environment.ENV_KEY_DB_PWD)
	hosts := os.Getenv(global_environment.ENV_KEY_MONGO_ID)
	dbName := os.Getenv(global_environment.ENV_KEY_MONGO_DB_NAME)
	dbCollection := os.Getenv(global_environment.ENV_KEY_MONGO_COLLECTION_NAME)
	dbTimeout := os.Getenv(global_environment.ENV_KEY_DB_TIMEOUT)

	if dbCollection == "" || dbName == "" || username == "" || password == "" || hosts == "" || dbTimeout == "" {
		errMsg := "Error reading DB table, required credentials not set"
		LOGGER.Errorf(errMsg)
		panic(errMsg)
	}

	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)

	credential := options.Credential{
		Username: username,
		Password: password,
	}

	// mongo driver expects to read Root CA from local file, so we need to disable ssl , then construct tls config ourselves
	ConnectionURI := "mongodb://" + hosts + "/ibmclouddb?replicaSet=replset&ssl=false"

	// read cert from vault/secrets
	var cert64 string
	var exists bool
	if cert64, exists = os.LookupEnv(global_environment.ENV_KEY_MONGO_CONNECTION_CERT); !exists {
		return nil, errors.New("Cannot locate mongo cert")
	}
	cert, _ := base64.StdEncoding.DecodeString(cert64)

	// parsing cert & append into mongo connection options
	tlsConfig := &tls.Config{}

	// decode certs
	certBlock, err := loadCert([]byte(cert))
	if err != nil {
		return nil, err
	}

	// parse and validate cert format
	certs, err := x509.ParseCertificate(certBlock)
	if err != nil {
		return nil, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()

	// append cert as RootCAs
	tlsConfig.RootCAs.AddCert(certs)
	mongoClient := &options.ClientOptions{TLSConfig: tlsConfig}

	clientOpts := mongoClient.ApplyURI(ConnectionURI).SetAuth(credential)

	var client = &MongoClient{}

	// connecting
	client.Session, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Session.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	LOGGER.Infof("\t* Mongo IBM Cloud connection established!")
	return client, nil
}

func (client *MongoClient) GetCollection() (*mongo.Collection, context.Context) {
	dbTimeout, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_DB_TIMEOUT))
	dbName := os.Getenv(global_environment.ENV_KEY_MONGO_DB_NAME)
	dbCollection := os.Getenv(global_environment.ENV_KEY_MONGO_COLLECTION_NAME)
	LOGGER.Infof("\t* Getting collection: %s from DB %s", dbCollection, dbName)
	collection := client.Session.Database(dbName).Collection(dbCollection)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(dbTimeout))
	return collection, ctx
}

func (client *MongoClient) GetSpecificCollection(dbName, dbCollection string) (*mongo.Collection, context.Context) {
	if dbName == "" || dbCollection == "" {
		LOGGER.Errorf("No DB/Collection specified")
		return nil, context.TODO()
	}
	dbTimeout, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_DB_TIMEOUT))
	LOGGER.Infof("\t* Getting collection: %s from DB %s", dbCollection, dbName)
	collection := client.Session.Database(dbName).Collection(dbCollection)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(dbTimeout))
	return collection, ctx
}

func loadCert(data []byte) ([]byte, error) {
	var certBlock *pem.Block

	for certBlock == nil {
		if data == nil || len(data) == 0 {
			return nil, errors.New(".pem file must have both a CERTIFICATE and an RSA PRIVATE KEY section")
		}

		block, rest := pem.Decode(data)
		if block == nil {
			return nil, errors.New("invalid .pem file")
		}

		switch block.Type {
		case "CERTIFICATE":
			if certBlock != nil {
				return nil, errors.New("multiple CERTIFICATE sections in .pem file")
			}

			certBlock = block
		}

		data = rest
	}

	return certBlock.Bytes, nil
}

// used when return result will be array
func ParseResult(cursor *mongo.Cursor, ctx context.Context) ([]byte, error) {
	var interfaces []interface{}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var temp map[string]interface{}
		err := cursor.Decode(&temp)
		if err != nil {
			return []byte{}, errors.New("Encounter error while decoding cursor")
		}
		interfaces = append(interfaces, temp)
	}
	if err := cursor.Err(); err != nil {
		return []byte{}, errors.New("Encounter error while traversing through cursor")
	}

	bytes, err := json.Marshal(interfaces)
	if err != nil {
		return []byte{}, errors.New("Encounter error while marshaling mongo result")
	}

	return bytes, nil
}
