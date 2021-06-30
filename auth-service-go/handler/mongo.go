package handler

import (
	"github.com/op/go-logging"
	"github.com/patrickmn/go-cache"
	"github.ibm.com/gftn/world-wire-services/utility/database"
	"net/http"
	"time"
)

var LOGGER = logging.MustGetLogger("auth-handlers")

type AuthOperations struct {
	session *database.MongoClient
	dbName  string
	c       *cache.Cache
}

func CreateAuthServiceOperations() (AuthOperations, error) {
	authOP := AuthOperations{}

	client, err := database.InitializeIbmCloudConnection()
	if err != nil {
		LOGGER.Errorf("IBM Cloud Mongo DB connection failed! %s", err)
		panic("IBM Cloud Mongo DB connection failed! " + err.Error())
	}

	authOP.session = client
	authOP.c = cache.New(5*time.Minute, 10*time.Minute)
	LOGGER.Infof("\t* CreateAuthServiceOperations DB is set")

	return authOP, nil
}

func (op *AuthOperations) ServiceCheck(w http.ResponseWriter, r *http.Request) {
	LOGGER.Debugf("%+v", w.Header())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))
	return
}
