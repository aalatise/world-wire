package database

import (
	"errors"
	"fmt"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgreDatabaseClient struct {
	Host      string
	Port      int
	User      string
	Password  string
	Dbname    string
	db        *sqlx.DB
	Tablename string
}

//CreateConnection opens DB connection
func (dbc *PostgreDatabaseClient) CreateConnection() error {

	port, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_PORT))
	dbc.Host = os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_URL)
	dbc.Port = port
	dbc.User = os.Getenv(global_environment.ENV_KEY_POSTGRESQL_USER)
	dbc.Password = os.Getenv(global_environment.ENV_KEY_POSTGRESQL_PWD)
	dbc.Dbname = os.Getenv(global_environment.ENV_KEY_POSTGRESQL_DB_NAME)
	dbc.Tablename = os.Getenv(global_environment.ENV_KEY_POSTGRESQL_TABLE_NAME)

	if dbc.Host == "" || dbc.Port == 0 || dbc.User == "" || dbc.Password == "" || dbc.Dbname == "" || dbc.Tablename == "" {
		return errors.New("Missing necessary parameters to connect to PostgreSQL")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require",
		dbc.Host, dbc.Port, dbc.User, dbc.Password, dbc.Dbname)
	db, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	LOGGER.Info("Successfully connected!")
	dbc.db = db
	return nil
}

//CloseConnection closes DB connection
func (dbc *PostgreDatabaseClient) CloseConnection() {
	dbc.db.Close()
}
