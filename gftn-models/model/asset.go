// Code generated by go-swagger; DO NOT EDIT.

package model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	"github.com/go-openapi/strfmt"

	"github.com/asaskevich/govalidator"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
	"strings"
)

// Asset asset
//
// Details of the asset being transacted
// swagger:model Asset
type Asset struct {
	// Alphanumeric code for the asset - USD, XLM, etc
	// Required: true
	AssetCode *string `json:"asset_code" bson:"asset_code"`

	// The type of asset. Options include digital obligation, "DO", digital asset "DA", or a cryptocurrency "native".
	// Required: true
	// Enum: [DO DA native]
	AssetType *string `json:"asset_type" bson:"asset_type"`

	// The asset issuer's participant id.
	// Max Length: 32
	// Min Length: 5
	// Pattern: ^[a-zA-Z0-9-]{5,32}$
	IssuerID string `json:"issuer_id,omitempty" bson:"issuer_id"`
}

// Validate validates this asset
func (m *Asset) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAssetCode(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateAssetType(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIssuerID(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Asset) validateAssetCode(formats strfmt.Registry) error {

	if err := validate.Required("asset_code", "body", m.AssetCode); err != nil {
		return err
	}

	if err := IsValidAssetCode(*m.AssetCode); err != nil {
		return err
	}

	return nil
}

var assetTypeAssetTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["DO","DA","native"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		assetTypeAssetTypePropEnum = append(assetTypeAssetTypePropEnum, v)
	}
}

const (
	// AssetAssetTypeDO captures enum value "DO"
	AssetAssetTypeDO string = "DO"

	// AssetAssetTypeDA captures enum value "DA"
	AssetAssetTypeDA string = "DA"

	// AssetAssetTypeNative captures enum value "native"
	AssetAssetTypeNative string = "native"
)

// prop value enum
func (m *Asset) validateAssetTypeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, assetTypeAssetTypePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *Asset) validateAssetType(formats strfmt.Registry) error {

	if err := validate.Required("asset_type", "body", m.AssetType); err != nil {
		return err
	}

	// value enum
	if err := m.validateAssetTypeEnum("asset_type", "body", *m.AssetType); err != nil {
		return err
	}

	return nil
}

func (m *Asset) validateIssuerID(formats strfmt.Registry) error {

	if swag.IsZero(m.IssuerID) { // not required
		return nil
	}

	if err := validate.MinLength("issuer_id", "body", string(m.IssuerID), 5); err != nil {
		return err
	}

	if err := validate.MaxLength("issuer_id", "body", string(m.IssuerID), 32); err != nil {
		return err
	}

	if err := validate.Pattern("issuer_id", "body", string(m.IssuerID), `^[a-zA-Z0-9-]{5,32}$`); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Asset) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Asset) UnmarshalBinary(b []byte) error {
	var res Asset
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

//Check if currency code for DA is valid ISO4217 currency code
func IsValidDACode(currencyCode string) error {
	if !govalidator.IsISO4217(currencyCode) {
		return errors.New(1, "asset code in the payload is not in valid ISO DA format")
	}
	return nil
}

//Check if currency code for DO is valid ISO4217 currency code
func IsValidDOCode(currencyCode string) error {
	isoAssetCode := strings.Replace(currencyCode, AssetAssetTypeDO, "", -1)
	if IsValidDACode(isoAssetCode) != nil {
		return errors.New(2, "asset code in the payload is not in valid ISO DO format")
	}
	return nil
}

//Check if DO or DA currency code is valid ISO4217 currency code
func IsValidAssetCode(currencyCode string) error {
	assetType := GetAssetType(currencyCode)
	if assetType == AssetAssetTypeDA {
		return IsValidDACode(currencyCode)
	} else if assetType == AssetAssetTypeDO {
		return IsValidDOCode(currencyCode)
	}
	return nil
}

func GetAssetType(assetCode string) string {
	if assetCode == "XLM" || assetCode == "xlm" {
		return AssetAssetTypeNative
	} else if strings.HasSuffix(assetCode, AssetAssetTypeDO) {
		return AssetAssetTypeDO
	}
	return AssetAssetTypeDA
}
