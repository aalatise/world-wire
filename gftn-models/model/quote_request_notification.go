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

// QuoteRequestNotification quoteRequestNotification
//
// Quote Request
// swagger:model QuoteRequestNotification
type QuoteRequestNotification struct {

	// Unique id for this quote as set by the quote service
	// Required: true
	QuoteID *string `json:"quote_id"`

	// quote request
	// Required: true
	QuoteRequest *QuoteRequest `json:"quote_request"`
}

// Validate validates this quote request notification
func (m *QuoteRequestNotification) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateQuoteID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateQuoteRequest(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *QuoteRequestNotification) validateQuoteID(formats strfmt.Registry) error {

	if err := validate.Required("quote_id", "body", m.QuoteID); err != nil {
		return err
	}

	return nil
}

func (m *QuoteRequestNotification) validateQuoteRequest(formats strfmt.Registry) error {

	if err := validate.Required("quote_request", "body", m.QuoteRequest); err != nil {
		return err
	}

	if m.QuoteRequest != nil {
		if err := m.QuoteRequest.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("quote_request")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *QuoteRequestNotification) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *QuoteRequestNotification) UnmarshalBinary(b []byte) error {
	var res QuoteRequestNotification
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
