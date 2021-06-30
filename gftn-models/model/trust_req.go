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

// TrustReq TrustRequest
// swagger:model TrustReq
type TrustReq struct {

	// account name
	// Required: true
	AccountName *string `json:"account_name" validate:"required"`

	// allow approved by
	AllowApprovedBy string `json:"allowApprovedBy,omitempty" validate:"omitempty"`

	// allow initiated by
	AllowInitiatedBy string `json:"allowInitiatedBy,omitempty" validate:"omitempty"`

	// approval ids
	// Required: true
	ApprovalIds []string `json:"approval_ids" validate:"required"`

	// asset code
	// Required: true
	AssetCode *string `json:"asset_code" validate:"required"`

	// issuer id
	// Required: true
	IssuerID *string `json:"issuer_id" validate:"required"`

	// key
	Key string `json:"key,omitempty" validate:"omitempty"`

	// limit
	// Required: true
	Limit *int64 `json:"limit" validate:"required"`

	// loaded
	Loaded string `json:"loaded,omitempty" validate:"omitempty"`

	// reason rejected
	ReasonRejected string `json:"reason_rejected,omitempty" validate:"omitempty"`

	// reject approved by
	RejectApprovedBy string `json:"rejectApprovedBy,omitempty" validate:"omitempty"`

	// reject initiated by
	RejectInitiatedBy string `json:"rejectInitiatedBy,omitempty" validate:"omitempty"`

	// request approved by
	RequestApprovedBy string `json:"requestApprovedBy,omitempty" validate:"omitempty"`

	// request initiated by
	RequestInitiatedBy string `json:"requestInitiatedBy,omitempty" validate:"omitempty"`

	// requestor id
	// Required: true
	RequestorID *string `json:"requestor_id" validate:"required"`

	// revoke approved by
	RevokeApprovedBy string `json:"revokeApprovedBy,omitempty" validate:"omitempty"`

	// revoke initiated by
	RevokeInitiatedBy string `json:"revokeInitiatedBy,omitempty" validate:"omitempty"`

	// status
	Status string `json:"status,omitempty" validate:"omitempty"`

	// time updated
	// Required: true
	TimeUpdated *int64 `json:"time_updated" validate:"omitempty"`
}

// Validate validates this trust req
func (m *TrustReq) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAccountName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateApprovalIds(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateAssetCode(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIssuerID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLimit(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRequestorID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTimeUpdated(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *TrustReq) validateAccountName(formats strfmt.Registry) error {

	if err := validate.Required("account_name", "body", m.AccountName); err != nil {
		return err
	}

	return nil
}

func (m *TrustReq) validateApprovalIds(formats strfmt.Registry) error {

	if err := validate.Required("approval_ids", "body", m.ApprovalIds); err != nil {
		return err
	}

	return nil
}

func (m *TrustReq) validateAssetCode(formats strfmt.Registry) error {

	if err := validate.Required("asset_code", "body", m.AssetCode); err != nil {
		return err
	}

	return nil
}

func (m *TrustReq) validateIssuerID(formats strfmt.Registry) error {

	if err := validate.Required("issuer_id", "body", m.IssuerID); err != nil {
		return err
	}

	return nil
}

func (m *TrustReq) validateLimit(formats strfmt.Registry) error {

	if err := validate.Required("limit", "body", m.Limit); err != nil {
		return err
	}

	return nil
}

func (m *TrustReq) validateRequestorID(formats strfmt.Registry) error {

	if err := validate.Required("requestor_id", "body", m.RequestorID); err != nil {
		return err
	}

	return nil
}

func (m *TrustReq) validateTimeUpdated(formats strfmt.Registry) error {

	if err := validate.Required("time_updated", "body", m.TimeUpdated); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *TrustReq) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TrustReq) UnmarshalBinary(b []byte) error {
	var res TrustReq
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
