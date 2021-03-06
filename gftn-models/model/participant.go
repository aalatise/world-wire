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

// Participant participant
//
// Participant
// swagger:model Participant
type Participant struct {

	// The business identifier code of each participant
	// Required: true
	// Max Length: 11
	// Min Length: 11
	// Pattern: ^[A-Z]{3}[A-Z]{3}[A-Z2-9]{1}[A-NP-Z0-9]{1}[A-Z0-9]{3}$
	Bic *string `json:"bic" bson:"bic"`

	// Participant's country of residence, country code in ISO 3166-1 format
	// Required: true
	// Max Length: 3
	// Min Length: 3
	CountryCode *string `json:"country_code" bson:"country_code"`

	// The participant id for the participant
	// Required: true
	// Max Length: 32
	// Min Length: 5
	// Pattern: ^[a-zA-Z0-9-]{5,32}$
	ID *string `json:"id" bson:"id"`

	// The ledger address belonging to the issuing account.
	IssuingAccount string `json:"issuing_account,omitempty" bson:"issuing_account"`

	// Accounts
	OperatingAccounts []*Account `json:"operating_accounts" bson:"operating_accounts"`

	// The Role of this registered participant, it can be MM for Market Maker and IS for Issuer or anchor
	// Required: true
	// Max Length: 2
	// Min Length: 2
	// Enum: [MM IS]
	Role *string `json:"role" bson:"role"`

	// Participant active status on WW network, inactive, active, suspended
	Status string `json:"status,omitempty" bson:"status"`
}

// Validate validates this participant
func (m *Participant) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBic(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCountryCode(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOperatingAccounts(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRole(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Participant) validateBic(formats strfmt.Registry) error {

	if err := validate.Required("bic", "body", m.Bic); err != nil {
		return err
	}

	if err := validate.MinLength("bic", "body", string(*m.Bic), 11); err != nil {
		return err
	}

	if err := validate.MaxLength("bic", "body", string(*m.Bic), 11); err != nil {
		return err
	}

	if err := validate.Pattern("bic", "body", string(*m.Bic), `^[A-Z]{3}[A-Z]{3}[A-Z2-9]{1}[A-NP-Z0-9]{1}[A-Z0-9]{3}$`); err != nil {
		return err
	}

	return nil
}

func (m *Participant) validateCountryCode(formats strfmt.Registry) error {

	if err := validate.Required("country_code", "body", m.CountryCode); err != nil {
		return err
	}

	if err := validate.MinLength("country_code", "body", string(*m.CountryCode), 3); err != nil {
		return err
	}

	if err := validate.MaxLength("country_code", "body", string(*m.CountryCode), 3); err != nil {
		return err
	}

	return nil
}

func (m *Participant) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	if err := validate.MinLength("id", "body", string(*m.ID), 5); err != nil {
		return err
	}

	if err := validate.MaxLength("id", "body", string(*m.ID), 32); err != nil {
		return err
	}

	if err := validate.Pattern("id", "body", string(*m.ID), `^[a-zA-Z0-9-]{5,32}$`); err != nil {
		return err
	}

	return nil
}

func (m *Participant) validateOperatingAccounts(formats strfmt.Registry) error {

	if swag.IsZero(m.OperatingAccounts) { // not required
		return nil
	}

	for i := 0; i < len(m.OperatingAccounts); i++ {
		if swag.IsZero(m.OperatingAccounts[i]) { // not required
			continue
		}

		if m.OperatingAccounts[i] != nil {
			if err := m.OperatingAccounts[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("operating_accounts" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

var participantTypeRolePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["MM","IS"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		participantTypeRolePropEnum = append(participantTypeRolePropEnum, v)
	}
}

const (

	// ParticipantRoleMM captures enum value "MM"
	ParticipantRoleMM string = "MM"

	// ParticipantRoleIS captures enum value "IS"
	ParticipantRoleIS string = "IS"
)

// prop value enum
func (m *Participant) validateRoleEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, participantTypeRolePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *Participant) validateRole(formats strfmt.Registry) error {

	if err := validate.Required("role", "body", m.Role); err != nil {
		return err
	}

	if err := validate.MinLength("role", "body", string(*m.Role), 2); err != nil {
		return err
	}

	if err := validate.MaxLength("role", "body", string(*m.Role), 2); err != nil {
		return err
	}

	// value enum
	if err := m.validateRoleEnum("role", "body", *m.Role); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Participant) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Participant) UnmarshalBinary(b []byte) error {
	var res Participant
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
