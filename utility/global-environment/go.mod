module github.com/IBM/world-wire/utility/global-environment

go 1.16

replace (
	github.com/IBM/world-wire/gftn-models => ./../../gftn-models
	github.com/IBM/world-wire/utility/common => ./../common
	github.com/IBM/world-wire/utility/nodeconfig => ./../nodeconfig
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)
