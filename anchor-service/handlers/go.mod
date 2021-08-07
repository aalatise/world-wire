module github.com/IBM/world-wire/anchor-service/handlers

go 1.16

replace (
	github.com/IBM/world-wire/administration-service => ../../../../../github.com/IBM/world-wire/administration-service
	github.com/IBM/world-wire/anchor-service => ../../../../../github.com/IBM/world-wire/anchor-service
	github.com/IBM/world-wire/api-service => ../../../../../github.com/IBM/world-wire/api-service
	github.com/IBM/world-wire/auth-service-go/session => ../../../../../github.com/IBM/world-wire/auth-service-go/session
	github.com/IBM/world-wire/crypto-service-client => ../../../../../github.com/IBM/world-wire/crypto-service-client
	github.com/IBM/world-wire/gftn-models => ../../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/participant-registry-client => ../../../../../github.com/IBM/world-wire/participant-registry-client
	github.com/IBM/world-wire/gas-service-client => ../../../../../github.com/IBM/world-wire/gas-service-client
	github.com/IBM/world-wire/utility => ../../../../../github.com/IBM/world-wire/utility
	github.com/IBM/world-wire/utility/common => ../../../../../github.com/IBM/world-wire/utility/common
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0

)
