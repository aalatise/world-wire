package database

import (
	"os"
	"strconv"
	"testing"
	"time"

	global_environment "github.com/IBM/world-wire/utility/global-environment"
)

var instructionId = strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

func TestCreateConnection(t *testing.T) {
	port, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_PORT))
	pdg := PostgreDatabaseClient{
		Host:     os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_URL),
		Port:     port,
		User:     os.Getenv(global_environment.ENV_KEY_POSTGRESQL_USER),
		Password: os.Getenv(global_environment.ENV_KEY_POSTGRESQL_PWD),
		Dbname:   os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_NAME),
	}
	err := pdg.CreateConnection()
	if err != nil {
		LOGGER.Error(err)
	}
	pdg.CloseConnection()

}
