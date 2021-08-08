module github.com/IBM/world-wire/crypto-service

go 1.16

replace (
	github.com/IBM/world-wire/auth-service-go => ./../auth-service-go
	github.com/IBM/world-wire/gftn-models => ./../gftn-models
	github.com/IBM/world-wire/participant-registry-client => ./../participant-registry-client
	github.com/IBM/world-wire/utility => ./../utility
	github.com/IBM/world-wire/utility/common => ./../utility/common
	github.com/IBM/world-wire/utility/database => ./../utility/database
	github.com/IBM/world-wire/utility/global-environment => ./../utility/global-environment
	github.com/IBM/world-wire/utility/nodeconfig => ./../utility/nodeconfig
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	fknsrs.biz/p/xml v0.0.0-20141012122126-75d8d9641d0d // indirect
	github.com/IBM/world-wire/auth-service-go v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/gftn-models v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/participant-registry-client v0.0.0-00010101000000-000000000000 // indirect
	github.com/IBM/world-wire/utility v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/global-environment v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/nodeconfig v0.0.0-00010101000000-000000000000
	github.com/beevik/etree v1.1.0
	github.com/go-openapi/strfmt v0.20.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/vault/api v1.1.1 // indirect
	github.com/miekg/pkcs11 v1.0.3
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/smartystreets/goconvey v1.6.4
	github.com/stellar/go v0.0.0-00010101000000-000000000000
	github.com/urfave/negroni v1.0.0
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
)
