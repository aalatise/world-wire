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

// PayoutLocationOpeningHour payoutLocationOpeningHours
//
// The opening hours of each payout location. Based on https://schema.org/OpeningHoursSpecification
// swagger:model PayoutLocationOpeningHour
type PayoutLocationOpeningHour struct {

	// The closing hour of the payout location on the given day(s) of the week
	// Required: true
	Closes *string `json:"closes"`

	// The day of the week for which these opening hours are valid
	// Required: true
	DayOfWeek []string `json:"day_of_week" bson:"day_of_week"`

	// The opening hour of the payout location on the given day(s) of the week
	// Required: true
	Opens *string `json:"opens"`
}

// Validate validates this payout location opening hour
func (m *PayoutLocationOpeningHour) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCloses(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDayOfWeek(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOpens(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *PayoutLocationOpeningHour) validateCloses(formats strfmt.Registry) error {

	if err := validate.Required("closes", "body", m.Closes); err != nil {
		return err
	}

	return nil
}

func (m *PayoutLocationOpeningHour) validateDayOfWeek(formats strfmt.Registry) error {

	if err := validate.Required("day_of_week", "body", m.DayOfWeek); err != nil {
		return err
	}

	return nil
}

func (m *PayoutLocationOpeningHour) validateOpens(formats strfmt.Registry) error {

	if err := validate.Required("opens", "body", m.Opens); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *PayoutLocationOpeningHour) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *PayoutLocationOpeningHour) UnmarshalBinary(b []byte) error {
	var res PayoutLocationOpeningHour
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
