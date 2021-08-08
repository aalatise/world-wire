module github.com/IBM/world-wire/utility/asset

go 1.16

replace (
	github.com/IBM/world-wire/crypto-service-client => ./../../crypto-service-client
	github.com/IBM/world-wire/gas-service-client => ./../../gas-service-client
	github.com/IBM/world-wire/gftn-models => ./../../gftn-models
	github.com/IBM/world-wire/participant-registry-client => ./../../participant-registry-client
	github.com/IBM/world-wire/utility/common => ./../common
	github.com/IBM/world-wire/utility/global-environment => ./../global-environment
	github.com/IBM/world-wire/utility/nodeconfig => ./../nodeconfig
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	github.com/IBM/world-wire/crypto-service-client v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/gas-service-client v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/gftn-models v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/participant-registry-client v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/global-environment v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/nodeconfig v0.0.0-00010101000000-000000000000
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.2.0
	github.com/stellar/go v0.0.0-00010101000000-000000000000
)
