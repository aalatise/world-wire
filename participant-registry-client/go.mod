module github.com/IBM/world-wire/participant-registry-client

go 1.16

replace (
	github.com/IBM/world-wire/gftn-models => ../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/utility/common => ../../../../github.com/IBM/world-wire/utility/common
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	github.com/IBM/world-wire/gftn-models v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.6.4
	gopkg.in/resty.v1 v1.12.0 // indirect
)
