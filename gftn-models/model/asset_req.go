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

// AssetReq AssetRequest
// swagger:model AssetReq
type AssetReq struct {

	// approval ids
	// Required: true
	ApprovalIds []string `json:"approvalIds" validate:"required"`

	// asset code
	// Required: true
	AssetCode *string `json:"asset_code" validate:"required"`

	// asset type
	// Required: true
	AssetType *string `json:"asset_type" validate:"required"`

	// balance
	Balance int64 `json:"balance,omitempty" validate:"omitempty"`

	// currency
	Currency string `json:"currency,omitempty" validate:"omitempty"`

	// issuer id
	IssuerID string `json:"issuer_id,omitempty" validate:"omitempty"`

	// participant Id
	// Required: true
	ParticipantID *string `json:"participantId" validate:"required"`

	// status
	Status string `json:"status,omitempty" validate:"omitempty"`

	// time updated
	TimeUpdated int64 `json:"timeUpdated,omitempty" validate:"omitempty"`
}

// Validate validates this asset req
func (m *AssetReq) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateApprovalIds(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateAssetCode(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateAssetType(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateParticipantID(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *AssetReq) validateApprovalIds(formats strfmt.Registry) error {

	if err := validate.Required("approvalIds", "body", m.ApprovalIds); err != nil {
		return err
	}

	return nil
}

func (m *AssetReq) validateAssetCode(formats strfmt.Registry) error {

	if err := validate.Required("asset_code", "body", m.AssetCode); err != nil {
		return err
	}

	return nil
}

func (m *AssetReq) validateAssetType(formats strfmt.Registry) error {

	if err := validate.Required("asset_type", "body", m.AssetType); err != nil {
		return err
	}

	return nil
}

func (m *AssetReq) validateParticipantID(formats strfmt.Registry) error {

	if err := validate.Required("participantId", "body", m.ParticipantID); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *AssetReq) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AssetReq) UnmarshalBinary(b []byte) error {
	var res AssetReq
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
