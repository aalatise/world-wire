package payment

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/IBM/world-wire/gftn-models/model"
	"github.com/IBM/world-wire/utility/payment/constant"
)

func BuildFiToFiCCTTxnMemo(opType string, tr model.Send, stellarTxnID, orgnlMsgId, orgnlInstrId, messageType, messageName string) model.FitoFICCTMemoData {

	var piiData model.FItoFICCTPiiData
	var txnMemo model.FitoFICCTMemoData

	txnMemo.Fitoficctnonpiidata = &model.FitoFICCTNonPiiData{}
	// build non-pii-data
	txnMemo.Fitoficctnonpiidata.OriginalMessageID = &orgnlMsgId
	txnMemo.Fitoficctnonpiidata.OriginalInstructionID = &orgnlInstrId
	txnMemo.Fitoficctnonpiidata.EndToEndID = &tr.EndToEndID
	txnMemo.Fitoficctnonpiidata.ExchangeRate = &tr.ExchangeRate
	txnMemo.Fitoficctnonpiidata.AccountNameSend = &tr.AccountNameSend
	txnMemo.Fitoficctnonpiidata.InstructionID = &tr.InstructionID
	txnMemo.Fitoficctnonpiidata.Transactiondetails = tr.TransactionDetails

	// build pii data
	if opType == constant.LOG_INIT {
		piiData.DebtorInformation = tr.Debtor
		piiData.CreditorInformation = tr.Creditor
		piiBytes, _ := json.Marshal(piiData)
		piiHash := sha256.Sum256(piiBytes)
		piiHashString := fmt.Sprintf("%x", piiHash)
		txnMemo.FitoficctPiiHash = &piiHashString
	}

	// build txn memo
	txnMemo.MessageName = &messageName

	if len(stellarTxnID) > 0 {
		txnMemo.TransactionIdentifier = append(txnMemo.TransactionIdentifier, stellarTxnID)
	}

	txnMemo.MessageType = &messageType

	return txnMemo
}
