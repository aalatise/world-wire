module github.com/IBM/world-wire/global-whitelist-service

go 1.16

replace (
	github.com/IBM/world-wire/gftn-models => ../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/utility/common => ../../../../github.com/IBM/world-wire/utility/common
	github.com/IBM/world-wire/utility/global-environment => ../../../../github.com/IBM/world-wire/utility/global-environment
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	github.com/IBM/world-wire/gftn-models v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/global-environment v0.0.0-00010101000000-000000000000
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/smartystreets/goconvey v1.6.4
)
