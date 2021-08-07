module github.com/IBM/world-wire/payment

go 1.16

replace (
	github.com/IBM/world-wire/administration-service => ./../../administration-service
	github.com/IBM/world-wire/anchor-service/handlers => ./../../anchor-service/handlers
	github.com/IBM/world-wire/auth-service-go/jwt => ./../../auth-service-go/jwt
	github.com/IBM/world-wire/auth-service-go/session => ./../../auth-service-go/session
	github.com/IBM/world-wire/crypto-service-client => ./../../crypto-service-client
	github.com/IBM/world-wire/gas-service-client => ./../../gas-service-client
	github.com/IBM/world-wire/global-whitelist-service => ./../../global-whitelist-service
	github.com/IBM/world-wire/iso20022 => ./../../iso20022
	github.com/IBM/world-wire/participant-registry-client => ./../../participant-registry-client
	github.com/IBM/world-wire/gftn-models => ./../../gftn-models
	github.com/IBM/world-wire/iso20022 => ./../../iso20022
	github.com/IBM/world-wire/utility/common => ./../common
	github.com/IBM/world-wire/utility/global-environment => ./../global-environment
	github.com/IBM/world-wire/utility/database => ./../database
	github.com/IBM/world-wire/utility => ./..
	github.com/IBM/world-wire/utility/nodeconfig => ./../nodeconfig
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)
