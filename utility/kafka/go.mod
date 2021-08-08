module github.com/IBM/world-wire/utility/kafka

go 1.16

replace (
	github.com/IBM/world-wire/crypto-service-client => ./../../crypto-service-client
	github.com/IBM/world-wire/gftn-models => ./../../gftn-models
	github.com/IBM/world-wire/gas-service-client => ./../../gas-service-client
	github.com/IBM/world-wire/global-whitelist-service => ./../../global-whitelist-service
	github.com/IBM/world-wire/iso20022 => ./../../iso20022
	github.com/IBM/world-wire/utility => ./../../utility
	github.com/IBM/world-wire/utility/common => ./../common
	github.com/IBM/world-wire/utility/database => ./../database
	github.com/IBM/world-wire/utility/global-environment => ./../global-environment
	github.com/IBM/world-wire/utility/nodeconfig => ./../nodeconfig
	github.com/IBM/world-wire/utility/payment => ./../payment
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
)
