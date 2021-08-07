module github.com/IBM/world-wire/administration-service

go 1.16

replace (
	github.com/IBM/world-wire/auth-service-go => ../../../../github.com/IBM/world-wire/auth-service-go
	github.com/IBM/world-wire/utility => ../../../../github.com/IBM/world-wire/utility
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)
