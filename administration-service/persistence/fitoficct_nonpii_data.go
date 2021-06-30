package persistence

type FItoFICCTNonPiiData struct {
	//
	//// administrator charges information
	//AdministratorChargesInformation *ChargesInformation `json:"administrator_charges_information,omitempty" bson:"administrator_charges_information,omitempty"`
	//
	//// The amount the beneficiary should receive in beneficiary currency
	//// Required: true
	//BeneficiaryAmount *float64 `json:"beneficiary_amount" bson:"beneficiary_amount"`
	//
	//// The asset code for the beneficiary
	//// Required: true
	//BeneficiaryAssetCode *string `json:"beneficiary_asset_code" bson:"beneficiary_asset_code"`
	//
	//// The WorldWire domain name of the beneficiary institution (i.e. uk.barclays.payments.worldwire.io)
	//// Required: true
	//BeneficiaryWwDomain *string `json:"beneficiary_ww_domain" bson:"beneficiary_ww_domain"`
	//
	//// creditor charges information
	//CreditorChargesInformation *ChargesInformation `json:"creditor_charges_information,omitempty" bson:"creditor_charges_information,omitempty"`
	//
	//// the RFI address where the payment is to be sent - received during federation protocol
	//CreditorPaymentAddress string `json:"creditor_payment_address,omitempty" bson:"creditor_payment_address,omitempty"`
	//
	//// debtor charges information
	//DebtorChargesInformation *ChargesInformation `json:"debtor_charges_information,omitempty" bson:"debtor_charges_information,omitempty"`
	//
	//// The name of the operating or issuing account from which the payment is to be sent
	//SendingAccountName string `json:"sending_account_name,omitempty" bson:"sending_account_name,omitempty"`
	//
	//// the exchange rate between settlement asset and beneficiary currency. not required if currency is same
	//ExchangeRate float64 `json:"exchange_rate,omitempty" bson:"exchange_rate,omitempty"`
	//
	//// payment identification
	//// Required: true
	//PaymentIdentification *PaymentIdentification `json:"payment_identification" bson:"payment_identification"`
	//
	//// the amount to be settled using settlement asset (exclusive of fees)
	//// Required: true
	//SettlementAmount *float64 `json:"settlement_amount" bson:"settlement_amount"`
	//
	//// settlement asset
	//// Required: true
	//SettlementAsset *Asset `json:"settlement_asset" bson:"settlement_asset"`
	//
	//// Depends on the method and asset
	//// Required: true
	//SettlementDetails *string `json:"settlement_details" bson:"settlement_details"`
	//
	//// The preferred settlement method for this payment request (DA, DO, or XLM)
	//// Required: true
	//SettlementMethod *string `json:"settlement_method" bson:"settlement_method"`

	// The name of the operating or issuing account from which the payment is to be sent
	// Required: true
	AccountNameSend *string `json:"account_name_send" bson:"account_name_send"`

	// The RFI address where the payment is to be sent - received during federation protocol
	CreditorPaymentAddress string `json:"creditor_payment_address,omitempty" bson:"creditor_payment_address"`

	// Generated by originator, a unique ID for this entire use case
	// Required: true
	EndToEndID *string `json:"end_to_end_id" bson:"end_to_end_id"`

	// The exchange rate between settlement asset and beneficiary asset. not required if asset is same
	// Required: true
	// Multiple Of: 1e-07
	ExchangeRate *float64 `json:"exchange_rate" bson:"exchange_rate"`

	// Generated by originator, a unique ID for this specific request
	// Required: true
	//InstructionID *string `json:"instruction_id" bson:"instruction_id"`

	// transaction details
	// Required: true
	TransactionDetails *TransactionDetails `json:"transactiondetails" bson:"transactiondetails"`
}

type TransactionDetails struct {

	// The amount the beneficiary should receive in beneficiary currency
	// Required: true
	// Multiple Of: 1e-07
	AmountBeneficiary *float64 `json:"amount_beneficiary" bson:"amount_beneficiary"`

	// The amount of the settlement.
	// Required: true
	// Multiple Of: 1e-07
	AmountSettlement *float64 `json:"amount_settlement" bson:"amount_settlement"`

	// The asset code for the beneficiary
	// Required: true
	AssetCodeBeneficiary *string `json:"asset_code_beneficiary" bson:"asset_code_beneficiary"`

	// asset settlement
	// Required: true
	AssetSettlement *Asset `json:"assetsettlement" bson:"assetsettlement"`

	// fee creditor
	// Required: true
	FeeCreditor *Fee `json:"feecreditor" bson:"feecreditor"`

	// The ID that identifies the OFI Participant on the WorldWire network (i.e. uk.yourbankintheUK.payments.ibm.com).
	// Required: true
	// Max Length: 32
	// Min Length: 5
	// Pattern: ^[a-zA-Z0-9-]{5,32}$
	OfiID *string `json:"ofi_id" bson:"ofi_id"`

	// The ID that identifies the RFI Participant on the WorldWire network (i.e. uk.yourbankintheUK.payments.ibm.com).
	// Required: true
	// Max Length: 32
	// Min Length: 5
	// Pattern: ^[a-zA-Z0-9-]{5,32}$
	RfiID *string `json:"rfi_id" bson:"rfi_id"`

	// The preferred settlement method for this payment request (DA, DO, or XLM)
	// Required: true
	SettlementMethod *string `json:"settlement_method" bson:"settlement_method"`
}

type Fee struct {

	// The fee amount, should be a float64 number
	// Required: true
	// Multiple Of: 1e-07
	Cost *float64 `json:"cost" bson:"cost"`

	// cost asset
	// Required: true
	CostAsset *Asset `json:"costasset" bson:"costasset"`
}