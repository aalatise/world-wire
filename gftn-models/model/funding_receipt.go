// Code generated by go-swagger; DO NOT EDIT.

package model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// FundingReceipt fundingReceipt
//
// Funding Receipt
// swagger:model FundingReceipt
type FundingReceipt struct {

	// details funding
	DetailsFunding *Funding `json:"details_funding,omitempty"`

	// receipt funding
	ReceiptFunding *TransactionReceipt `json:"receipt_funding,omitempty"`
}

// Validate validates this funding receipt
func (m *FundingReceipt) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDetailsFunding(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateReceiptFunding(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *FundingReceipt) validateDetailsFunding(formats strfmt.Registry) error {

	if swag.IsZero(m.DetailsFunding) { // not required
		return nil
	}

	if m.DetailsFunding != nil {
		if err := m.DetailsFunding.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("details_funding")
			}
			return err
		}
	}

	return nil
}

func (m *FundingReceipt) validateReceiptFunding(formats strfmt.Registry) error {

	if swag.IsZero(m.ReceiptFunding) { // not required
		return nil
	}

	if m.ReceiptFunding != nil {
		if err := m.ReceiptFunding.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("receipt_funding")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *FundingReceipt) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FundingReceipt) UnmarshalBinary(b []byte) error {
	var res FundingReceipt
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
