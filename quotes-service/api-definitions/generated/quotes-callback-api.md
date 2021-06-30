# World Wire API Service Callbacks
API endpoints clients are expected to implement in order to receive notifications of transactions

## Version: 1.0.0

### /quote

#### POST
##### Summary:

Create a quote

##### Description:

Provides a quote in response to requests for a given target asset in exchange for a source asset, using the source asset as its price.


##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| QuoteRequestNotification | body | Quote request to RFI detailing quoteID target asset, source asset, and amount desired to exchange.  | Yes | [quoteRequestNotification](#quoterequestnotification) |

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | Successfully receive a valid quote request.  |
| 404 | Unsuccessfully receive a valid quote request. |

### Models


#### asset

Details of the asset being transacted

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| asset_code | string | Alphanumeric code for the asset - USD, XLM, etc | Yes |
| asset_type | string | The type of asset. Options include digital obligation, "DO", digital asset "DA", or a cryptocurrency "native". | Yes |
| issuer_id | string | The asset issuer's participant id. | No |

#### quoteRequest

Quote Request

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| limit_max | number | Maximum units for which this quote is valid | Yes |
| limit_min | number | Minimum units for which this quote is valid | No |
| ofi_id | string | The ID that identifies the OFI Participant on the WorldWire network (i.e. uk.yourbankintheUK.payments.ibm.com). | Yes |
| source_asset | [asset](#asset) |  | No |
| target_asset | [asset](#asset) |  | No |
| time_expire | number (int64) | End-time for this quote request to be valid | Yes |

#### quoteRequestNotification

Quote Request

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| quote_id | string | Unique id for this quote as set by the quote service | Yes |
| quote_request | [quoteRequest](#quoterequest) |  | Yes |