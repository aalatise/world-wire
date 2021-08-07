module github.com/IBM/world-wire/utility/nodeconfig

go 1.16

replace (
	github.com/IBM/world-wire/gftn-models => ../../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/utility/common => ../../../../../github.com/IBM/world-wire/utility/common
	github.com/IBM/world-wire/utility/global-environment => ../../../../../github.com/IBM/world-wire/utility/global-environment
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/IBM/world-wire/gftn-models v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/global-environment v0.0.0-00010101000000-000000000000
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.9.3
	github.com/pkg/errors v0.9.1
)
