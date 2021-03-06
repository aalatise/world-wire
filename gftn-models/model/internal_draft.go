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

// InternalDraft internalDraft
//
// draft
// swagger:model internalDraft
type InternalDraft struct {

	// The name of the account with which the transactions needs to be signed
	// Required: true
	AccountName *string `json:"account_name"`

	// Identifier for transaction.
	TransactionID string `json:"transaction_id,omitempty"`

	// The unsigned transaction envelope to be signed by IBM account.
	// Required: true
	// Format: byte
	TransactionUnsigned *strfmt.Base64 `json:"transaction_unsigned"`
}

// Validate validates this internal draft
func (m *InternalDraft) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAccountName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTransactionUnsigned(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *InternalDraft) validateAccountName(formats strfmt.Registry) error {

	if err := validate.Required("account_name", "body", m.AccountName); err != nil {
		return err
	}

	return nil
}

func (m *InternalDraft) validateTransactionUnsigned(formats strfmt.Registry) error {

	if err := validate.Required("transaction_unsigned", "body", m.TransactionUnsigned); err != nil {
		return err
	}

	// Format "byte" (base64 string) is already validated when unmarshalled

	return nil
}

// MarshalBinary interface implementation
func (m *InternalDraft) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *InternalDraft) UnmarshalBinary(b []byte) error {
	var res InternalDraft
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
