// Code generated by go-swagger; DO NOT EDIT.

package model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// PayoutLocationCategory payoutLocationCategory
//
// Details of each payout location offer category. Based on https://schema.org/hasOfferCatalog
// swagger:model PayoutLocationCategory
type PayoutLocationCategory struct {

	// name of the category
	// Required: true
	// Enum: [delivery cash_pickup agency_pickup mobile bank_account]
	Name *string `json:"name"`

	// offer list of the category
	// Required: true
	Options []*PayoutLocationOption `json:"options"`
}

// Validate validates this payout location category
func (m *PayoutLocationCategory) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOptions(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var payoutLocationCategoryTypeNamePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["delivery","cash_pickup","agency_pickup","mobile","bank_account"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		payoutLocationCategoryTypeNamePropEnum = append(payoutLocationCategoryTypeNamePropEnum, v)
	}
}

const (

	// PayoutLocationCategoryNameDelivery captures enum value "delivery"
	PayoutLocationCategoryNameDelivery string = "delivery"

	// PayoutLocationCategoryNameCashPickup captures enum value "cash_pickup"
	PayoutLocationCategoryNameCashPickup string = "cash_pickup"

	// PayoutLocationCategoryNameAgencyPickup captures enum value "agency_pickup"
	PayoutLocationCategoryNameAgencyPickup string = "agency_pickup"

	// PayoutLocationCategoryNameMobile captures enum value "mobile"
	PayoutLocationCategoryNameMobile string = "mobile"

	// PayoutLocationCategoryNameBankAccount captures enum value "bank_account"
	PayoutLocationCategoryNameBankAccount string = "bank_account"
)

// prop value enum
func (m *PayoutLocationCategory) validateNameEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, payoutLocationCategoryTypeNamePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *PayoutLocationCategory) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	// value enum
	if err := m.validateNameEnum("name", "body", *m.Name); err != nil {
		return err
	}

	return nil
}

func (m *PayoutLocationCategory) validateOptions(formats strfmt.Registry) error {

	if err := validate.Required("options", "body", m.Options); err != nil {
		return err
	}

	for i := 0; i < len(m.Options); i++ {
		if swag.IsZero(m.Options[i]) { // not required
			continue
		}

		if m.Options[i] != nil {
			if err := m.Options[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("options" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *PayoutLocationCategory) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *PayoutLocationCategory) UnmarshalBinary(b []byte) error {
	var res PayoutLocationCategory
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
