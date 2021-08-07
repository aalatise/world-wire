package transaction

import (
	"encoding/base64"
	"encoding/json"
	"github.com/IBM/world-wire/utility/common"
	db "github.com/IBM/world-wire/utility/database"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"github.com/IBM/world-wire/payment/constant"
	"os"
	"strconv"
	"time"
)



type PaymentData struct {
	InstructionID    *string              `json:"instruction_id,omitempty" db:"instructionid"`
	TxData           *string              `json:"tx_data,omitempty" db:"txdata"`
	TxStatus         *string              `json:"tx_status,omitempty" db:"txstatus"`
	ResId            *string              `json:"res_id,omitempty" db:"resid"`
	TxDetail         Payment              `json:"tx_detail,omitempty" db:"txdetail"`
	TxDetail64       *string              `db:"txdetail"`
	CreatedTimeStamp int64                `json:"created_timestamp,omitempty" db:"created_timestamp"`
	UpdatedTimeStamp int64                `json:"updated_timestamp,omitempty" db:"updated_timestamp"`
}

// Create new tx record to DB
func (dbc *db.PostgreDatabaseClient) CreateTx(input *PaymentData) error {

	err := insertCheck(input)
	if err != nil {
		return err
	}

	var retryInterval time.Duration
	var retryTimes float64

	// default value of retry time interval is 5 seconds
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL); !exists {
		retryInterval = time.Duration(5)
	} else {
		temp, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL))
		retryInterval = time.Duration(temp)
	}

	// default value of retry times is 2
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES); !exists {
		retryTimes = 2
	} else {
		retryTimes, _ = strconv.ParseFloat(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES), 64)
	}

	LOGGER.Infof("Creating Transaction in Database")
	err = common.Retry(retryTimes, retryInterval*time.Second, func() error {

		LOGGER.Infof("Inserting Tx with instruction ID %v", *input.InstructionID)

		now := time.Now().UTC().UnixNano()
		input.CreatedTimeStamp = now
		input.UpdatedTimeStamp = now

		paymentStatuses, _ := json.Marshal(input.TxDetail)
		paymentStatuses64 := base64.StdEncoding.EncodeToString(paymentStatuses)
		sqlStatement := `INSERT INTO ` + dbc.Tablename + ` ( instructionid, txdata, txstatus, resid, txdetail, created_timestamp, updated_timestamp )
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING instructionid;`

		id := ""
		err = dbc.db.QueryRow(sqlStatement,
			*input.InstructionID,
			*input.TxData,
			*input.TxStatus,
			*input.ResId,
			paymentStatuses64,
			input.CreatedTimeStamp,
			input.UpdatedTimeStamp,
		).Scan(&id)

		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				if string(err.Code) == constant.POSTGRESQL_ERROR_CODE_DUPLICATE_ID {
					LOGGER.Errorf("Duplicate Instruction ID detected")
					return err
				}
			}
			dbc.CreateConnection()
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	LOGGER.Infof("Inserting Tx with instruction ID %v success!", *input.InstructionID)

	return nil
}

// Get tx
func (dbc *PostgreDatabaseClient) GetTx(instructionId string) (*PaymentData, error) {

	if instructionId == "" {
		return nil, errors.New("No instruction id is specified")
	}

	var retryInterval time.Duration
	var retryTimes float64

	// default value of retry time interval is 5 seconds
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL); !exists {
		retryInterval = time.Duration(5)
	} else {
		temp, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL))
		retryInterval = time.Duration(temp)
	}

	// default value of retry times is 2
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES); !exists {
		retryTimes = 2
	} else {
		retryTimes, _ = strconv.ParseFloat(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES), 64)
	}

	var rows *sqlx.Rows
	LOGGER.Infof("Retrieving Transaction from Database")
	err := common.Retry(retryTimes, retryInterval*time.Second, func() error {

		var err error
		LOGGER.Infof("Retrieving Tx with instruction ID %v", instructionId)

		sqlStatement := `SELECT * FROM ` + dbc.Tablename + ` WHERE instructionid=$1;`
		rows, err = dbc.db.Queryx(sqlStatement, instructionId)
		if err != nil {
			dbc.CreateConnection()
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		request := &PaymentData{}
		err = rows.StructScan(&request)
		if err != nil {
			return nil, err
		}

		if request.TxDetail64 == nil {
			return nil, errors.New("Encounter error while unmarshaling payment status")
		}

		paymentInfo, err := base64.StdEncoding.DecodeString(*request.TxDetail64)
		if err != nil {
			LOGGER.Error(err.Error())
			return nil, err
		}
		json.Unmarshal(paymentInfo, &request.TxDetail)
		LOGGER.Infof("Retrieving Tx with instruction ID %v success!", instructionId)

		return request, nil
	}

	return nil, errors.New("no record found")
}

func (dbc *PostgreDatabaseClient) UpdateTx(input *PaymentData) error {

	err := updateCheck(input)
	if err != nil {
		return err
	}
	LOGGER.Infof("Updating Tx with instruction ID %v", *input.InstructionID)

	input.UpdatedTimeStamp = time.Now().UTC().UnixNano()

	sqlStatement := "UPDATE transactions t SET updated_timestamp = $1 "

	if input.TxData != nil {
		sqlStatement = sqlStatement + ", txdata = '" + *input.TxData + "'"
	}

	if input.TxDetail != nil {
		paymentStatuses, _ := json.Marshal(input.TxDetail)
		paymentStatuses64 := base64.StdEncoding.EncodeToString(paymentStatuses)
		sqlStatement = sqlStatement + ", txdetail = '" + paymentStatuses64 + "'"
	}

	if input.ResId != nil {
		sqlStatement = sqlStatement + ", resid = '" + *input.ResId + "'"
	}

	if input.TxStatus != nil {
		sqlStatement = sqlStatement + ", txstatus = '" + *input.TxStatus + "'"
	}

	sqlStatement = sqlStatement + " WHERE t.instructionid = $2;"
	result, err := dbc.db.Exec(
		sqlStatement,
		input.UpdatedTimeStamp,
		*input.InstructionID,
	)
	if err != nil {
		return err
	}
	rowAffeced, err := result.RowsAffected()
	if rowAffeced > 1 {
		return errors.New("Updated more than one row. Expecting one or none")
	}

	LOGGER.Infof("Updating Tx with instruction ID %v success!", *input.InstructionID)

	return nil
}

func updateCheck(input *PaymentData) error {
	if input.InstructionID == nil {
		return errors.New("Instruction ID of this tx is empty")
	}
	/*
		if input.TxStatus == nil {
			return errors.New("Status of this tx is empty")
		}
	*/
	return nil
}

func insertCheck(input *PaymentData) error {
	if input.InstructionID == nil {
		return errors.New("Instruction ID of this tx is empty")
	}

	if input.TxStatus == nil {
		return errors.New("Status of this tx is empty")
	}

	if input.TxData == nil {
		return errors.New("Tx hash of this tx is empty")
	}

	if input.TxDetail == nil {
		return errors.New("Tx detail of this tx is empty")
	}

	if input.ResId == nil {
		return errors.New("Response ID of this tx is empty")
	}

	return nil
}
