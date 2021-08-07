module github.com/IBM/world-wire/anchor-service

go 1.16

replace (
	github.com/IBM/world-wire/administration-service => ../../../../github.com/IBM/world-wire/administration-service
	github.com/IBM/world-wire/api-service => ../../../../github.com/IBM/world-wire/api-service
	github.com/IBM/world-wire/crypto-service-client => ../../../../github.com/IBM/world-wire/crypto-service-client
	github.com/IBM/world-wire/gas-service-client => ../../../../github.com/IBM/world-wire/gas-service-client
	github.com/IBM/world-wire/gftn-models => ../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/global-whitelist-service => ../../../../github.com/IBM/world-wire/global-whitelist-service
	github.com/IBM/world-wire/iso20022 => ../../../../github.com/IBM/world-wire/iso20022
	github.com/IBM/world-wire/participant-registry-client => ../../../../github.com/IBM/world-wire/participant-registry-client
	github.com/IBM/world-wire/utility => ../../../../github.com/IBM/world-wire/utility
	github.com/IBM/world-wire/utility/common => ../../../../github.com/IBM/world-wire/utility/common
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	github.com/IBM/world-wire/administration-service v0.0.0-00010101000000-000000000000 // indirect
	github.com/IBM/world-wire/api-service v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/auth-service-go v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/crypto-service-client v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/gas-service-client v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/gftn-models v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/global-whitelist-service v0.0.0-00010101000000-000000000000 // indirect
	github.com/IBM/world-wire/iso20022 v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/participant-registry-client v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/confluentinc/confluent-kafka-go v1.7.0 // indirect
	github.com/go-openapi/strfmt v0.20.1
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/lestrrat-go/libxml2 v0.0.0-20201123224832-e6d9de61b80d // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/stellar/go v0.0.0-00010101000000-000000000000
	github.com/urfave/negroni v1.0.0
	gopkg.in/xmlpath.v1 v1.0.0-20140413065638-a146725ea6e7 // indirect
)
