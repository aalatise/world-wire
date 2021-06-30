package persistence

type ChargesInformation struct {

	// amount of fee charged
	// Required: true
	Amount *float64 `json:"amount" bson:"amount"`

	// asset
	// Required: true
	Asset *Asset `json:"asset" bson:"asset"`
}
