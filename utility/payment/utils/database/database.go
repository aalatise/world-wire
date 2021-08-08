package database

import (
	"encoding/base64"
	"encoding/json"
	"github.com/IBM/world-wire/utility/common/constant"

	"github.com/IBM/world-wire/gftn-models/model"
	"github.com/IBM/world-wire/utility/payment/utils/sendmodel"
	"github.com/IBM/world-wire/utility/payment/utils/transaction"
)

func SyncWithPortalDB(
	operationType, msgType, msgName, originalMsgId, orginalInstructionId, instructionId, txHash, creditorPaymentAddress string,
	logHandler transaction.Payment,
	fundHandler *transaction.CreateFundingOpereations,
	firebaseData *sendmodel.StatusData) {
	LOGGER.Infof("Synchronizing message to firebase with message type: %v, instruction ID: %v", msgType, instructionId)
	var txMemo model.FitoFICCTMemoData
	if firebaseData != nil {
		txMemo = logHandler.BuildTXMemo(operationType, firebaseData, txHash, originalMsgId, orginalInstructionId, firebaseData.IdDbtr, msgType, msgName)
	}
	if creditorPaymentAddress != "" {
		txMemo.Fitoficctnonpiidata.CreditorPaymentAddress = creditorPaymentAddress
	}

	// use instruction ID as index in firebase
	fundHandler.SendToAdm(logHandler.PaymentStatuses, operationType, instructionId, txMemo)
	LOGGER.Infof("Synchronizing complete!")
	return
}

func SyncWithDynamo(opType, primaryIndex, txData, txStatus, resId string, logHandler transaction.Payment) error {
	LOGGER.Infof("Synchronizing message to dynamo DB with ID: %v, status ID: %v", primaryIndex, txStatus)
	paymentData, _ := json.Marshal(logHandler.PaymentStatuses)
	base64PaymentData := base64.StdEncoding.EncodeToString(paymentData)
	if opType == constant.DATABASE_UPDATE {
		//20200301 - Mayuran - Start Update should happen in the same go routine
		DC.UpdateTransactionData(primaryIndex, txData, txStatus, resId, base64PaymentData)
		//20200301 - Mayuran - End
	} else if opType == constant.DATABASE_INIT {
		//initialize should not go parallel
		return DC.AddTransactionData(primaryIndex, txData, txStatus, resId, base64PaymentData)
	}
	return nil
}
