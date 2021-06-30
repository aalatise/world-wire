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

// QuoteEnvelope quoteEnvelope
//
// Quote Envelope
// swagger:model QuoteEnvelope
type QuoteEnvelope struct {

	// base64 encoded [quote](??base_url??/docs/??version??/api/participant-client-api?jump=model_quote) object
	// Required: true
	Quote *string `json:"quote"`

	// base64 encoded quote object signature
	// Required: true
	Signature *string `json:"signature"`
}

// Validate validates this quote envelope
func (m *QuoteEnvelope) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateQuote(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSignature(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *QuoteEnvelope) validateQuote(formats strfmt.Registry) error {

	if err := validate.Required("quote", "body", m.Quote); err != nil {
		return err
	}

	return nil
}

func (m *QuoteEnvelope) validateSignature(formats strfmt.Registry) error {

	if err := validate.Required("signature", "body", m.Signature); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *QuoteEnvelope) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *QuoteEnvelope) UnmarshalBinary(b []byte) error {
	var res QuoteEnvelope
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
