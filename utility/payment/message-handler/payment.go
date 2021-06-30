package message_handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"github.com/stellar/go/xdr"
	"github.ibm.com/gftn/world-wire-services/utility/payment/environment"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils"
	"github.ibm.com/gftn/world-wire-services/utility/xmldsig"

	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/horizon"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/transaction"
	whitelist_handler "github.ibm.com/gftn/world-wire-services/utility/payment/utils/whitelist-handler"

	DB "github.ibm.com/gftn/world-wire-services/utility/database"
	"github.ibm.com/gftn/world-wire-services/utility/payment/client"
	message_converter "github.ibm.com/gftn/world-wire-services/utility/payment/message-converter"

	ibwf001struct "github.ibm.com/gftn/iso20022/ibwf00100101"
	pacs002struct "github.ibm.com/gftn/iso20022/pacs00200109"
	pacs008struct "github.ibm.com/gftn/iso20022/pacs00800107"
	ibwf001pbstruct "github.ibm.com/gftn/iso20022/proto/github.ibm.com/gftn/iso20022/ibwf00100101"

	pbstruct "github.ibm.com/gftn/iso20022/proto/github.ibm.com/gftn/iso20022/pacs00800107"
	blocklist_client "github.ibm.com/gftn/world-wire-services/administration-service/blocklist-client"
	"github.ibm.com/gftn/world-wire-services/utility/common"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/kafka"
	"github.ibm.com/gftn/world-wire-services/utility/payment/constant"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/database"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/parse"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/sendmodel"
)

// pacs.008 message handler at OFI side
func (op *PaymentOperations) Pacs008(pacs008 message_converter.Pacs008) ([]byte, error) {

	structData := pacs008.Message
	BIC := os.Getenv(environment.ENV_KEY_PARTICIPANT_BIC)
	target := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	msgName := constant.PACS008
	instructionId := string(*structData.Body.CdtTrfTxInf[0].PmtId.InstrId)

	originalGrpInf := &pacs002struct.OriginalGroupInformation29{
		OrgnlMsgId:   getReportMax35Text(string(*structData.Body.GrpHdr.MsgId)),
		OrgnlMsgNmId: getReportMax35Text(msgName),
	}

	// Message type for payment status log : credit_transfer
	msgType := constant.PAYMENT_TYPE_CREDIT_TRANSFER
	logHandler := transaction.InitiatePaymentLogOperation()

	/*
		write into DB
	*/
	// Initialize log handler and set the payment status to `INITIAL`
	logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_INITIAL)

	err := op.DbClient.CreateTx(&DB.PaymentData{
		InstructionID: &instructionId,
		TxData:        &constant.DATABASE_STATUS_EMPTY,
		TxStatus:      &constant.DATABASE_STATUS_PENDING,
		ResId:         &constant.DATABASE_STATUS_NONE,
		TxDetail:      &logHandler,
	})
	if err != nil {
		LOGGER.Errorf(err.Error())
		if err, ok := err.(*pq.Error); ok {
			if string(err.Code) == constant.POSTGRESQL_ERROR_CODE_DUPLICATE_ID {
				LOGGER.Errorf("Duplicate Instruction ID detected")
			}
		}
		report := op.ResponseHandler.CreatePacs002(BIC, instructionId, target, constant.STATUS_CODE_DUP_ID, originalGrpInf)
		return report, err
	}

	/*
		payload check
	*/
	// validate content in the pacs008 message and get all the necessary data from it
	xmlData, statsData := getCriticalInfoFromPacs008(structData, op.homeDomain)
	statusCode := xmlData.ErrorCode
	if statusCode != constant.STATUS_CODE_DEFAULT {
		LOGGER.Errorf("Something wrong with the transaction information")
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, statusCode, originalGrpInf)
		return report, errors.New("something wrong with the transaction information")
	}

	/*
		blocklist check
	*/

	// Create admin-service client for connecting to admin-service
	blockListClient := blocklist_client.Client{
		HTTPClient: &http.Client{Timeout: time.Second * 80},
		AdminUrl:   os.Getenv(global_environment.ENV_KEY_ADMIN_SVC_URL),
	}

	// preparing the data that need to be verify against the block-list
	var countries []string
	countries = append(countries, strconv.Quote(xmlData.OfiCountry))
	countries = append(countries, strconv.Quote(xmlData.RfiCountry))

	var currencies []string
	currencies = append(currencies, strconv.Quote(xmlData.CurrencyCode))

	var participants []string
	participants = append(participants, strconv.Quote(xmlData.OFIId))
	participants = append(participants, strconv.Quote(xmlData.RFIId))

	// validate block-list
	res, err := blockListClient.ValidateFromBlocklist(countries, currencies, participants)
	if err != nil {
		LOGGER.Errorf("%v", err)
		logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, err
	}
	if res == common.BlocklistDeniedString {
		LOGGER.Errorf("The transaction currency/country/institution is within the blocklist, transaction forbidden!")
		logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_BLOCKLIST, originalGrpInf)
		return report, errors.New("the transaction currency/country/institution is within the blocklist, transaction forbidden")
	}

	/*
		Check mutual whitelist
	*/
	LOGGER.Infof("Check whether RFI is in OFI's whitelist and vice versa")
	pkey, whiteListErr := op.whitelistHandler.CheckWhiteListParticipant(xmlData.OFIId, xmlData.RFIId, constant.EMPTY_STRING)
	if whiteListErr != nil {
		LOGGER.Errorf(whiteListErr.Error())
		logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, whiteListErr
	}

	if pkey == "" {
		errMsg := "OFI can not find RFI in whitelist and vice versa"
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_RFI_OR_OFI_NOT_IN_WL, originalGrpInf)
		return report, nil
	}
	LOGGER.Infof("Yes, RFI is in OFI's whitelist and vice versa")

	/*
		Parse the pacs008 message with signature into ProtoBuffer
	*/

	LOGGER.Infof("Start parsing Go struct to ProtoBuffer")
	protoBufData, parseErr := proto.Marshal(&pacs008.SendPayload)
	if parseErr != nil {
		errMsg := "Parse data to ProtoBuf error: " + parseErr.Error()
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, parseErr
	}
	LOGGER.Infof("Finished parsing Go struct to ProtoBuffer")

	// Process done with OFI side, now update the payment status
	logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_VALIDATION_SUCCESS)

	// dbData to be written into DB
	reportData := parse.CreatePacs008DbData(pacs008.Message.Body)
	dbData, _ := json.Marshal(reportData)
	base64DBData := parse.EncodeBase64(dbData)

	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &xmlData.InstructionId,
		TxData:        &base64DBData,
		TxStatus:      &constant.DATABASE_STATUS_PENDING,
		ResId:         &constant.DATABASE_STATUS_NONE,
		TxDetail:      &logHandler,
	})

	database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)

	/*
		Send the ProtoBuffer to the request topic of RFI on Kafka broker
	*/
	LOGGER.Infof("Start to send request to Kafka broker")
	kafkaErr := op.KafkaActor.Produce(xmlData.RFIId+kafka.REQUEST_TOPIC, protoBufData)
	if kafkaErr != nil {
		errMsg := "Error while submit message to Kafka broker: " + kafkaErr.Error()
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, kafkaErr
	}

	LOGGER.Infof("Successfully produce message to Kafka broker")
	LOGGER.Debug("-----------------------------------------------------------------")

	// Send status back to OFI
	report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_OFI_SEND_TO_KAFKA, originalGrpInf)

	return report, nil
}

// ibwf.001 message handler at RFI side
func (op *PaymentOperations) Ibwf001(ibwf001 message_converter.Ibwf001) ([]byte, error) {

	BIC := os.Getenv(environment.ENV_KEY_PARTICIPANT_BIC)
	target := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	structData := ibwf001.Message
	msgName := constant.IBWF001
	instrId := string(structData.Body.ResBody[0].Id)

	originalGrpInf := &pacs002struct.OriginalGroupInformation29{
		OrgnlMsgId:   getReportMax35Text(string(structData.Body.GrpHdr.MsgId)),
		OrgnlMsgNmId: getReportMax35Text(msgName),
	}

	// Initialize log handler and set the payment status to `INITIAL`
	msgType := constant.PAYMENT_TYPE_CREDIT_TRANSFER
	ibwf001LogHandler := transaction.InitiatePaymentLogOperation()
	ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_INITIAL)

	/*
		write into DB
	*/
	err := op.DbClient.CreateTx(&DB.PaymentData{
		InstructionID: &instrId,
		TxData:        &constant.DATABASE_STATUS_EMPTY,
		TxStatus:      &constant.DATABASE_STATUS_PENDING,
		ResId:         &constant.DATABASE_STATUS_NONE,
		TxDetail:      &ibwf001LogHandler,
	})
	if err != nil {
		LOGGER.Errorf(err.Error())
		report := op.ResponseHandler.CreatePacs002(BIC, instrId, target, constant.STATUS_CODE_DUP_ID, originalGrpInf)
		return report, err
	}

	/*
		retrieve necessary info from IBWF001
	*/
	xmlData, statusData, getDataErr := getCriticalInfoFromIBWF001(structData, op.prServiceURL, op.homeDomain, op.DbClient)
	errCode := xmlData.ErrorCode
	if getDataErr != nil {
		LOGGER.Error(getDataErr.Error())
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, errCode, originalGrpInf)
		return report, getDataErr
	}

	/*
		blocklist check
	*/
	// Create admin-service client for connecting to admin-service
	blockListClient := blocklist_client.Client{
		HTTPClient: &http.Client{Timeout: time.Second * 80},
		AdminUrl:   os.Getenv(global_environment.ENV_KEY_ADMIN_SVC_URL),
	}

	// preparing the data that need to be verify against the block-list
	var participants []string
	participants = append(participants, strconv.Quote(xmlData.OFIId))
	participants = append(participants, strconv.Quote(xmlData.RFIId))

	// validate block-list
	res, err := blockListClient.ValidateFromBlocklist([]string{}, []string{}, participants)
	if err != nil {
		LOGGER.Errorf("%v", err)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.OriginalMsgId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", ibwf001LogHandler, &op.fundHandler, statusData)
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, err
	}
	if res == common.BlocklistDeniedString {
		LOGGER.Errorf("The transaction currency/country/institution is within the blocklist, transaction forbidden!")
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.OriginalMsgId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", ibwf001LogHandler, &op.fundHandler, statusData)
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_BLOCKLIST, originalGrpInf)
		return report, errors.New("the transaction currency/country/institution is within the blocklist, transaction forbidden")
	}

	/*
		check mutual whitelist
	*/
	LOGGER.Infof("Check whether OFI is in RFI's whitelist and vice versa")
	pKey, whiteListErr := op.whitelistHandler.CheckWhiteListParticipant(xmlData.RFIId, xmlData.OFIId, constant.EMPTY_STRING)
	if whiteListErr != nil {
		LOGGER.Errorf(whiteListErr.Error())
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.OriginalMsgId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", ibwf001LogHandler, &op.fundHandler, statusData)
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, whiteListErr
	}

	if pKey == "" {
		errMsg := "RFI can not find OFI in whitelist and vice versa"
		LOGGER.Errorf(errMsg)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.OriginalMsgId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", ibwf001LogHandler, &op.fundHandler, statusData)
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_RFI_OR_OFI_NOT_IN_WL, originalGrpInf)
		return report, whiteListErr
	}
	LOGGER.Infof("Yes, OFI is in RFI's whitelist and vice versa")

	/*
		Parse the ibwf001 message with signature into ProtoBuffer
	*/
	LOGGER.Infof("Start parsing Go struct to ProtoBuffer")
	protoBufData, parseErr := proto.Marshal(&ibwf001.SendPayload)
	if parseErr != nil {
		errMsg := "Parse struct to ProtoBuf error " + parseErr.Error()
		LOGGER.Errorf(errMsg)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.OriginalMsgId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", ibwf001LogHandler, &op.fundHandler, statusData)
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, parseErr
	}
	LOGGER.Infof("Finished parsing Go struct to ProtoBuffer")

	ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_VALIDATION_SUCCESS)

	dbData := sendmodel.DBData{
		MessageId: string(structData.Body.GrpHdr.MsgId),
	}

	dbDataByte, _ := json.Marshal(dbData)
	base64DBData := parse.EncodeBase64(dbDataByte)

	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &xmlData.InstructionId,
		TxData:        &base64DBData,
		TxStatus:      &constant.DATABASE_STATUS_PENDING,
		ResId:         &constant.DATABASE_STATUS_NONE,
		TxDetail:      &ibwf001LogHandler,
	})
	database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.OriginalMsgId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", ibwf001LogHandler, &op.fundHandler, statusData)

	/*
		Send status back to RFI
	*/
	// Send the ProtoBuffer to the response topic of OFI on Kafka broker
	LOGGER.Infof("Start to send response to Kafka broker")
	kafkaErr := op.KafkaActor.Produce(xmlData.OFIId+kafka.RESPONSE_TOPIC, protoBufData)
	if kafkaErr != nil {
		LOGGER.Errorf("Error while submit message to Kafka broker %+v: ", kafkaErr.Error())
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_NONE,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_INIT, msgType, msgName, xmlData.OriginalMsgId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", ibwf001LogHandler, &op.fundHandler, statusData)
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, kafkaErr
	}
	LOGGER.Infof("Successfully produce message to Kafka broker")
	LOGGER.Debug("-----------------------------------------------------------")

	report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant.STATUS_CODE_RFI_SEND_TO_KAFKA, originalGrpInf)

	return report, nil
}

// pacs.008 message handler at RFI side
func RFI_Pacs008(sendPayload pbstruct.SendPayload, op *kafka.KafkaOpreations) {
	// Parse the ProtoBuffer into Go struct and reconstruct it into pacs008 message
	LOGGER.Infof("Parsing ProtoBuffer to XML")

	pacs008InstructionId := sendPayload.InstructionId
	reqMsgType := sendPayload.MsgType
	ofiId := sendPayload.OfiId
	originalMsgId := sendPayload.OriginalMsgId
	msgId := sendPayload.MsgId
	msgName := constant.PACS008

	standardType := constant.ISO20022
	paymentStatusMsgType := constant.PAYMENT_TYPE_CREDIT_TRANSFER
	rfiId := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	topicName := rfiId + "_" + kafka.TRANSACTION_TOPIC

	/*
		get pacs008 data from DB
	*/
	pacs008LogHandler := transaction.InitiatePaymentLogOperation()
	paymentData, dbErr := op.DbClient.GetTx(pacs008InstructionId)
	if dbErr != nil {
		LOGGER.Errorf("Encounter error while lookup tx data in DB: %v", dbErr)
		op.SendErrMsg(msgId, pacs008InstructionId, standardType, reqMsgType, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	if paymentData.TxData == nil || paymentData.TxDetail == nil {
		LOGGER.Errorf("The original message ID %v does not exist in DB", pacs008InstructionId)
		op.SendErrMsg(msgId, pacs008InstructionId, standardType, reqMsgType, ofiId, rfiId, constant.STATUS_CODE_WRONG_ORIGINAL_ID)
		return
	}
	pacs008LogHandler.PaymentStatuses = paymentData.TxDetail.PaymentStatuses
	LOGGER.Debugf("RFI side DB Data: %+v", *paymentData.TxData)
	/*
		verify signature
	*/
	LOGGER.Infof("Verifying the signature from OFI")
	result := xmldsig.VerifySignature(string(sendPayload.Message))
	if !result {
		LOGGER.Errorf("signature verification failed")
		pacs008LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_RFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &pacs008InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs008LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, paymentStatusMsgType, msgName, originalMsgId, pacs008InstructionId, pacs008InstructionId, "", "", pacs008LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		op.SendErrMsg(msgId, pacs008InstructionId, standardType, reqMsgType, ofiId, rfiId, constant.STATUS_CODE_OFI_SIGNATURE_FAIL)
		return
	}
	LOGGER.Infof("OFI signature verified!")

	/*
		constructing protobuffer to go struct
	*/
	pacs008 := &message_converter.Pacs008{SendPayload: sendPayload}
	xmlData, err := pacs008.ProtobuftoStruct()
	if xmlData == nil {
		LOGGER.Errorf("Encounter error while construncting proto buffer to go struct")
		op.SendErrMsg(msgId, pacs008InstructionId, standardType, reqMsgType, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR_PARSE)
		pacs008LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_RFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &pacs008InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs008LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, paymentStatusMsgType, msgName, originalMsgId, pacs008InstructionId, pacs008InstructionId, "", "", pacs008LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		return
	} else if err != nil {
		LOGGER.Errorf("Parse request from kafka failed: %s", err.Error())
		op.SendErrMsg(msgId, pacs008InstructionId, standardType, reqMsgType, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR_PARSE)
		pacs008LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_RFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &pacs008InstructionId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs008LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, paymentStatusMsgType, msgName, originalMsgId, pacs008InstructionId, pacs008InstructionId, "", "", pacs008LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		return
	}

	// Get important data from the XML data
	reqData := xmlData.RequestXMLMsg
	settlementAccountName := xmlData.OFISettlementAccountName
	instructionId := xmlData.InstructionId

	LOGGER.Infof("Finished paring ProtoBuffer to Go struct")

	// Generate payment status data
	// Aggregate necessary data for transaction memo
	feeAmount, _ := strconv.ParseFloat(pacs008.Message.Body.CdtTrfTxInf[0].ChrgsInf[0].Amt.Value, 64)
	payoutAmount, _ := strconv.ParseFloat(pacs008.Message.Body.CdtTrfTxInf[0].InstdAmt.Value, 64)
	settlementAmount, _ := strconv.ParseFloat(pacs008.Message.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Value, 64)
	exchangeRate, _ := strconv.ParseFloat(string(*pacs008.Message.Body.CdtTrfTxInf[0].XchgRate), 64)

	statusData := &sendmodel.StatusData{
		CityCdtr:             string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.TwnNm),
		CountryCdtr:          string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.Ctry),
		NameCdtr:             string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.Nm),
		IdCdtr:               string(*pacs008.Message.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id),
		CityDbtr:             string(*pacs008.Message.Body.CdtTrfTxInf[0].Dbtr.PstlAdr.TwnNm),
		CountryDbtr:          string(*pacs008.Message.Body.CdtTrfTxInf[0].Dbtr.PstlAdr.Ctry),
		NameDbtr:             string(*pacs008.Message.Body.CdtTrfTxInf[0].Dbtr.Nm),
		IdDbtr:               string(*pacs008.Message.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id),
		CurrencyCode:         pacs008.Message.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency,
		AssetType:            string(*pacs008.Message.Body.GrpHdr.SttlmInf.SttlmMtd),
		FeeCost:              feeAmount,
		FeeCurrencyCode:      pacs008.Message.Body.CdtTrfTxInf[0].ChrgsInf[0].Amt.Currency,
		FeeAssetType:         string(*pacs008.Message.Body.GrpHdr.SttlmInf.SttlmMtd),
		CreditorStreet:       string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.StrtNm),
		CreditorBuildingNo:   string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.BldgNb),
		CreditorPostalCode:   string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.PstCd),
		AmountBeneficiary:    payoutAmount,
		AssetCodeBeneficiary: pacs008.Message.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency[:3],
		CrtyCcy:              string(*pacs008.Message.Body.GrpHdr.InstdAgt.FinInstnId.BICFI)[:3],
		CustomerStreet:       string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.StrtNm),
		CustomerBuildingNo:   string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.BldgNb),
		CustomerCountry:      string(*pacs008.Message.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.PstCd),
		AccountNameSend:      string(*pacs008.Message.Body.GrpHdr.SttlmInf.SttlmAcct.Nm),
		EndToEndID:           string(*pacs008.Message.Body.CdtTrfTxInf[0].PmtId.EndToEndId),
		InstructionID:        string(*pacs008.Message.Body.CdtTrfTxInf[0].PmtId.InstrId),
		AmountSettlement:     settlementAmount,
		IssuerID:             string(*pacs008.Message.Body.GrpHdr.PmtTpInf.SvcLvl.Prtry),
		ExchangeRate:         exchangeRate,
	}

	rfiVerifyRequestAndSendToKafka(topicName, msgId, msgName, msgId, ofiId, settlementAccountName, standardType, msgName, instructionId, instructionId, paymentStatusMsgType, pacs008LogHandler, reqData, statusData, *paymentData.TxData, op)
	return
}

// ibwf.001 message handler at OFI side
func OFI_Ibwf001(data ibwf001pbstruct.SendPayload, op *kafka.KafkaOpreations) {
	// Parse the ProtoBuffer to Go struct and reconstruct it into ibwf001 message
	msgName := constant.IBWF001
	standardType := constant.ISO20022
	// Initialize the payment status
	msgType := constant.PAYMENT_TYPE_CREDIT_TRANSFER
	pacs008LogHandler := transaction.InitiatePaymentLogOperation()
	ibwf001LogHandler := transaction.InitiatePaymentLogOperation()
	originalInstrId := data.OriginalInstructionId
	instrId := data.InstructionId
	ofiId := data.OfiId
	rfiId := data.RfiId
	originalReqMsgId := data.OriginalMsgId
	msgId := data.MsgId

	originalGrpInf := &pacs002struct.OriginalGroupInformation29{
		OrgnlMsgId:   getReportMax35Text(msgId),
		OrgnlMsgNmId: getReportMax35Text(msgName),
	}

	/*
	 Get pacs008 from database
	*/
	LOGGER.Infof("Get transaction related information from database")
	paymentData, dbErr := op.DbClient.GetTx(originalInstrId)
	if dbErr != nil {
		LOGGER.Errorf("Encounter error while lookup tx data in DB: %v", dbErr)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	if paymentData.TxData == nil || paymentData.TxDetail == nil {
		LOGGER.Error("Can not get original pacs008 message from database")
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_WRONG_ORIGINAL_ID)
		return
	}

	var pacs008Data sendmodel.DBData
	pbDBData, err := base64.StdEncoding.DecodeString(*paymentData.TxData)
	if err != nil {
		LOGGER.Error(err.Error())
		return
	}
	json.Unmarshal(pbDBData, &pacs008Data)
	/*
	 Get ibwf001 from database
	*/
	ibwfPaymentData, dbErr := op.DbClient.GetTx(instrId)
	if dbErr != nil {
		LOGGER.Errorf("Encounter error while lookup tx data in DB: %v", dbErr)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	if ibwfPaymentData.TxData == nil || ibwfPaymentData.TxDetail == nil {
		LOGGER.Error("Can not get original ibwf001 message from database")
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_WRONG_INSTRUCTION_ID)
		return
	}

	pacs008LogHandler.PaymentStatuses = paymentData.TxDetail.PaymentStatuses
	ibwf001LogHandler.PaymentStatuses = ibwfPaymentData.TxDetail.PaymentStatuses
	ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_OFI_PROCESSING)

	/*
		verify signature
	*/
	LOGGER.Infof("Verifying the signature from RFI")
	result := xmldsig.VerifySignature(string(data.Message))
	if !result {
		LOGGER.Errorf("signature verification failed")
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_RFI_SIGNATURE_FAIL)
		return
	}
	LOGGER.Infof("RFI signature verified!")

	LOGGER.Infof("Parsing ProtoBuffer to Go Struct")

	var ibwf001 message_converter.MessageInterface = &message_converter.Ibwf001{SendPayload: data}
	xmlData, err := ibwf001.ProtobuftoStruct()
	if xmlData == nil {
		LOGGER.Errorf("Encounter error while construncting proto buffer to go struct")
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR_PARSE)
		return
	}

	// Get important data from the XML data
	status := xmlData.StatusCode
	statusCode, _ := strconv.Atoi(status)
	rfiAccount := xmlData.RFIAccount
	rfiAccountName := xmlData.RFISettlementAccountName
	ofiBIC := xmlData.OFIBIC
	homeDomain := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)

	// Aggregate necessary data for transaction memo
	commonStatusData := &sendmodel.StatusData{
		IdCdtr:                xmlData.RFIId,
		IdDbtr:                xmlData.OFIId,
		EndToEndID:            xmlData.OriginalEndtoEndId,
		InstructionID:         xmlData.InstructionId,
		OriginalInstructionID: xmlData.OriginalMsgId,
	}

	target := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	// report federation and compliance rejection
	if statusCode == constant.STATUS_CODE_FED_COMP_RJCT {
		report := op.ResponseHandler.CreatePacs002(ofiBIC, xmlData.InstructionId, target, statusCode, originalGrpInf)
		// Send to rejection result to OFI topic
		op.SendRequestToKafka(ofiId+"_"+kafka.TRANSACTION_TOPIC, report)
		pacs008LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FED_COMP_REJECT)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FED_COMP_REJECT)

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_DONE,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &originalInstrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_DONE,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs008LogHandler,
		})

		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, constant.PACS008, originalReqMsgId, originalInstrId, originalInstrId, "", "", pacs008LogHandler, &op.FundHandler, commonStatusData)
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, commonStatusData)
		return
	} else if err != nil || statusCode != constant.STATUS_CODE_DEFAULT {
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, commonStatusData)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, statusCode)
		return
	}

	LOGGER.Infof("Finished paring ProtoBuffer to XML")

	/*
		blocklist check
	*/

	// Create admin-service client for connecting to admin-service
	blockListClient := blocklist_client.Client{
		HTTPClient: &http.Client{Timeout: time.Second * 80},
		AdminUrl:   os.Getenv(global_environment.ENV_KEY_ADMIN_SVC_URL),
	}

	// preparing the data that need to be verify against the block-list
	var countries []string
	countries = append(countries, strconv.Quote(xmlData.OfiCountry))
	countries = append(countries, strconv.Quote(xmlData.RfiCountry))

	var currencies []string
	currencies = append(currencies, strconv.Quote(xmlData.CurrencyCode))

	var participants []string
	participants = append(participants, strconv.Quote(xmlData.RFIId))
	participants = append(participants, strconv.Quote(xmlData.OFIId))

	// validate block-list
	res, err := blockListClient.ValidateFromBlocklist(countries, currencies, participants)
	if err != nil {
		LOGGER.Errorf("%v", err)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, commonStatusData)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	if res == common.BlocklistDeniedString {
		LOGGER.Errorf("The transaction currency/country/institution is within the blocklist, transaction forbidden!")
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, commonStatusData)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_BLOCKLIST)
		return
	}

	/*
		Check if RFI was whitelisted by OFI and vice versa, if not, reject the payment request
	*/
	whitelistHandler := whitelist_handler.CreateWhiteListServiceOperations()
	LOGGER.Infof("Check whether RFI is in OFI's whitelist and vice versa.")
	pKey, whiteListErr := whitelistHandler.CheckWhiteListParticipant(homeDomain, rfiId, constant.EMPTY_STRING)
	if whiteListErr != nil {
		LOGGER.Errorf(whiteListErr.Error())
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, commonStatusData)
		return
	}
	if pKey == "" {
		LOGGER.Errorf("Can not find RFI or OFI in whitelist and vice versa")
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_RFI_OR_OFI_NOT_IN_WL)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", ibwf001LogHandler, &op.FundHandler, commonStatusData)
		return
	}
	LOGGER.Infof("Yes, RFI is in OFI's whitelist and vice versa.")

	// Initialize the data needed for signing transaction
	signData := &sendmodel.SignData{
		SettlementAccountName: pacs008Data.SettlementAccountName,
		SettlementAmount:      pacs008Data.SettlementAmount,
		CurrencyCode:          pacs008Data.SettlementCurrency,
		AssetIssuerId:         pacs008Data.AssetIssuer,
	}

	ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_OFI_VALIDATION_SUCCESS)

	// Prepare and submit the transaction to Stellar
	// Retrieve IBM account and sequence number from the gas service
	// Find account address of OFI base on the settlement information in ibwf001 message
	// Create the transaction
	// Submit the transaction to Stellar
	submitResult, txHash, _ := op.FundHandler.FundAndSubmitPaymentTransaction(rfiAccount, instrId, msgName, rfiAccountName, *signData, xdr.Memo{})
	report := op.ResponseHandler.CreatePacs002(ofiBIC, instrId, target, submitResult, originalGrpInf)

	if submitResult != constant.STATUS_CODE_TX_SEND_TO_STELLAR {
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, submitResult)
		// record the payment status "SUBMIT_FAIL"
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		pacs008LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &originalInstrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs008LogHandler,
		})

		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", rfiAccount, ibwf001LogHandler, &op.FundHandler, commonStatusData)
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, constant.PACS008, originalReqMsgId, originalInstrId, originalInstrId, "", rfiAccount, pacs008LogHandler, &op.FundHandler, commonStatusData)
		return
	} else {
		if utils.StringsEqual(pacs008Data.SettlementMethod, constant.DO_SETTLEMENT) {
			// record the payment status "CLEARED"
			ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_CLEARED, txHash)
			pacs008LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_CLEARED, txHash)
		} else {
			// If settlement method is not "DO", mark the payment status as settled
			// record the payment status "SETTLED"
			ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_SETTLED, txHash)
			pacs008LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_SETTLED, txHash)
		}
	}

	/*
		sending message to Kafka
	*/

	err = op.SendRequestToKafka(ofiId+"_"+kafka.TRANSACTION_TOPIC, report)
	if err != nil {
		LOGGER.Errorf("Encounter error while producing message to Kafka topic: %v", ofiId+"_"+kafka.TRANSACTION_TOPIC)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant.STATUS_CODE_INTERNAL_ERROR)
		ibwf001LogHandler.RecordPaymentStatus(constant.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant.DATABASE_STATUS_FAILED,
			ResId:         &constant.DATABASE_STATUS_FAILED,
			TxDetail:      &ibwf001LogHandler,
		})
		go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, txHash, "", ibwf001LogHandler, &op.FundHandler, commonStatusData)
		return
	}

	// Update transaction related information inside the DB base on message ID
	// (request ID, transaction hash, done, response ID, done)
	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &instrId,
		TxData:        &txHash,
		TxStatus:      &constant.TX_DONE,
		ResId:         &xmlData.MessageId,
		TxDetail:      &ibwf001LogHandler,
	})

	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &originalInstrId,
		TxData:        &txHash,
		TxStatus:      &constant.DATABASE_STATUS_CLEARED,
		ResId:         &xmlData.MessageId,
		TxDetail:      &pacs008LogHandler,
	})

	// Store the transaction information into the administration service and FireBase
	go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, txHash, rfiAccount, ibwf001LogHandler, &op.FundHandler, commonStatusData)
	go database.SyncWithPortalDB(constant.LOG_UPDATE, msgType, constant.PACS008, originalReqMsgId, originalInstrId, originalInstrId, txHash, rfiAccount, pacs008LogHandler, &op.FundHandler, commonStatusData)

	LOGGER.Debug("---------------------------------------------------------------------")
	return
}

// retrieving necessary data from ibwf.001
func getCriticalInfoFromIBWF001(document *ibwf001struct.Message, prServiceURL, homeDomain string, dbClient *DB.PostgreDatabaseClient) (*sendmodel.XMLData, *sendmodel.StatusData, error) {
	fedEndtoEndId := string(document.Body.ResBody[0].FedRes.PmtId.EndToEndId)
	fedInstructionId := string(document.Body.ResBody[0].FedRes.PmtId.InstrId)
	compEndtoEndId := string(document.Body.ResBody[0].CmpRes.PmtId.EndToEndId)
	compInstructionId := string(document.Body.ResBody[0].CmpRes.PmtId.InstrId)
	settlementMethod := string(document.Body.GrpHdr.SttlmInf.SttlmMtd)
	msgId := string(document.Body.GrpHdr.MsgId)
	instrId := string(document.Body.ResBody[0].Id)
	instructedAgent := string(*document.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id)
	instructingAgent := string(*document.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id)
	rfiSettlementAccountName := strings.ToLower(string(document.Body.GrpHdr.SttlmInf.SttlmAcct.Nm))
	publicKey := string(document.Body.ResBody[0].FedRes.AccId)
	fedSts := string(document.Body.ResBody[0].FedRes.FedSts)
	compInfSts := string(document.Body.ResBody[0].CmpRes.InfSts)
	compTxnSts := string(document.Body.ResBody[0].CmpRes.TxnSts)

	checkData := &sendmodel.XMLData{
		OriginalMsgId:            fedInstructionId,
		MessageId:                msgId,
		OFIId:                    instructedAgent,
		RFIId:                    instructingAgent,
		RFISettlementAccountName: rfiSettlementAccountName,
		ErrorCode:                constant.STATUS_CODE_DEFAULT,
		OriginalInstructionId:    fedInstructionId,
		InstructionId:            instrId,
	}

	if fedInstructionId == "" || fedEndtoEndId == "" || compEndtoEndId == "" || compInstructionId == "" {
		LOGGER.Error("End to end ID or intrusction ID is empty")
		checkData.ErrorCode = constant.STATUS_CODE_WRONG_ORIGINAL_ID
		return checkData, &sendmodel.StatusData{}, errors.New("End to end ID or intrusction ID is empty")
	}

	if !utils.StringsEqual(instructingAgent, homeDomain) {
		LOGGER.Error("Instructing agent is an incorrect participant")
		checkData.ErrorCode = constant.STATUS_CODE_WRONG_FI
		return checkData, &sendmodel.StatusData{}, errors.New("instructing agent is an incorrect participant")
	}

	if (!utils.StringsEqual(fedSts, constant.PAYMENT_STATUS_ACTC) && !utils.StringsEqual(fedSts, constant.PAYMENT_STATUS_RJCT)) ||
		(!utils.StringsEqual(compInfSts, constant.PAYMENT_STATUS_ACTC) && !utils.StringsEqual(compInfSts, constant.PAYMENT_STATUS_RJCT)) ||
		(!utils.StringsEqual(compTxnSts, constant.PAYMENT_STATUS_ACTC) && !utils.StringsEqual(compTxnSts, constant.PAYMENT_STATUS_RJCT)) {
		LOGGER.Error("Unknown response status code")
		checkData.ErrorCode = constant.STATUS_CODE_WRONG_RESPONSE_CODE
		return checkData, &sendmodel.StatusData{}, errors.New("unknown response status code")
	}

	account := client.GetParticipantAccount(prServiceURL, homeDomain, rfiSettlementAccountName)

	if account == nil {
		LOGGER.Error("No corresponding account for participant")
		checkData.ErrorCode = constant.STATUS_CODE_ACCOUNT_NOT_EXIST
		return checkData, &sendmodel.StatusData{}, errors.New("no corresponding account for participant")
	}

	if !utils.StringsEqual(publicKey, *account) {
		LOGGER.Error("Wrong public key")
		checkData.ErrorCode = constant.STATUS_CODE_WRONG_ID
		return checkData, &sendmodel.StatusData{}, errors.New("wrong public key")
	}

	/*
		check if original pacs008 exists or not
	*/
	paymentData, dbErr := dbClient.GetTx(fedInstructionId)
	if dbErr != nil {
		LOGGER.Errorf("Encounter error while lookup tx data in DB: %v", dbErr)
		return checkData, &sendmodel.StatusData{}, errors.New("database query error")
	}
	if paymentData.TxData == nil || paymentData.TxDetail == nil {
		checkData.ErrorCode = constant.STATUS_CODE_INTERNAL_ERROR
		return checkData, &sendmodel.StatusData{}, errors.New("database query error")
	}
	if *paymentData.TxData == "" {
		checkData.ErrorCode = constant.STATUS_CODE_WRONG_ORIGINAL_ID
		return checkData, &sendmodel.StatusData{}, errors.New("wrong original instruction id")
	}

	if !utils.StringsEqual(*paymentData.ResId, constant.DATABASE_STATUS_NONE) {
		if *paymentData.ResId == constant.DATABASE_STATUS_FAILED {
			checkData.ErrorCode = constant.STATUS_CODE_REQUEST_CLOSE
			return checkData, &sendmodel.StatusData{}, errors.New("request was closed due to internal errors")
		} else {
			checkData.ErrorCode = constant.STATUS_CODE_ALREADY_REPLIED
			return checkData, &sendmodel.StatusData{}, errors.New("the original pacs.008 is already being replied")
		}
	}

	if *paymentData.TxStatus == constant.DATABASE_STATUS_DONE {
		byteData, _ := parse.DecodeBase64(*paymentData.TxData)

		var txData sendmodel.DBData
		json.Unmarshal(byteData, &txData)
		reqSettlementMethod := txData.SettlementMethod
		if !utils.StringsEqual(reqSettlementMethod, settlementMethod) {
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_SETTLEMENT_METHOD
			LOGGER.Debugf("Original settlement method: %v, the current settlement method: %v", reqSettlementMethod, settlementMethod)
			LOGGER.Debugf("original tx data: %+v", string(byteData))
			return checkData, &sendmodel.StatusData{}, errors.New("settlement method is not the same as payment request")
		}

		if utils.StringsEqual(settlementMethod, constant.DO_SETTLEMENT) {
			if !utils.StringsEqual(rfiSettlementAccountName, common.ISSUING) {
				LOGGER.Error("Account name should be \"issuing\", if settlement method is WWDO")
				checkData.ErrorCode = constant.STATUS_CODE_WRONG_ACCOUNT_NAME
				return checkData, &sendmodel.StatusData{}, errors.New("wrong account name for DO")
			}
		}
	} else {
		LOGGER.Error("Data not found in database")
		checkData.ErrorCode = constant.STATUS_CODE_ORIGINAL_REQUEST_NOT_INIT
		return checkData, &sendmodel.StatusData{}, errors.New("Data not found in database")
	}

	/*
		transaction memo for firebase
	*/

	// Aggregate necessary data for transaction memo
	statsData := &sendmodel.StatusData{
		IdCdtr:                instructingAgent,
		IdDbtr:                instructedAgent,
		InstructionID:         instrId,
		OriginalInstructionID: fedInstructionId,
	}

	return checkData, statsData, nil
}

// retrieving necessary data from pacs.008
func getCriticalInfoFromPacs008(document *pacs008struct.Message, homeDomain string) (*sendmodel.XMLData, *sendmodel.StatusData) {
	settlementMethod := string(*document.Body.GrpHdr.SttlmInf.SttlmMtd)
	accountName := strings.ToLower(string(*document.Body.GrpHdr.SttlmInf.SttlmAcct.Nm))
	currencyCode := document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency
	feeCurrencyCode := document.Body.CdtTrfTxInf[0].ChrgsInf[0].Amt.Currency
	creditorAddress := document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr
	debtorAddress := document.Body.CdtTrfTxInf[0].Dbtr.PstlAdr

	msgId := string(*document.Body.GrpHdr.MsgId)
	instructingAgent := string(*document.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id)
	instructedAgent := string(*document.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id)
	assetIssuerId := string(*document.Body.GrpHdr.PmtTpInf.SvcLvl.Prtry)
	instructionId := string(*document.Body.CdtTrfTxInf[0].PmtId.InstrId)

	/*
		validate data
	*/

	checkData := &sendmodel.XMLData{}

	checkData.OfiCountry = string(*debtorAddress.Ctry)
	checkData.RfiCountry = string(*creditorAddress.Ctry)
	checkData.CurrencyCode = document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency
	checkData.OFIId = instructingAgent
	checkData.RFIId = instructedAgent
	checkData.OFISettlementAccountName = accountName
	checkData.MessageId = msgId
	checkData.OriginalMsgId = msgId
	checkData.OriginalInstructionId = instructionId
	checkData.InstructionId = instructionId
	checkData.ErrorCode = constant.STATUS_CODE_DEFAULT

	if !utils.StringsEqual(instructingAgent, homeDomain) {
		LOGGER.Error("Instructing agent is an incorrect participant")
		checkData.ErrorCode = constant.STATUS_CODE_WRONG_FI
		return checkData, &sendmodel.StatusData{}
	}

	if utils.StringsEqual(settlementMethod, constant.DO_SETTLEMENT) {
		// check if this DO was issued by either OFI or RFI
		if !utils.StringsEqual(instructingAgent, assetIssuerId) && !utils.StringsEqual(instructedAgent, assetIssuerId) {
			LOGGER.Error("Either OFI or RFI should be the asset issuer, if settlement method is WWDO")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_ASSET_ISSUER
			return checkData, &sendmodel.StatusData{}
		}

		if utils.StringsEqual(instructingAgent, instructedAgent) {
			LOGGER.Error("Internal DO transfer is not allowed")
			checkData.ErrorCode = constant.STATUS_CODE_DO_INTERNAL_TRANSFER_ERROR
			return checkData, &sendmodel.StatusData{}
		}

		// check if the settlement account name is "issuing"
		if !utils.StringsEqual(accountName, common.ISSUING) {
			LOGGER.Error("The settlement method is WWDO, so the account name should be \"issuing\"")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_ACCOUNT_NAME
			return checkData, &sendmodel.StatusData{}
		}

		if !horizon.IsIssuer(assetIssuerId, currencyCode) {
			LOGGER.Errorf("The asset %v is not issued by the asset issuer %v", currencyCode, assetIssuerId)
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_ASSET_ISSUER
			return checkData, &sendmodel.StatusData{}
		}

		// check if settlement asset currency code is ended with "DO"
		if !strings.HasSuffix(currencyCode, constant.SETTLEMENT_METHOD_DIGITAL_OBLIGATION) {
			LOGGER.Error("Settlement method is DO, please use DO as settlement currency")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_SETTLEMENT_METHOD
			return checkData, &sendmodel.StatusData{}
		}
		// check if fee currency code is ended with "DO"
		if !strings.HasSuffix(feeCurrencyCode, constant.SETTLEMENT_METHOD_DIGITAL_OBLIGATION) {
			LOGGER.Error("Settlement method is DO, please use DO as fee currency")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_SETTLEMENT_METHOD
			return checkData, &sendmodel.StatusData{}
		}
	} else if utils.StringsEqual(settlementMethod, constant.XLM_SETTLEMENT) {
		if !strings.HasSuffix(currencyCode, constant.SETTLEMENT_METHOD_XLM) {
			LOGGER.Error("Settlement method is XLM, please use XLM as settlement currency")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_SETTLEMENT_METHOD
			return checkData, &sendmodel.StatusData{}
		}
		if !strings.HasSuffix(feeCurrencyCode, constant.SETTLEMENT_METHOD_XLM) {
			LOGGER.Error("Settlement method is XLM, please use XLM as fee currency")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_SETTLEMENT_METHOD
			return checkData, &sendmodel.StatusData{}
		}
	} else {
		if len(currencyCode) != 3 {
			LOGGER.Error("Settlement method is DA, please use correct asset code as settlement currency")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_SETTLEMENT_METHOD
			return checkData, &sendmodel.StatusData{}
		}
		if len(feeCurrencyCode) != 3 {
			LOGGER.Error("Settlement method is DA, please use correct asset code as fee currency")
			checkData.ErrorCode = constant.STATUS_CODE_WRONG_SETTLEMENT_METHOD
			return checkData, &sendmodel.StatusData{}
		}
	}

	if creditorAddress == nil || debtorAddress == nil || creditorAddress.Ctry == nil || debtorAddress.Ctry == nil {
		LOGGER.Error("The address of either OFI or RFI is empty")
		checkData.ErrorCode = constant.STATUS_CODE_EMPTY_ADDRESS
		return checkData, &sendmodel.StatusData{}
	}

	/*
		transaction memo for firebase
	*/

	feeAmount, _ := strconv.ParseFloat(document.Body.CdtTrfTxInf[0].ChrgsInf[0].Amt.Value, 64)
	payoutAmount, _ := strconv.ParseFloat(document.Body.CdtTrfTxInf[0].InstdAmt.Value, 64)
	settlementAmount, _ := strconv.ParseFloat(document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Value, 64)
	exchangeRate, _ := strconv.ParseFloat(string(*document.Body.CdtTrfTxInf[0].XchgRate), 64)

	// Aggregate necessary data for transaction memo
	statsData := &sendmodel.StatusData{
		CityCdtr:             string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.TwnNm),
		CountryCdtr:          string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.Ctry),
		NameCdtr:             string(*document.Body.CdtTrfTxInf[0].Cdtr.Nm),
		IdCdtr:               string(*document.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id),
		CityDbtr:             string(*document.Body.CdtTrfTxInf[0].Dbtr.PstlAdr.TwnNm),
		CountryDbtr:          string(*document.Body.CdtTrfTxInf[0].Dbtr.PstlAdr.Ctry),
		NameDbtr:             string(*document.Body.CdtTrfTxInf[0].Dbtr.Nm),
		IdDbtr:               string(*document.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id),
		CurrencyCode:         document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency,
		AssetType:            string(*document.Body.GrpHdr.SttlmInf.SttlmMtd),
		FeeCost:              feeAmount,
		FeeCurrencyCode:      document.Body.CdtTrfTxInf[0].ChrgsInf[0].Amt.Currency,
		FeeAssetType:         string(*document.Body.GrpHdr.SttlmInf.SttlmMtd),
		CreditorStreet:       string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.StrtNm),
		CreditorBuildingNo:   string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.BldgNb),
		CreditorPostalCode:   string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.PstCd),
		AmountBeneficiary:    payoutAmount,
		AssetCodeBeneficiary: document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency[:3],
		CrtyCcy:              string(*document.Body.GrpHdr.InstdAgt.FinInstnId.BICFI)[:3],
		CustomerStreet:       string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.StrtNm),
		CustomerBuildingNo:   string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.BldgNb),
		CustomerCountry:      string(*document.Body.CdtTrfTxInf[0].Cdtr.PstlAdr.PstCd),
		AccountNameSend:      string(*document.Body.GrpHdr.SttlmInf.SttlmAcct.Nm),
		EndToEndID:           string(*document.Body.CdtTrfTxInf[0].PmtId.EndToEndId),
		InstructionID:        string(*document.Body.CdtTrfTxInf[0].PmtId.InstrId),
		AmountSettlement:     settlementAmount,
		IssuerID:             string(*document.Body.GrpHdr.PmtTpInf.SvcLvl.Prtry),
		ExchangeRate:         exchangeRate,
	}

	return checkData, statsData
}
