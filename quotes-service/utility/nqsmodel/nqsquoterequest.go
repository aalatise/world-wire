package nqsmodel

import "github.com/shopspring/decimal"

// NqsQuoteRequest nqsQuoteRequest
//
// Asset Price Quote Request
// swagger:model NqsQuoteRequest
type NqsQuoteRequest struct {
	TimeExpireOfi            *int64           `json:"time_expire_ofi"`
	LimitMaxOfi              *decimal.Decimal `json:"limit_max_ofi"`
	LimitMinOfi              *decimal.Decimal `json:"limit_min_ofi"`
	OfiId                    *string          `json:"ofi_id"`
	SourceAsset              *Asset           `json:"source_asset"`
	TargetAsset              *Asset           `json:"target_asset"`
	IssuerAddressSourceAsset *string
	IssuerAddressTargetAsset *string
}
