module github.com/IBM/world-wire/crypto-service-client

go 1.16

replace (
	github.com/IBM/world-wire/gftn-models => ../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/utility => ../../../../github.com/IBM/world-wire/utility
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
)
