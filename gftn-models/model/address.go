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

// Address address
//
// Address
// swagger:model Address
type Address struct {

	// The building number or identifier.
	// Required: true
	BuildingNumber *string `json:"building_number" bson:"building_number"`

	// Name of the city or town.
	// Required: true
	City *string `json:"city"`

	// Country code of the location.
	// Required: true
	Country *string `json:"country"`

	// Postal code for the location.
	// Required: true
	PostalCode *string `json:"postal_code" bson:"postal_code"`

	// Name of the state.
	// Required: true
	State *string `json:"state"`

	// The street name.
	// Required: true
	Street *string `json:"street"`
}

// Validate validates this address
func (m *Address) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBuildingNumber(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCity(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCountry(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePostalCode(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateState(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStreet(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Address) validateBuildingNumber(formats strfmt.Registry) error {

	if err := validate.Required("building_number", "body", m.BuildingNumber); err != nil {
		return err
	}

	return nil
}

func (m *Address) validateCity(formats strfmt.Registry) error {

	if err := validate.Required("city", "body", m.City); err != nil {
		return err
	}

	return nil
}

func (m *Address) validateCountry(formats strfmt.Registry) error {

	if err := validate.Required("country", "body", m.Country); err != nil {
		return err
	}

	return nil
}

func (m *Address) validatePostalCode(formats strfmt.Registry) error {

	if err := validate.Required("postal_code", "body", m.PostalCode); err != nil {
		return err
	}

	return nil
}

func (m *Address) validateState(formats strfmt.Registry) error {

	if err := validate.Required("state", "body", m.State); err != nil {
		return err
	}

	return nil
}

func (m *Address) validateStreet(formats strfmt.Registry) error {

	if err := validate.Required("street", "body", m.Street); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Address) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Address) UnmarshalBinary(b []byte) error {
	var res Address
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}