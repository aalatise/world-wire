package transaction

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"

	global_environment "github.com/IBM/world-wire/utility/global-environment"
	utility_db "github.com/IBM/world-wire/utility/database"
)

var instructionId = strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

func TestCreateTx(t *testing.T) {
	port, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_PORT))
	fmt.Println(port)
	pdg := utility_db.PostgreDatabaseClient{
		Host:      os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_URL),
		Port:      port,
		User:      os.Getenv(global_environment.ENV_KEY_POSTGRESQL_USER),
		Password:  os.Getenv(global_environment.ENV_KEY_POSTGRESQL_PWD),
		Dbname:    os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_NAME),
		Tablename: os.Getenv(global_environment.ENV_KEY_POSTGRESQL_TABLE_NAME),
	}
	err := pdg.CreateConnection()
	if err != nil {
		LOGGER.Error(err)
		t.FailNow()
	}
	defer pdg.CloseConnection()
	Convey("Create Tx ", t, func(c C) {
		payloadRequest, _ := ioutil.ReadFile("./unit-test-data/create.json")
		paymentReq := &PaymentData{}
		err = json.Unmarshal(payloadRequest, &paymentReq)
		if err != nil {
			LOGGER.Debug(err)
		}
		*paymentReq.InstructionID = instructionId
		paymentReq.CreatedTimeStamp = time.Now().UTC().UnixNano()

		err := pdg.CreateTx(paymentReq)
		So(err, ShouldEqual, nil)
	})
}

func TestUpdateTx(t *testing.T) {
	port, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_PORT))
	pdg := PostgreDatabaseClient{
		Host:      os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_URL),
		Port:      port,
		User:      os.Getenv(global_environment.ENV_KEY_POSTGRESQL_USER),
		Password:  os.Getenv(global_environment.ENV_KEY_POSTGRESQL_PWD),
		Dbname:    os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_NAME),
		Tablename: os.Getenv(global_environment.ENV_KEY_POSTGRESQL_TABLE_NAME),
	}
	err := pdg.CreateConnection()
	if err != nil {
		LOGGER.Error(err)
		t.FailNow()
	}
	defer pdg.CloseConnection()
	Convey("Update Tx and its status ", t, func(c C) {
		payloadRequest, _ := ioutil.ReadFile("./unit-test-data/update.json")
		paymentReq := &PaymentData{}
		err = json.Unmarshal(payloadRequest, &paymentReq)
		if err != nil {
			LOGGER.Debug(err)
		}
		*paymentReq.InstructionID = instructionId
		paymentReq.UpdatedTimeStamp = time.Now().UTC().UnixNano()
		err := pdg.UpdateTx(paymentReq)
		So(err, ShouldEqual, nil)
	})
}

func TestGetTx(t *testing.T) {
	port, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_PORT))
	fmt.Println(port)
	pdg := PostgreDatabaseClient{
		Host:      os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_URL),
		Port:      port,
		User:      os.Getenv(global_environment.ENV_KEY_POSTGRESQL_USER),
		Password:  os.Getenv(global_environment.ENV_KEY_POSTGRESQL_PWD),
		Dbname:    os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_NAME),
		Tablename: os.Getenv(global_environment.ENV_KEY_POSTGRESQL_TABLE_NAME),
	}
	err := pdg.CreateConnection()
	if err != nil {
		LOGGER.Error(err)
		t.FailNow()
	}
	defer pdg.CloseConnection()
	Convey("Get Tx and updated status ", t, func(c C) {
		payloadRequest, _ := ioutil.ReadFile("./unit-test-data/update.json")
		paymentReq := &PaymentData{}
		err = json.Unmarshal(payloadRequest, &paymentReq)
		if err != nil {
			LOGGER.Debug(err)
		}
		result, err := pdg.GetTx(instructionId)
		So(err, ShouldEqual, nil)
		So(*result.InstructionID, ShouldEqual, instructionId)
		So(*result.TxStatus, ShouldEqual, *paymentReq.TxStatus)

	})
}

type Operations struct {
	session *MongoClient
}

func TestGetTx2(t *testing.T) {

	client, err := InitializeIbmCloudConnection()
	if err != nil {
		LOGGER.Errorf("IBM Cloud Mongo DB connection failed!")
		panic(err.Error())
	}
	op := Operations{session: client}
	collection, ctx := op.session.GetCollection()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		LOGGER.Debugf("Error during query")
	}
	bytes, err := ParseResult(cursor, ctx)
	if err != nil {
		LOGGER.Debugf("Error parsing mongo data")
	}
	var results []bson.M
	_ = json.Unmarshal(bytes, &results)
	fmt.Println(results)
}
