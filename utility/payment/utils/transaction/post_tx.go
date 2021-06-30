package transaction

import (
	"encoding/json"
	"errors"
	"time"

	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	"github.ibm.com/gftn/world-wire-services/utility/wwfirebase"
)

// FbTrxLog : Logs Non-PII transaction data for the portal
type FbTrxLog struct {
	// Don't log "TransferRequest" remove this as it may include PII
	// The only reason we might want to consider keeping
	// this info is for addresses for GEO locations
	//  and payout point for tracking endpoint financial institutions
	// TransferRequest interface{}

	// HomeDomain
	ParticipantID *string `json:"participant_id"`

	// Transaction details included in transaction memo
	// debtor/creditor amounts, currencies, status
	TransactionMemo interface{} `json:"transaction_memo"`
}

type FbTrxUpdateLog struct {
	ParticipantID   string                 `json:"participant_id"`
	TransactionMemo map[string]interface{} `json:"transaction_memo"`
}

// SendToAdm - logs transaction in admin service, firebase, etc.
func (op *CreateFundingOpereations) SendToAdm(paymentInfo []model.TransactionReceipt, method, instructionId string, txMemo model.FitoFICCTMemoData) {
	timeStamp := time.Now().Unix()
	LOGGER.Debug("Store transactionMemo to WW admin-service and FireBase DB")

	op.sendTransactionToWWAdministrator(paymentInfo, timeStamp, method, instructionId, txMemo)
}

func (op *CreateFundingOpereations) sendTransactionToWWAdministrator(paymentInfo []model.TransactionReceipt, timeStamp int64, method, instructionId string, txMemo model.FitoFICCTMemoData) {

	LOGGER.Infof("Sending transaction to admin service")

	newMemoData := txMemo
	newMemoData.TimeStamp = &timeStamp

	for _, p := range paymentInfo {
		pByte, _ := json.Marshal(p)

		var pi *model.TransactionReceipt
		json.Unmarshal(pByte, &pi)
		newMemoData.TransactionStatus = append(newMemoData.TransactionStatus, pi)
	}

	op.admClient.StoreFITOFICCTMemo(newMemoData)

}

func updateOriginalTxDetails(txMemo model.FitoFICCTMemoData, release, originalMsgName string) error {
	ofiId := *txMemo.Fitoficctnonpiidata.Transactiondetails.OfiID
	rfiId := *txMemo.Fitoficctnonpiidata.Transactiondetails.RfiID
	instructionId := *txMemo.Fitoficctnonpiidata.OriginalInstructionID
	LOGGER.Infof("Updating Firebase record %v for both OFI: %v and RFI: %v", instructionId, ofiId, rfiId)
	participantIds := []string{ofiId, rfiId}

	for _, pID := range participantIds {
		var log interface{}

		LOGGER.Infof("Update result to FireBase for participant: %s", pID)
		// Get all the txn logs from FireBase
		ref := wwfirebase.FbClient.NewRef("/" + release + "/txn/transfer/" + pID + "/" + instructionId + "/transaction_memo")
		ref.Get(wwfirebase.AppContext, &log)
		if log == nil {
			LOGGER.Error("Unable to find instruction id %s for participant: %s", instructionId, pID)
			return errors.New("Unable to find original instruction id: " + instructionId)
		}

		updatedLog := make(map[string]interface{})
		updatedMemoMap := make(map[string]interface{})

		byteData, _ := json.Marshal(log)
		var oldMemoData *model.FitoFICCTMemoData
		json.Unmarshal(byteData, &oldMemoData)

		oldMemoData.Fitoficctnonpiidata.AccountNameSend = txMemo.Fitoficctnonpiidata.AccountNameSend
		oldMemoData.Fitoficctnonpiidata.ExchangeRate = txMemo.Fitoficctnonpiidata.ExchangeRate
		oldMemoData.Fitoficctnonpiidata.Transactiondetails = txMemo.Fitoficctnonpiidata.Transactiondetails
		byteMemo, _ := json.Marshal(oldMemoData)
		json.Unmarshal(byteMemo, &updatedMemoMap)

		updatedLog["fitoficctnonpiidata"] = txMemo.Fitoficctnonpiidata

		// Update the log into FireBase
		err := ref.Update(wwfirebase.AppContext, updatedLog)
		if err != nil {
			LOGGER.Error("Error posting log to Firebase for participant %s: %s", pID, err.Error())
		}

	}
	return nil
}

func getTxDetails(txMemo model.FitoFICCTMemoData, release, originalMsgName string) (model.FitoFICCTMemoData, error) {
	ofiId := *txMemo.Fitoficctnonpiidata.Transactiondetails.OfiID
	originalInstructionId := *txMemo.Fitoficctnonpiidata.OriginalInstructionID

	var log interface{}

	LOGGER.Infof("Update result to FireBase for participant: %s", ofiId)
	ref := wwfirebase.FbClient.NewRef("/" + release + "/txn/transfer/" + ofiId + "/" + originalInstructionId + "/transaction_memo/fitoficctnonpiidata")
	ref.Get(wwfirebase.AppContext, &log)
	if log == nil {
		LOGGER.Error("Unable to find instruction id %s for participant: %s", originalInstructionId, ofiId)
		return model.FitoFICCTMemoData{}, errors.New("Unable to find instruction id: " + originalInstructionId)
	}

	e := log.(map[string]interface{})
	byteData, _ := json.Marshal(e)
	var oldMemoData *model.FitoFICCTNonPiiData
	json.Unmarshal(byteData, &oldMemoData)

	txMemo.Fitoficctnonpiidata.AccountNameSend = oldMemoData.AccountNameSend
	txMemo.Fitoficctnonpiidata.CreditorPaymentAddress = oldMemoData.CreditorPaymentAddress
	txMemo.Fitoficctnonpiidata.EndToEndID = oldMemoData.EndToEndID
	txMemo.Fitoficctnonpiidata.ExchangeRate = oldMemoData.ExchangeRate
	txMemo.Fitoficctnonpiidata.Transactiondetails = oldMemoData.Transactiondetails

	return txMemo, nil
}
