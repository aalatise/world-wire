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

// Geo geo
//
// Geographic coordinates for a location. Based on https://schema.org/geo
// swagger:model Geo
type Geo struct {

	// The geo coordinates
	// Required: true
	Coordinates []*Coordinate `json:"coordinates"`

	// The type of location. Options include "point" if the location is a single pickup location, or "area" if it's a region.
	//
	// Required: true
	// Enum: [area point]
	Type *string `json:"type"`
}

// Validate validates this geo
func (m *Geo) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCoordinates(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Geo) validateCoordinates(formats strfmt.Registry) error {

	if err := validate.Required("coordinates", "body", m.Coordinates); err != nil {
		return err
	}

	for i := 0; i < len(m.Coordinates); i++ {
		if swag.IsZero(m.Coordinates[i]) { // not required
			continue
		}

		if m.Coordinates[i] != nil {
			if err := m.Coordinates[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("coordinates" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

var geoTypeTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["area","point"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		geoTypeTypePropEnum = append(geoTypeTypePropEnum, v)
	}
}

const (

	// GeoTypeArea captures enum value "area"
	GeoTypeArea string = "area"

	// GeoTypePoint captures enum value "point"
	GeoTypePoint string = "point"
)

// prop value enum
func (m *Geo) validateTypeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, geoTypeTypePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *Geo) validateType(formats strfmt.Registry) error {

	if err := validate.Required("type", "body", m.Type); err != nil {
		return err
	}

	// value enum
	if err := m.validateTypeEnum("type", "body", *m.Type); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Geo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Geo) UnmarshalBinary(b []byte) error {
	var res Geo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}