package message_handler

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/IBM/world-wire/global-whitelist-service/whitelist-handler"
	constant2 "github.com/IBM/world-wire/utility/common/constant"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/world-wire/gftn-models/model"
	"github.com/IBM/world-wire/utility/common"
	DB "github.com/IBM/world-wire/utility/database"

	pr_client "github.com/IBM/world-wire/participant-registry-client/pr-client"
	"github.com/IBM/world-wire/utility/payment/environment"
	"github.com/IBM/world-wire/utility/payment/utils"
	"github.com/IBM/world-wire/utility/xmldsig"
	"github.com/golang/protobuf/proto"
	"github.com/stellar/go/xdr"

	pacs002struct "github.com/IBM/world-wire/iso20022/pacs00200109"
	pacs009struct "github.com/IBM/world-wire/iso20022/pacs00900108"
	pacs002Pbstruct "github.com/IBM/world-wire/iso20022/proto/github.ibm.com/gftn/iso20022/pacs00200109"
	pacs009Pbstruct "github.com/IBM/world-wire/iso20022/proto/github.ibm.com/gftn/iso20022/pacs00900108"
	client "github.com/IBM/world-wire/utility/payment/client"
	"github.com/IBM/world-wire/utility/payment/utils/database"
	"github.com/IBM/world-wire/utility/payment/utils/horizon"
	"github.com/IBM/world-wire/utility/payment/utils/signing"
	"github.com/IBM/world-wire/utility/payment/utils/transaction"

	blocklist_client "github.com/IBM/world-wire/administration-service/blocklist-client"
	comn "github.com/IBM/world-wire/utility/common"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	"github.com/IBM/world-wire/utility/kafka"
	message_converter "github.com/IBM/world-wire/utility/payment/message-converter"
	"github.com/IBM/world-wire/utility/payment/utils/parse"
	"github.com/IBM/world-wire/utility/payment/utils/sendmodel"
)

var letterRunes = []rune("0123456789")

// pacs.009 message handler at OFI side
func (op *PaymentOperations) Pacs009(pacs009 message_converter.Pacs009) ([]byte, error) {

	structData := pacs009.Message
	BIC := os.Getenv(environment.ENV_KEY_PARTICIPANT_BIC)
	target := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	rfiId := strings.ToLower(os.Getenv(environment.ENV_KEY_WW_ID))
	msgName := constant2.PACS009
	instrId := string(*structData.Body.CdtTrfTxInf[0].PmtId.InstrId)

	originalGrpInf := &pacs002struct.OriginalGroupInformation29{
		OrgnlMsgId:   getReportMax35Text(string(*structData.Body.GrpHdr.MsgId)),
		OrgnlMsgNmId: getReportMax35Text(msgName),
	}
	/*
		write into dynamo
	*/
	// Initialize log handler and set the payment status to `INITIAL`
	logHandler := transaction.InitiatePaymentLogOperation()
	logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_INITIAL)

	err := op.DbClient.CreateTx(&DB.PaymentData{
		InstructionID: &instrId,
		TxData:        &constant2.DATABASE_STATUS_EMPTY,
		TxStatus:      &constant2.DATABASE_STATUS_PENDING,
		ResId:         &constant2.DATABASE_STATUS_NONE,
		TxDetail:      &logHandler,
	})
	if err != nil {
		LOGGER.Errorf(err.Error())
		report := op.ResponseHandler.CreatePacs002(BIC, instrId, target, constant2.STATUS_CODE_DUP_ID, originalGrpInf)
		return report, err
	}

	/*
		payload check
	*/
	// validate content in the pacs009 message and get all the necessary data from it
	xmlData, statsData := getCriticalInfoFromPacs009(structData, op.homeDomain)
	statusCode := xmlData.ErrorCode
	if statusCode != constant2.STATUS_CODE_DEFAULT {
		LOGGER.Errorf("Something wrong with the transaction information")
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, statusCode, originalGrpInf)
		return report, errors.New("something wrong with the transaction information")
	}

	msgType := constant2.PAYMENT_TYPE_ASSET_REDEMPTION

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
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, err
	}
	if res == common.BlocklistDeniedString {
		LOGGER.Errorf("The transaction currency/country/institution is within the blocklist, transaction forbidden!")
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_BLOCKLIST, originalGrpInf)
		return report, errors.New("the transaction currency/country/institution is within the blocklist, transaction forbidden")
	}

	/*
		Check mutual whitelist
	*/
	LOGGER.Infof("Check whether RFI is in OFI's whitelist and vice versa")
	pkey, whiteListErr := op.whitelistHandler.CheckWhiteListParticipant(xmlData.OFIId, xmlData.RFIId, constant2.EMPTY_STRING)
	if whiteListErr != nil {
		LOGGER.Errorf(whiteListErr.Error())
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, whiteListErr
	}

	if pkey == "" {
		errMsg := "OFI can not find RFI in whitelist and vice versa"
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_RFI_OR_OFI_NOT_IN_WL, originalGrpInf)
		return report, nil
	}
	LOGGER.Infof("Yes, RFI is in OFI's whitelist and vice versa")

	/*
		Parse the pacs009 message with signature into ProtoBuffer
	*/

	LOGGER.Infof("Start parsing Go struct to ProtoBuffer")
	protoBufData, parseErr := proto.Marshal(&pacs009.SendPayload)
	if parseErr != nil {
		errMsg := "Parse data to ProtoBuf error: " + parseErr.Error()
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, parseErr
	}
	LOGGER.Infof("Finished parsing Go struct to ProtoBuffer")

	// Process done with OFI side, now update the payment status
	logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_SUCCESS)

	// dbData to be written into Dynamo DB
	reportData := parse.CreatePacs009DbData(pacs009.Message.Body)
	dbData, _ := json.Marshal(reportData)
	base64DBData := parse.EncodeBase64(dbData)

	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &xmlData.InstructionId,
		TxData:        &base64DBData,
		TxStatus:      &constant2.DATABASE_STATUS_PENDING,
		ResId:         &constant2.DATABASE_STATUS_NONE,
		TxDetail:      &logHandler,
	})
	database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)

	/*
		Send the ProtoBuffer to the request topic of RFI on Kafka broker
	*/
	LOGGER.Infof("Start to send request to Kafka broker")
	kafkaErr := op.KafkaActor.Produce(rfiId+kafka.REQUEST_TOPIC, protoBufData)
	if kafkaErr != nil {
		errMsg := "Error while submit message to Kafka broker: " + kafkaErr.Error()
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, kafkaErr
	}

	LOGGER.Infof("Successfully produce message to Kafka broker")
	LOGGER.Debug("-----------------------------------------------------------------")

	// Send status back to OFI
	report := parse.CreateSuccessPacs002(BIC, target, constant2.STATUS_CODE_OFI_SEND_TO_KAFKA, xmlData)

	return report, nil
}

// pacs.009 message handler at RFI side
func RFI_Pacs009(data pacs009Pbstruct.SendPayload, op *kafka.KafkaOpreations) {
	// Parse the ProtoBuffer into Go struct and reconstruct it into pacs009 message
	LOGGER.Infof("Parsing ProtoBuffer to XML")
	standardType := constant2.ISO20022

	// Message type for payment status log : credit_transfer
	paymentStatusMsgType := constant2.PAYMENT_TYPE_ASSET_REDEMPTION
	msgName := constant2.PACS009

	instructionId := data.InstructionId
	reqMsgType := data.MsgType
	rfiId := data.RfiId
	ofiId := data.OfiId
	originalMsgId := data.OriginalMsgId
	msgId := data.MsgId
	topicName := rfiId + "_" + kafka.TRANSACTION_TOPIC

	pacs009LogHandler := transaction.InitiatePaymentLogOperation()

	/*
		get pacs009 data from DB
	*/

	paymentData, dbErr := op.DbClient.GetTx(instructionId)
	if dbErr != nil {
		LOGGER.Errorf("database query error: %v", dbErr)
		op.SendErrMsg(msgId, instructionId, standardType, reqMsgType, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	if paymentData.TxData == nil || paymentData.TxDetail == nil {
		LOGGER.Errorf("The original message ID %v does not exist in DB", instructionId)
		op.SendErrMsg(msgId, instructionId, standardType, reqMsgType, ofiId, rfiId, constant2.STATUS_CODE_WRONG_ORIGINAL_ID)
		return
	}
	pacs009LogHandler.PaymentStatuses = paymentData.TxDetail.PaymentStatuses

	/*
		verify signature
	*/
	LOGGER.Infof("Verifying the signature from OFI")
	result := xmldsig.VerifySignature(string(data.Message))
	if !result {
		LOGGER.Errorf("signature verification failed")
		pacs009LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_RFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs009LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, paymentStatusMsgType, msgName, originalMsgId, instructionId, instructionId, "", "", pacs009LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		op.SendErrMsg(msgId, instructionId, standardType, reqMsgType, ofiId, rfiId, constant2.STATUS_CODE_OFI_SIGNATURE_FAIL)
		return
	}
	LOGGER.Infof("OFI signature verified!")

	/*
		constructing protobuffer to go struct
	*/
	pacs009 := &message_converter.Pacs009{SendPayload: data}
	xmlData, err := pacs009.ProtobuftoStruct()
	if xmlData == nil {
		LOGGER.Errorf("Encounter error while construncting proto buffer to go struct")
		op.SendErrMsg(msgId, instructionId, standardType, reqMsgType, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR_PARSE)
		pacs009LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_RFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs009LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, paymentStatusMsgType, msgName, originalMsgId, instructionId, instructionId, "", "", pacs009LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		return
	} else if err != nil {
		LOGGER.Errorf("Parse request from kafka failed: %s", err.Error())
		op.SendErrMsg(msgId, instructionId, standardType, reqMsgType, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR_PARSE)
		pacs009LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_RFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs009LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, paymentStatusMsgType, msgName, originalMsgId, instructionId, instructionId, "", "", pacs009LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		return
	}

	// Get important data from the XML data
	reqData := xmlData.RequestXMLMsg
	settlementAccountName := xmlData.OFISettlementAccountName
	LOGGER.Infof("Finished paring ProtoBuffer to XML")

	// Generate payment status data
	// Aggregate necessary data for transaction memo
	settlementAmount, _ := strconv.ParseFloat(pacs009.Message.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Value, 64)

	statusData := &sendmodel.StatusData{
		NameCdtr:              string(*pacs009.Message.Body.CdtTrfTxInf[0].Cdtr.FinInstnId.Nm),
		IdCdtr:                string(*pacs009.Message.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id),
		BICCdtr:               string(*pacs009.Message.Body.GrpHdr.InstdAgt.FinInstnId.BICFI),
		NameDbtr:              string(*pacs009.Message.Body.CdtTrfTxInf[0].Dbtr.FinInstnId.Nm),
		IdDbtr:                string(*pacs009.Message.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id),
		BICDbtr:               string(*pacs009.Message.Body.GrpHdr.InstgAgt.FinInstnId.BICFI),
		CurrencyCode:          pacs009.Message.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency,
		FeeCurrencyCode:       pacs009.Message.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency,
		AssetType:             string(*pacs009.Message.Body.GrpHdr.SttlmInf.SttlmMtd),
		AssetCodeBeneficiary:  pacs009.Message.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency[:3],
		CrtyCcy:               string(*pacs009.Message.Body.GrpHdr.InstdAgt.FinInstnId.BICFI)[:3],
		AccountNameSend:       string(*pacs009.Message.Body.GrpHdr.SttlmInf.SttlmAcct.Nm),
		EndToEndID:            string(*pacs009.Message.Body.CdtTrfTxInf[0].PmtId.EndToEndId),
		InstructionID:         string(*pacs009.Message.Body.CdtTrfTxInf[0].PmtId.InstrId),
		AmountSettlement:      settlementAmount,
		IssuerID:              string(*pacs009.Message.Body.GrpHdr.PmtTpInf.SvcLvl[0].Prtry),
		OriginalInstructionID: data.InstructionId,
		SupplementaryData:     xmlData.SupplementaryData,
	}

	rfiVerifyRequestAndSendToKafka(topicName, msgId, msgName, msgId, ofiId, settlementAccountName, standardType, msgName, instructionId, instructionId, paymentStatusMsgType, pacs009LogHandler, reqData, statusData, *paymentData.TxData, op)

	return
}

// pacs.002 message handler at RFI side
func (op *PaymentOperations) Pacs002(pacs002 message_converter.Pacs002, target, BIC string) ([]byte, error) {

	structData := pacs002.Message
	msgName := constant2.PACS002
	logHandler := transaction.InitiatePaymentLogOperation()
	instructionId := string(*structData.Body.TxInfAndSts[0].StsId)
	originalGrpInf := &pacs002struct.OriginalGroupInformation29{
		OrgnlMsgId:   getReportMax35Text(string(*structData.Body.GrpHdr.MsgId)),
		OrgnlMsgNmId: getReportMax35Text(msgName),
	}

	/*
		write into dynamo
	*/
	// Initialize log handler and set the payment status to `INITIAL`
	logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_INITIAL)
	err := op.DbClient.CreateTx(&DB.PaymentData{
		InstructionID: &instructionId,
		TxData:        &constant2.DATABASE_STATUS_EMPTY,
		TxStatus:      &constant2.DATABASE_STATUS_PENDING,
		ResId:         &constant2.DATABASE_STATUS_NONE,
		TxDetail:      &logHandler,
	})
	if err != nil {
		LOGGER.Errorf(err.Error())
		report := op.ResponseHandler.CreatePacs002(BIC, instructionId, target, constant2.STATUS_CODE_DUP_ID, originalGrpInf)
		return report, err
	}

	/*
		payload check
	*/
	// validate content in the pacs002 message and get all the necessary data from it
	xmlData, statsData := getCriticalInfoFromPacs002(structData, target, op.DbClient)
	statusCode := xmlData.ErrorCode
	if statusCode != constant2.STATUS_CODE_DEFAULT {
		LOGGER.Errorf("Something wrong with the transaction information")
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, statusCode, originalGrpInf)
		return report, errors.New("something wrong with the transaction information")
	}

	msgType := constant2.PAYMENT_TYPE_ASSET_REDEMPTION
	ofiId := xmlData.OFIId

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
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, err
	}
	if res == common.BlocklistDeniedString {
		LOGGER.Errorf("The transaction currency/country/institution is within the blocklist, transaction forbidden!")
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_BLOCKLIST, originalGrpInf)
		return report, errors.New("the transaction currency/country/institution is within the blocklist, transaction forbidden")
	}

	/*
		Check mutual whitelist
	*/
	LOGGER.Infof("Check whether RFI is in OFI's whitelist and vice versa")
	pkey, whiteListErr := op.whitelistHandler.CheckWhiteListParticipant(xmlData.OFIId, xmlData.RFIId, constant2.EMPTY_STRING)
	if whiteListErr != nil {
		LOGGER.Errorf(whiteListErr.Error())
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, whiteListErr
	}

	if pkey == "" {
		errMsg := "OFI can not find RFI in whitelist and vice versa"
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_RFI_OR_OFI_NOT_IN_WL, originalGrpInf)
		return report, nil
	}
	LOGGER.Infof("Yes, RFI is in OFI's whitelist and vice versa")

	/*
		Parse the pacs002 message with signature into ProtoBuffer
	*/

	LOGGER.Infof("Start parsing Go struct to ProtoBuffer")
	protoBufData, parseErr := proto.Marshal(&pacs002.SendPayload)
	if parseErr != nil {
		errMsg := "Parse data to ProtoBuf error: " + parseErr.Error()
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_FAIL)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, parseErr
	}
	LOGGER.Infof("Finished parsing Go struct to ProtoBuffer")

	// Process done with OFI side, now update the payment status
	logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_VALIDATION_SUCCESS)

	reportData := sendmodel.DBData{
		MessageId: string(*structData.Body.GrpHdr.MsgId),
	}

	// dbData to be written into Dynamo DB
	dbData, _ := json.Marshal(reportData)
	base64DBData := parse.EncodeBase64(dbData)

	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &xmlData.InstructionId,
		TxData:        &base64DBData,
		TxStatus:      &constant2.DATABASE_STATUS_PENDING,
		ResId:         &constant2.DATABASE_STATUS_NONE,
		TxDetail:      &logHandler,
	})
	database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)

	/*
		Send the ProtoBuffer to the request topic of RFI on Kafka broker
	*/
	LOGGER.Infof("Start to send request to Kafka broker")
	kafkaErr := op.KafkaActor.Produce(ofiId+kafka.RESPONSE_TOPIC, protoBufData)
	if kafkaErr != nil {
		errMsg := "Error while submit message to Kafka broker: " + kafkaErr.Error()
		LOGGER.Errorf(errMsg)
		logHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		go database.SyncWithPortalDB(constant2.LOG_INIT, msgType, msgName, xmlData.MessageId, xmlData.OriginalInstructionId, xmlData.InstructionId, "", "", logHandler, &op.fundHandler, statsData)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &xmlData.InstructionId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_NONE,
			TxDetail:      &logHandler,
		})
		report := op.ResponseHandler.CreatePacs002(BIC, xmlData.InstructionId, target, constant2.STATUS_CODE_INTERNAL_ERROR, originalGrpInf)
		return report, kafkaErr
	}

	LOGGER.Infof("Successfully produce message to Kafka broker")
	LOGGER.Debug("-----------------------------------------------------------------")

	// Send status back to OFI
	report := parse.CreateSuccessPacs002(BIC, target, constant2.STATUS_CODE_OFI_SEND_TO_KAFKA, xmlData)

	return report, nil
}

// pacs.002 message handler at OFI side
func OFI_Pacs002(data pacs002Pbstruct.SendPayload, op *kafka.KafkaOpreations) {

	msgName := constant2.PACS002
	standardType := constant2.ISO20022
	msgType := constant2.PAYMENT_TYPE_ASSET_REDEMPTION
	pacs009LogHandler := transaction.InitiatePaymentLogOperation()
	pacs002LogHandler := transaction.InitiatePaymentLogOperation()
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
		Get pacs009 from database
	*/
	LOGGER.Infof("Get transaction related information from database")

	payment009Data, dbErr := op.DbClient.GetTx(originalInstrId)
	if dbErr != nil {
		LOGGER.Errorf("database query error: %v", dbErr)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	if payment009Data.TxData == nil || payment009Data.TxDetail == nil {
		LOGGER.Error("Can not get original pacs009 message from database")
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_WRONG_ORIGINAL_ID)
		return
	}

	var pacs009Data sendmodel.DBData

	pbDBData, err := base64.StdEncoding.DecodeString(*payment009Data.TxData)
	if err != nil {
		LOGGER.Errorf("database query error: %v", err)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	json.Unmarshal(pbDBData, &pacs009Data)

	/*
		Get pacs002 from database
	*/

	paymentData, dbErr := op.DbClient.GetTx(instrId)
	if paymentData.TxData == nil || paymentData.TxDetail == nil {
		LOGGER.Error("Can not get original pacs002 message from database")
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_WRONG_INSTRUCTION_ID)
		return
	}

	pacs009LogHandler.PaymentStatuses = payment009Data.TxDetail.PaymentStatuses
	pacs002LogHandler.PaymentStatuses = paymentData.TxDetail.PaymentStatuses
	pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_PROCESSING)

	/*
		verify signature
	*/
	LOGGER.Infof("Verifying the signature from RFI")
	result := xmldsig.VerifySignature(string(data.Message))
	if !result {
		LOGGER.Errorf("signature verification failed")
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_RFI_SIGNATURE_FAIL)
		return
	}
	LOGGER.Infof("RFI signature verified!")

	LOGGER.Infof("Parsing ProtoBuffer to Go Struct")

	var pacs002 message_converter.MessageInterface = &message_converter.Pacs002{SendPayload: data}
	xmlData, err := pacs002.ProtobuftoStruct()
	if xmlData == nil {
		LOGGER.Errorf("Encounter error while construncting proto buffer to go struct")
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, &sendmodel.StatusData{})
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR_PARSE)
		return
	}

	// Aggregate necessary data for transaction memo
	commonStatusData := &sendmodel.StatusData{
		IdCdtr:                xmlData.RFIId,
		IdDbtr:                xmlData.OFIId,
		EndToEndID:            xmlData.OriginalEndtoEndId,
		InstructionID:         xmlData.InstructionId,
		OriginalInstructionID: xmlData.OriginalMsgId,
		SupplementaryData:     xmlData.SupplementaryData,
	}

	anchorFeeAmount, _ := strconv.ParseFloat(xmlData.FeeAmount, 64)
	redeemAmount, _ := strconv.ParseFloat(pacs009Data.SettlementAmount, 64)
	if redeemAmount < anchorFeeAmount {
		LOGGER.Errorf("The charged fee is greater that the redemption amount of OFI")
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR_PARSE)
		return
	}

	status := xmlData.StatusCode
	rfiAccount := xmlData.RFIAccount
	rfiAccountAddress := xmlData.RFISettlementAccountName
	ofiBIC := xmlData.OFIBIC
	homeDomain := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)

	target := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	if !utils.StringsEqual(status, constant2.PAYMENT_STATUS_ACTC) {
		report := op.ResponseHandler.CreatePacs002(ofiBIC, instrId, target, constant2.STATUS_CODE_ASSET_REDEMPTION_RJCT, originalGrpInf)
		op.SendRequestToKafka(ofiId+"_"+kafka.TRANSACTION_TOPIC, report)
		pacs009LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_ASSET_REDEMPTION_REJECT)
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_ASSET_REDEMPTION_REJECT)

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_DONE,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &originalInstrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_DONE,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs009LogHandler,
		})

		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, constant2.PACS008, originalReqMsgId, originalInstrId, originalInstrId, "", "", pacs009LogHandler, &op.FundHandler, commonStatusData)
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		return
	} else if err != nil {
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		op.SendErrMsg(msgId, instrId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR_PARSE)
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
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		return
	}
	if res == common.BlocklistDeniedString {
		LOGGER.Errorf("The transaction currency/country/institution is within the blocklist, transaction forbidden!")
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_BLOCKLIST)
		return
	}

	/*
		Check if RFI was whitelisted by OFI and vice versa, if not, reject the payment request
	*/
	whitelistHandler := whitelist_handler.CreateWhiteListServiceOperations()
	LOGGER.Infof("Check whether RFI is in OFI's whitelist and vice versa.")
	pKey, whiteListErr := whitelistHandler.CheckWhiteListParticipant(homeDomain, rfiId, constant2.EMPTY_STRING)
	if whiteListErr != nil {
		LOGGER.Errorf(whiteListErr.Error())
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		return
	}
	if pKey == "" {
		LOGGER.Errorf("Can not find RFI or OFI in whitelist and vice versa")
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_RFI_OR_OFI_NOT_IN_WL)
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_VALIDATION_FAIL)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		return
	}
	LOGGER.Infof("Yes, RFI is in OFI's whitelist and vice versa.")
	pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_OFI_VALIDATION_SUCCESS)

	/*
		prepare to submit stellar transaction
	*/

	// Initialize the data needed for signing transaction
	signData := &sendmodel.SignData{
		SettlementAccountName: pacs009Data.SettlementAccountName,
		SettlementAmount:      pacs009Data.SettlementAmount,
		CurrencyCode:          pacs009Data.SettlementCurrency,
		AssetIssuerId:         xmlData.AssetIssuer,
	}

	referenceNo, err := strconv.ParseInt(xmlData.SupplementaryData[constant2.PACS002_SUPPLEMENTARY_DATA_PAY_REFERENCE], 10, 64)
	if err != nil {
		errString :=
			"Memo reference: " + xmlData.SupplementaryData[constant2.PACS002_SUPPLEMENTARY_DATA_PAY_REFERENCE] + "is not in number format"
		LOGGER.Error(errString)
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		return
	}
	memoHash, err := xdr.NewMemo(xdr.MemoTypeMemoId, xdr.Uint64(referenceNo))
	if err != nil {
		errString :=
			"Stellar transaction creation failed with settle DA at stronghold anchor"
		LOGGER.Error(errString)
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		return
	}

	/*
		submit to stellar
	*/
	submitResult, txHash, _ := op.FundHandler.FundAndSubmitPaymentTransaction(rfiAccount, instrId, msgName, rfiAccountAddress, *signData, memoHash)
	report := parse.CreateSuccessPacs002(ofiBIC, target, submitResult, xmlData)

	if submitResult != constant2.STATUS_CODE_TX_SEND_TO_STELLAR {
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, submitResult)
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		pacs009LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})

		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &originalInstrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs009LogHandler,
		})

		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, "", rfiAccount, pacs002LogHandler, &op.FundHandler, commonStatusData)
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, constant2.PACS008, originalReqMsgId, originalInstrId, originalInstrId, "", rfiAccount, pacs009LogHandler, &op.FundHandler, commonStatusData)
		return
	} else {
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_SETTLED, txHash)
		pacs009LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_SETTLED, txHash)
	}

	/*
		sending message to Kafka
	*/

	err = op.SendRequestToKafka(ofiId+"_"+kafka.TRANSACTION_TOPIC, report)
	if err != nil {
		LOGGER.Errorf("Encounter error while producing message to Kafka topic: %v", ofiId+"_"+kafka.TRANSACTION_TOPIC)
		op.SendErrMsg(xmlData.MessageId, xmlData.InstructionId, standardType, msgName, ofiId, rfiId, constant2.STATUS_CODE_INTERNAL_ERROR)
		pacs002LogHandler.RecordPaymentStatus(constant2.PAYMENT_STATUS_FAILED)
		op.DbClient.UpdateTx(&DB.PaymentData{
			InstructionID: &instrId,
			TxData:        &constant2.DATABASE_STATUS_EMPTY,
			TxStatus:      &constant2.DATABASE_STATUS_FAILED,
			ResId:         &constant2.DATABASE_STATUS_FAILED,
			TxDetail:      &pacs002LogHandler,
		})
		go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, txHash, "", pacs002LogHandler, &op.FundHandler, commonStatusData)
		return
	}

	// Update transaction related information inside the DynamoDB base on message ID
	// (request ID, transaction hash, done, response ID, done)
	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &instrId,
		TxData:        &txHash,
		TxStatus:      &constant2.DATABASE_STATUS_DONE,
		ResId:         &xmlData.MessageId,
		TxDetail:      &pacs002LogHandler,
	})

	op.DbClient.UpdateTx(&DB.PaymentData{
		InstructionID: &originalInstrId,
		TxData:        &txHash,
		TxStatus:      &constant2.DATABASE_STATUS_SETTLED,
		ResId:         &xmlData.MessageId,
		TxDetail:      &pacs009LogHandler,
	})

	// Store the transaction information into the administration service and FireBase
	go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, msgName, originalReqMsgId, originalInstrId, instrId, txHash, rfiAccount, pacs002LogHandler, &op.FundHandler, commonStatusData)
	go database.SyncWithPortalDB(constant2.LOG_UPDATE, msgType, constant2.PACS008, originalReqMsgId, originalInstrId, originalInstrId, txHash, rfiAccount, pacs009LogHandler, &op.FundHandler, commonStatusData)

	LOGGER.Debug("---------------------------------------------------------------------")
	return

}

// retrieving necessary data from pacs.009
func getCriticalInfoFromPacs009(document *pacs009struct.Message, homeDomain string) (*sendmodel.XMLData, *sendmodel.StatusData) {
	settlementMethod := string(*document.Body.GrpHdr.SttlmInf.SttlmMtd)
	accountName := strings.ToLower(string(*document.Body.GrpHdr.SttlmInf.SttlmAcct.Nm))
	currencyCode := document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency

	msgId := string(*document.Body.GrpHdr.MsgId)
	instructingAgent := string(*document.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id)
	instructedAgent := string(*document.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id)
	assetIssuerId := string(*document.Body.GrpHdr.PmtTpInf.SvcLvl[0].Prtry)
	instructionId := string(*document.Body.CdtTrfTxInf[0].PmtId.InstrId)

	/*
		validate data
	*/

	checkData := &sendmodel.XMLData{}

	checkData.CurrencyCode = document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency
	checkData.OFIId = instructingAgent
	checkData.RFIId = instructedAgent
	checkData.OFISettlementAccountName = accountName
	checkData.MessageId = msgId
	checkData.OriginalMsgId = msgId
	checkData.OriginalInstructionId = instructionId
	checkData.InstructionId = instructionId
	checkData.ErrorCode = constant2.STATUS_CODE_DEFAULT
	checkData.AssetIssuer = assetIssuerId

	if !utils.StringsEqual(instructedAgent, assetIssuerId) {
		LOGGER.Error("RFI Id does not match asset issuer for the redemption flow")
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_FI
		return checkData, &sendmodel.StatusData{}
	}

	if !utils.StringsEqual(instructingAgent, homeDomain) {
		LOGGER.Error("Instructing agent is an incorrect participant")
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_FI
		return checkData, &sendmodel.StatusData{}
	}

	if !utils.StringsEqual(settlementMethod, constant2.DA_SETTLEMENT) {
		LOGGER.Error("Settlement method for redemption flow must be digital asset")
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_SETTLEMENT_METHOD
		return checkData, &sendmodel.StatusData{}
	}

	if len(currencyCode) != 3 {
		LOGGER.Error("Settlement method is DA, please use correct asset code as settlement currency")
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_SETTLEMENT_METHOD
		return checkData, &sendmodel.StatusData{}
	}

	/*
		check if rfi is anchor, and if he issue the asset
	*/
	prServiceUrl := os.Getenv(global_environment.ENV_KEY_PARTICIPANT_REGISTRY_URL)
	role := client.GetParticipantRole(prServiceUrl, instructedAgent)
	if role == nil {
		LOGGER.Errorf("Unable to fetch participant info of RFI %v from PR", instructedAgent)
		checkData.ErrorCode = constant2.STATUS_CODE_INTERNAL_ERROR
		return checkData, &sendmodel.StatusData{}
	}
	if !utils.StringsEqual(*role, constant2.ROLE_ANCHOR) {
		LOGGER.Errorf("The RFI %v is not a anchor", instructedAgent)
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_ASSET_ISSUER
		return checkData, &sendmodel.StatusData{}
	}

	if !utils.StringsEqual(instructedAgent, assetIssuerId) {
		LOGGER.Errorf("Asset issuer should be the instructed agent %v", instructedAgent)
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_ASSET_ISSUER
		return checkData, &sendmodel.StatusData{}
	}

	/*
		check current balance
	*/
	settlementAmount, _ := strconv.ParseFloat(document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Value, 64)
	prclient, _ := pr_client.CreateRestPRServiceClient(os.Getenv(global_environment.ENV_KEY_PARTICIPANT_REGISTRY_URL))
	issuingAccount, err := prclient.GetParticipantAccount(instructedAgent, comn.ISSUING)
	if err != nil {
		LOGGER.Errorf(err.Error())
		return checkData, &sendmodel.StatusData{}
	}
	currentBalance, err := horizon.CheckBalance(instructingAgent, currencyCode, accountName, issuingAccount)
	if err != nil {
		LOGGER.Errorf(err.Error())
		checkData.ErrorCode = constant2.STATUS_CODE_INTERNAL_ERROR
		return checkData, &sendmodel.StatusData{}
	}
	balanceFloat64, _ := strconv.ParseFloat(currentBalance, 64)
	if balanceFloat64 < settlementAmount {
		LOGGER.Errorf("Not enough balance in stellar account to make to redemption, current balance: %v", balanceFloat64)
		checkData.ErrorCode = constant2.STATUS_CODE_INSUFFICIENT_AMOUNT
		return checkData, &sendmodel.StatusData{}
	}

	/*
		transaction memo for firebase
	*/

	// Aggregate necessary data for transaction memo
	statsData := &sendmodel.StatusData{
		NameCdtr:             string(*document.Body.CdtTrfTxInf[0].Cdtr.FinInstnId.Nm),
		IdCdtr:               string(*document.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id),
		NameDbtr:             string(*document.Body.CdtTrfTxInf[0].Dbtr.FinInstnId.Nm),
		IdDbtr:               string(*document.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id),
		CurrencyCode:         document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency,
		AssetType:            string(*document.Body.GrpHdr.SttlmInf.SttlmMtd),
		FeeAssetType:         string(*document.Body.GrpHdr.SttlmInf.SttlmMtd),
		AssetCodeBeneficiary: document.Body.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency[:3],
		CrtyCcy:              string(*document.Body.GrpHdr.InstdAgt.FinInstnId.BICFI)[:3],
		AccountNameSend:      string(*document.Body.GrpHdr.SttlmInf.SttlmAcct.Nm),
		EndToEndID:           string(*document.Body.CdtTrfTxInf[0].PmtId.EndToEndId),
		InstructionID:        string(*document.Body.CdtTrfTxInf[0].PmtId.InstrId),
		AmountSettlement:     settlementAmount,
		IssuerID:             string(*document.Body.GrpHdr.PmtTpInf.SvcLvl[0].Prtry),
	}

	return checkData, statsData
}

// retrieving necessary data from pacs.002
func getCriticalInfoFromPacs002(document *pacs002struct.Message, homeDomain string, dbClient *DB.PostgreDatabaseClient) (*sendmodel.XMLData, *sendmodel.StatusData) {

	splmtryDataSet := make(map[string]string)

	for _, splmtryData := range document.Body.TxInfAndSts[0].SplmtryData {
		index := string(*splmtryData.PlcAndNm)
		if splmtryData.Envlp != nil && splmtryData.Envlp.Id != nil {
			splmtryDataSet[index] = string(*splmtryData.Envlp.Id)
		}
	}

	accountAddress := splmtryDataSet[constant2.PACS002_SUPPLEMENTARY_DATA_ACCOUNT_ADDRESS]
	currencyCode := document.Body.TxInfAndSts[0].OrgnlTxRef.IntrBkSttlmAmt.Currency

	msgId := string(*document.Body.GrpHdr.MsgId)
	instructingAgent := string(*document.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id)
	instructedAgent := string(*document.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id)

	assetIssuerId := splmtryDataSet[constant2.PACS002_SUPPLEMENTARY_DATA_ISSUER]
	instructionId := string(*document.Body.TxInfAndSts[0].StsId)
	originalInstructionId := string(*document.Body.TxInfAndSts[0].OrgnlInstrId)
	chargesTaker := string(*document.Body.TxInfAndSts[0].ChrgsInf[0].Agt.FinInstnId.Othr.Id)
	returnedRedeemAmount := document.Body.TxInfAndSts[0].OrgnlTxRef.IntrBkSttlmAmt.Value

	/*
		validate data
	*/

	checkData := &sendmodel.XMLData{}

	checkData.CurrencyCode = currencyCode
	checkData.OFIId = instructedAgent
	checkData.RFIId = instructingAgent
	checkData.RFISettlementAccountName = accountAddress
	checkData.MessageId = msgId
	checkData.OriginalInstructionId = originalInstructionId
	checkData.InstructionId = instructionId
	checkData.ErrorCode = constant2.STATUS_CODE_DEFAULT
	checkData.AssetIssuer = assetIssuerId

	if !utils.StringsEqual(instructingAgent, homeDomain) {
		LOGGER.Error("Instructing agent is an incorrect participant")
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_FI
		return checkData, &sendmodel.StatusData{}
	}

	if !utils.StringsEqual(chargesTaker, instructingAgent) {
		LOGGER.Error("Charges taker should be anchor")
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_FI
		return checkData, &sendmodel.StatusData{}
	}

	if len(currencyCode) != 3 {
		LOGGER.Error("Settlement method is DA, please use correct asset code as settlement currency")
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_SETTLEMENT_METHOD
		return checkData, &sendmodel.StatusData{}
	}

	/*
		get original pacs009 data
	*/
	// Get the status of the cancellation request data from DynamoDB
	paymentData, dbErr := dbClient.GetTx(originalInstructionId)

	// Query failed or data unmarshal failed
	if dbErr != nil {
		checkData.ErrorCode = constant2.STATUS_CODE_INTERNAL_ERROR
		return checkData, nil
	}

	// Not corresponding data exist in the database
	if paymentData.TxData == nil {
		checkData.ErrorCode = constant2.STATUS_CODE_WRONG_ORIGINAL_ID
		return checkData, nil
	}

	var pacs009Data sendmodel.DBData

	if *paymentData.TxData != constant2.EMPTY_STRING {
		pbDBData, _ := parse.DecodeBase64(*paymentData.TxData)
		json.Unmarshal(pbDBData, &pacs009Data)
	}

	if *paymentData.TxStatus != constant2.DATABASE_STATUS_ASSET_REDEMPTION_INIT &&
		!utils.StringsEqual(pacs009Data.InstructedAgentId, os.Getenv(global_environment.ENV_KEY_STRONGHOLD_ANCHOR_ID)) {
		LOGGER.Error("Data not found in database")
		checkData.ErrorCode = constant2.STATUS_CODE_ORIGINAL_REQUEST_NOT_INIT
		return checkData, nil
	}

	originalRedeemAmount, _ := strconv.ParseFloat(pacs009Data.SettlementAmount, 64)
	returnedRedeemFloatAmount, _ := strconv.ParseFloat(returnedRedeemAmount, 64)

	if originalRedeemAmount != returnedRedeemFloatAmount {
		LOGGER.Error("Redemption amount returned by anchor is different from the original request amount")
		checkData.ErrorCode = constant2.STATUS_CODE_REDEMPTION_AMOUNT_MISMATCH
		return checkData, nil
	}

	/*
		transaction memo for firebase
	*/

	settlementAmount, _ := strconv.ParseFloat(document.Body.TxInfAndSts[0].OrgnlTxRef.IntrBkSttlmAmt.Value, 64)

	// Aggregate necessary data for transaction memo
	statsData := &sendmodel.StatusData{
		IdCdtr:           string(*document.Body.GrpHdr.InstdAgt.FinInstnId.Othr.Id),
		IdDbtr:           string(*document.Body.GrpHdr.InstgAgt.FinInstnId.Othr.Id),
		CurrencyCode:     document.Body.TxInfAndSts[0].OrgnlTxRef.IntrBkSttlmAmt.Currency,
		AssetType:        constant2.DA_SETTLEMENT,
		FeeCurrencyCode:  document.Body.TxInfAndSts[0].ChrgsInf[0].Amt.Currency,
		InstructionID:    string(*document.Body.TxInfAndSts[0].StsId),
		AmountSettlement: settlementAmount,
		IssuerID:         assetIssuerId,
	}

	return checkData, statsData
}

// constructing pacs.002 message for stronghold
func constructPacsMessageForSH(originalData *sendmodel.StatusData, anchorRes *model.StrongholdWithdrawResponse, responseHanlder *signing.CreateSignOperations) ([]byte, string, error) {

	timeNow, _ := time.Parse("2006-01-02T15:04:05", time.Now().UTC().Format("2006-01-02T15:04:05"))
	t := pacs002struct.ISODateTime(timeNow.String())

	var currencyCode string

	dateToday := time.Now().Format("02-01-2006")
	dateToday = strings.Replace(dateToday, "-", "", -1)

	wwBIC := os.Getenv(environment.ENV_KEY_WW_BIC)

	bicfi := pacs002struct.BICFIIdentifier(os.Getenv(environment.ENV_KEY_WW_BIC))
	wwId := pacs002struct.Max35Text(os.Getenv(environment.ENV_KEY_WW_ID))
	ofiBicfi := pacs002struct.BICFIIdentifier(originalData.BICDbtr)
	rfiBicfi := pacs002struct.BICFIIdentifier(originalData.BICCdtr)
	credt := pacs002struct.ISONormalisedDateTime(timeNow.String())
	var anchorStatus pacs002struct.ExternalPaymentTransactionStatus1Code
	if *anchorRes.Success == true {
		anchorStatus = pacs002struct.ExternalPaymentTransactionStatus1Code(constant2.PAYMENT_STATUS_ACTC)
	} else {
		anchorStatus = pacs002struct.ExternalPaymentTransactionStatus1Code(constant2.PAYMENT_STATUS_RJCT)
	}

	b := make([]rune, 11)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	seqNum := string(b)
	instructionId := currencyCode + dateToday + wwBIC + seqNum

	report := &pacs002struct.Message{
		Body: &pacs002struct.FIToFIPaymentStatusReportV09{
			GrpHdr: &pacs002struct.GroupHeader53{
				MsgId:   getReportMax35Text(instructionId),
				CreDtTm: &t,
				InstgAgt: &pacs002struct.BranchAndFinancialInstitutionIdentification5{
					FinInstnId: &pacs002struct.FinancialInstitutionIdentification8{
						BICFI: &rfiBicfi,
						Othr: &pacs002struct.GenericFinancialIdentification1{
							Id: getReportMax35Text(originalData.IdCdtr),
						},
					},
				},
				InstdAgt: &pacs002struct.BranchAndFinancialInstitutionIdentification5{
					FinInstnId: &pacs002struct.FinancialInstitutionIdentification8{
						BICFI: &ofiBicfi,
						Othr: &pacs002struct.GenericFinancialIdentification1{
							Id: getReportMax35Text(originalData.IdDbtr),
						},
					},
				},
			},
			TxInfAndSts: []*pacs002struct.PaymentTransaction91{
				&pacs002struct.PaymentTransaction91{
					StsId:        getReportMax35Text(instructionId),
					OrgnlInstrId: getReportMax35Text(originalData.OriginalInstructionID),
					TxSts:        &anchorStatus,
					ChrgsInf: []*pacs002struct.Charges2{
						&pacs002struct.Charges2{
							Amt: &pacs002struct.ActiveOrHistoricCurrencyAndAmount{
								Currency: originalData.FeeCurrencyCode,
								Value:    anchorRes.Result.FeeAmount,
							},
							Agt: &pacs002struct.BranchAndFinancialInstitutionIdentification5{
								FinInstnId: &pacs002struct.FinancialInstitutionIdentification8{
									BICFI: &rfiBicfi,
									Othr: &pacs002struct.GenericFinancialIdentification1{
										Id: getReportMax35Text(originalData.IdCdtr),
									},
								},
							},
						},
					},
					OrgnlTxRef: &pacs002struct.OriginalTransactionReference27{
						IntrBkSttlmAmt: &pacs002struct.ActiveOrHistoricCurrencyAndAmount{
							Value:    anchorRes.Result.PaymentMethodInstructions.Amount,
							Currency: originalData.CurrencyCode,
						},
					},
					SplmtryData: []*pacs002struct.SupplementaryData1{
						&pacs002struct.SupplementaryData1{
							PlcAndNm: getReportMax350Text(constant2.PACS002_SUPPLEMENTARY_DATA_PAY_REFERENCE),
							Envlp: &pacs002struct.SupplementaryDataEnvelope1{
								Id: getReportMax34Text(anchorRes.Result.PaymentMethodInstructions.PayToReference),
							},
						},
						&pacs002struct.SupplementaryData1{
							PlcAndNm: getReportMax350Text(constant2.PACS002_SUPPLEMENTARY_DATA_ACCOUNT_ADDRESS),
							Envlp: &pacs002struct.SupplementaryDataEnvelope1{
								Id: getReportMax34Text(anchorRes.Result.PaymentMethodInstructions.PayToVenueSpecific),
							},
						},
						&pacs002struct.SupplementaryData1{
							PlcAndNm: getReportMax350Text(constant2.PACS002_SUPPLEMENTARY_DATA_ISSUER),
							Envlp: &pacs002struct.SupplementaryDataEnvelope1{
								Id: getReportMax34Text(originalData.IssuerID),
							},
						},
					},
				},
			},
		},
		Head: &pacs002struct.BusinessApplicationHeaderV01{
			Fr: &pacs002struct.Party9Choice{
				FIId: &pacs002struct.BranchAndFinancialInstitutionIdentification5{
					FinInstnId: &pacs002struct.FinancialInstitutionIdentification8{
						BICFI: &rfiBicfi,
						Othr: &pacs002struct.GenericFinancialIdentification1{
							Id: getReportMax35Text(originalData.IdCdtr),
						},
					},
				},
			},
			To: &pacs002struct.Party9Choice{
				FIId: &pacs002struct.BranchAndFinancialInstitutionIdentification5{
					FinInstnId: &pacs002struct.FinancialInstitutionIdentification8{
						BICFI: &bicfi,
						Othr: &pacs002struct.GenericFinancialIdentification1{
							Id: &wwId,
						},
					},
				},
			},
			BizMsgIdr: getReportMax35Text(currencyCode + dateToday + wwBIC + seqNum),
			MsgDefIdr: getReportMax35Text(constant2.PACS002),
			CreDt:     &credt,
		},
	}

	msg, _ := xml.MarshalIndent(report, "", "\t")

	header := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	cbMsg := []byte(header + string(msg))

	/*
		Signing message with IBM master account
	*/
	var signedMessage []byte
	var signErr error
	LOGGER.Infof("Signing with utility function")
	var signedRes string
	signedRes, signErr = responseHanlder.SignPayloadByMasterAccountLocally(string(cbMsg))
	signedMessage = []byte(signedRes)
	if signErr != nil {
		LOGGER.Errorf("Failed to sign payload: %v", signErr.Error())
		return nil, "", nil
	}
	/*
		gatewayMsg := parse.EncodeBase64(signedMessage)
		callBackMsg := &model.SendPacs{
			MessageType: &statusMsgType,
			Message:     &gatewayMsg,
		}
		res, _ := json.Marshal(callBackMsg)
	*/
	return signedMessage, instructionId, nil
}

func getReportMax35Text(text string) *pacs002struct.Max35Text {
	res := pacs002struct.Max35Text(text)
	return &res
}

func getReportMax350Text(text string) *pacs002struct.Max350Text {
	res := pacs002struct.Max350Text(text)
	return &res
}

func getReportMax34Text(text string) *pacs002struct.Max34Text {
	res := pacs002struct.Max34Text(text)
	return &res
}
