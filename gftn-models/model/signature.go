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

// Signature signature
//
// signature
// swagger:model signature
type Signature struct {

	// Transaction signed by Participant.
	// Required: true
	// Format: byte
	TransactionSigned *strfmt.Base64 `json:"transaction_signed"`
}

// Validate validates this signature
func (m *Signature) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateTransactionSigned(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Signature) validateTransactionSigned(formats strfmt.Registry) error {

	if err := validate.Required("transaction_signed", "body", m.TransactionSigned); err != nil {
		return err
	}

	// Format "byte" (base64 string) is already validated when unmarshalled

	return nil
}

// MarshalBinary interface implementation
func (m *Signature) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Signature) UnmarshalBinary(b []byte) error {
	var res Signature
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
