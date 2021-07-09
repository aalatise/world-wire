module github.com/IBM/world-wire/anchor-servie

go 1.13

replace (
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
	github.ibm.com/gftn/iso20022 => github.com/IBM/world-wire/iso20022 v0.0.0-20210708201302-0a511f5187f1 // indirect
	github.ibm.com/gftn/world-wire-services => github.com/IBM/world-wire v0.0.0-20210708201302-0a511f5187f1 // indirect
)

require (
	fknsrs.biz/p/xml v0.0.0-20141012122126-75d8d9641d0d // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/confluentinc/confluent-kafka-go v1.7.0 // indirect
	github.com/go-openapi/strfmt v0.20.1
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.20.2 // indirect
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/hashicorp/vault/api v1.1.0 // indirect
	github.com/lestrrat-go/libxml2 v0.0.0-20201123224832-e6d9de61b80d // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/stellar/go v0.0.0-00010101000000-000000000000
	github.com/urfave/negroni v1.0.0
	github.ibm.com/gftn/iso20022 v0.0.0-00010101000000-000000000000
	github.ibm.com/gftn/world-wire-services v0.0.0-00010101000000-000000000000
	gopkg.in/resty.v1 v1.12.0 // indirect
	gopkg.in/xmlpath.v1 v1.0.0-20140413065638-a146725ea6e7 // indirect
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
	launchpad.net/xmlpath v0.0.0-20130614043138-000000000004 // indirect
)
