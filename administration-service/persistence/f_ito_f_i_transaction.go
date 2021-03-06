package persistence

type FItoFITransaction struct {

	// transaction details
	// Required: true
	TransactionDetails *TransactionDetails `json:"transaction_details"`

	// transaction receipt
	// Required: true
	TransactionReceipt *TransactionReceipt `json:"transaction_receipt"`
}

type TransactionReceipt struct {

	// The timestamp of the transaction.
	// Required: true
	TimeStamp *int64 `json:"time_stamp"`

	// A unique transaction identifier generated by the ledger.
	// Required: true
	TransactionID *string `json:"transaction_id"`

	// For DA (digital asset) or DO (digital obligation) ops, this will be "cleared".  For cryptocurrencies, this will be "settled".
	// Required: true
	TransactionStatus *string `json:"transaction_status"`
}
