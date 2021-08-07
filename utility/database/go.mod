module github.com/IBM/world-wire/utility/database

go 1.16

replace (
	github.com/IBM/world-wire/utility/common => ../../../../../github.com/IBM/world-wire/utility/common
	github.com/IBM/world-wire/utility/global-environment => ../../../../../github.com/IBM/world-wire/utility/global-environment
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	github.com/IBM/world-wire/utility/global-environment v0.0.0-00010101000000-000000000000
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.2
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/smartystreets/goconvey v1.6.4
	go.mongodb.org/mongo-driver v1.7.1
)
