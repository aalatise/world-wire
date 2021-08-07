module github.com/IBM/world-wire/utility

go 1.16

replace (
	github.com/IBM/world-wire/administration-service => ../../../../github.com/IBM/world-wire/administration-service
	github.com/IBM/world-wire/anchor-service => ../../../../github.com/IBM/world-wire/anchor-service
	github.com/IBM/world-wire/api-service => ../../../../github.com/IBM/world-wire/api-service
	github.com/IBM/world-wire/auth-service-go/session => ../../../../github.com/IBM/world-wire/auth-service-go/session
	github.com/IBM/world-wire/auth-service-go/jwt => ../../../../github.com/IBM/world-wire/auth-service-go/jwt
	github.com/IBM/world-wire/crypto-service => ../../../../github.com/IBM/world-wire/crypto-service
	github.com/IBM/world-wire/crypto-service-client => ../../../../github.com/IBM/world-wire/crypto-service-client
	github.com/IBM/world-wire/gas-service => ../../../../github.com/IBM/world-wire/gas-service
	github.com/IBM/world-wire/gas-service-client => ../../../../github.com/IBM/world-wire/gas-service-client
	github.com/IBM/world-wire/gftn-models => ../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/global-whitelist-service => ../../../../github.com/IBM/world-wire/global-whitelist-service
	github.com/IBM/world-wire/iso20022 => ../../../../github.com/IBM/world-wire/iso20022
	github.com/IBM/world-wire/participant-registry-client => ../../../../github.com/IBM/world-wire/participant-registry-client
	github.com/IBM/world-wire/utility/common => ../../../../github.com/IBM/world-wire/utility/common
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

