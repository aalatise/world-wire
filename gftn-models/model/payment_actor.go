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

// PaymentActor actor
//
// Actor
// swagger:model PaymentActor
type PaymentActor struct {

	// customer
	Customer *PaymentAddress `json:"customer,omitempty"`

	// institution
	// Required: true
	Institution *FinancialInstitutionDefinition `json:"institution"`

	// payout location
	PayoutLocation *PayoutLocation `json:"payout_location,omitempty"`
}

// Validate validates this payment actor
func (m *PaymentActor) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCustomer(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInstitution(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePayoutLocation(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *PaymentActor) validateCustomer(formats strfmt.Registry) error {

	if swag.IsZero(m.Customer) { // not required
		return nil
	}

	if m.Customer != nil {
		if err := m.Customer.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("customer")
			}
			return err
		}
	}

	return nil
}

func (m *PaymentActor) validateInstitution(formats strfmt.Registry) error {

	if err := validate.Required("institution", "body", m.Institution); err != nil {
		return err
	}

	if m.Institution != nil {
		if err := m.Institution.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("institution")
			}
			return err
		}
	}

	return nil
}

func (m *PaymentActor) validatePayoutLocation(formats strfmt.Registry) error {

	if swag.IsZero(m.PayoutLocation) { // not required
		return nil
	}

	if m.PayoutLocation != nil {
		if err := m.PayoutLocation.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("payout_location")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *PaymentActor) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *PaymentActor) UnmarshalBinary(b []byte) error {
	var res PaymentActor
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
