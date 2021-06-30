package persistence

type PaymentIdentification struct {

	// Generated by originator, a unique ID for this entire use case
	EndToEndID string `json:"end_to_end_id,omitempty" bson:"end_to_end_id,omitempty"`

	// Generated by originator, a unique ID for this specific request
	InstructionID string `json:"instruction_id,omitempty" bson:"instruction_id,omitempty"`
}
