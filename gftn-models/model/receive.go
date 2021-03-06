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

// Receive receive
//
// Receive
// swagger:model Receive
type Receive struct {

	// The operating or issuing account name where the payment was received.
	//
	// Required: true
	AccountName *string `json:"account_name"`

	// The unique identifier for the compliance check done prior to this payment.
	//
	// Required: true
	TransactionID *string `json:"transaction_id"`

	// Optional info about the transaction.
	//
	TransactionMemo string `json:"transaction_memo,omitempty"`

	// Cursor location of this payment on the ledger.
	// Required: true
	TransactionReference *string `json:"transaction_reference"`
}

// Validate validates this receive
func (m *Receive) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAccountName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTransactionID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTransactionReference(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Receive) validateAccountName(formats strfmt.Registry) error {

	if err := validate.Required("account_name", "body", m.AccountName); err != nil {
		return err
	}

	return nil
}

func (m *Receive) validateTransactionID(formats strfmt.Registry) error {

	if err := validate.Required("transaction_id", "body", m.TransactionID); err != nil {
		return err
	}

	return nil
}

func (m *Receive) validateTransactionReference(formats strfmt.Registry) error {

	if err := validate.Required("transaction_reference", "body", m.TransactionReference); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Receive) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Receive) UnmarshalBinary(b []byte) error {
	var res Receive
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
