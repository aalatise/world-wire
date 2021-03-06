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

// Node InstitutionNode
// swagger:model Node
type Node struct {

	// approval ids
	ApprovalIds []string `json:"approvalIds" validate:"omitempty"`

	// bic
	// Required: true
	Bic *string `json:"bic" validate:"required"`

	// country code
	// Required: true
	CountryCode *string `json:"countryCode" validate:"required"`

	// initialized
	// Required: true
	Initialized *bool `json:"initialized" validate:"required"`

	// institution Id
	// Required: true
	InstitutionID *string `json:"institutionId" validate:"required"`

	// participant Id
	// Required: true
	ParticipantID *string `json:"participantId" validate:"required"`

	// role
	// Required: true
	Role *string `json:"role" validate:"required"`

	// status
	// Required: true
	Status []string `json:"status" validate:"required"`

	// update
	Update *Node `json:"update,omitempty"`

	// version
	Version string `json:"version,omitempty" validate:"omitempty"`
}

// Validate validates this node
func (m *Node) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBic(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCountryCode(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInitialized(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInstitutionID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateParticipantID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRole(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUpdate(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Node) validateBic(formats strfmt.Registry) error {

	if err := validate.Required("bic", "body", m.Bic); err != nil {
		return err
	}

	return nil
}

func (m *Node) validateCountryCode(formats strfmt.Registry) error {

	if err := validate.Required("countryCode", "body", m.CountryCode); err != nil {
		return err
	}

	return nil
}

func (m *Node) validateInitialized(formats strfmt.Registry) error {

	if err := validate.Required("initialized", "body", m.Initialized); err != nil {
		return err
	}

	return nil
}

func (m *Node) validateInstitutionID(formats strfmt.Registry) error {

	if err := validate.Required("institutionId", "body", m.InstitutionID); err != nil {
		return err
	}

	return nil
}

func (m *Node) validateParticipantID(formats strfmt.Registry) error {

	if err := validate.Required("participantId", "body", m.ParticipantID); err != nil {
		return err
	}

	return nil
}

func (m *Node) validateRole(formats strfmt.Registry) error {

	if err := validate.Required("role", "body", m.Role); err != nil {
		return err
	}

	return nil
}

func (m *Node) validateStatus(formats strfmt.Registry) error {

	if err := validate.Required("status", "body", m.Status); err != nil {
		return err
	}

	return nil
}

func (m *Node) validateUpdate(formats strfmt.Registry) error {

	if swag.IsZero(m.Update) { // not required
		return nil
	}

	if m.Update != nil {
		if err := m.Update.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("update")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Node) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Node) UnmarshalBinary(b []byte) error {
	var res Node
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
