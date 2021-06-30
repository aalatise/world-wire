// Code generated by go-swagger; DO NOT EDIT.

package model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Send send
//
// Send
// swagger:model Send
type Send struct {

	// The name of the operating or issuing account from which the payment is to be sent
	AccountNameSend string `json:"account_name_send,omitempty"`

	// creditor
	// Required: true
	Creditor *PaymentActor `json:"creditor"`

	// debtor
	// Required: true
	Debtor *PaymentActor `json:"debtor"`

	// Generated by originator, a unique ID for this entire use case
	EndToEndID string `json:"end_to_end_id,omitempty"`

	// The exchange rate between settlement asset and beneficiary asset. not required if asset is same
	// Multiple Of: 1e-07
	ExchangeRate float64 `json:"exchange_rate,omitempty"`

	// Generated by originator, a unique ID for this specific request
	InstructionID string `json:"instruction_id,omitempty"`

	// transaction details
	// Required: true
	TransactionDetails *TransactionDetails `json:"transaction_details"`
}

// Validate validates this send
func (m *Send) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCreditor(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDebtor(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateExchangeRate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTransactionDetails(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Send) validateCreditor(formats strfmt.Registry) error {

	if err := validate.Required("creditor", "body", m.Creditor); err != nil {
		return err
	}

	if m.Creditor != nil {
		if err := m.Creditor.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("creditor")
			}
			return err
		}
	}

	return nil
}

func (m *Send) validateDebtor(formats strfmt.Registry) error {

	if err := validate.Required("debtor", "body", m.Debtor); err != nil {
		return err
	}

	if m.Debtor != nil {
		if err := m.Debtor.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("debtor")
			}
			return err
		}
	}

	return nil
}

func (m *Send) validateExchangeRate(formats strfmt.Registry) error {

	if swag.IsZero(m.ExchangeRate) { // not required
		return nil
	}

	if err := validate.MultipleOf("exchange_rate", "body", float64(m.ExchangeRate), 1e-07); err != nil {
		return err
	}

	return nil
}

func (m *Send) validateTransactionDetails(formats strfmt.Registry) error {

	if err := validate.Required("transaction_details", "body", m.TransactionDetails); err != nil {
		return err
	}

	if m.TransactionDetails != nil {
		if err := m.TransactionDetails.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("transaction_details")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Send) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Send) UnmarshalBinary(b []byte) error {
	var res Send
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}