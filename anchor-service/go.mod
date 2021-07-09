module github.com/IBM/world-wire/anchor-service

go 1.16

replace (
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
	github.ibm.com/gftn/iso20022 => github.com/IBM/world-wire/iso20022 v0.0.0-20210708201302-0a511f5187f1 // indirect
	github.ibm.com/gftn/world-wire-services => github.com/IBM/world-wire v0.0.0-20210708201302-0a511f5187f1 // indirect
)

require (
	cloud.google.com/go/firestore v1.1.0 // indirect
	fknsrs.biz/p/xml v0.0.0-20141012122126-75d8d9641d0d // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/confluentinc/confluent-kafka-go v1.7.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-openapi/analysis v0.20.1 // indirect
	github.com/go-openapi/errors v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/runtime v0.19.29 // indirect
	github.com/go-openapi/strfmt v0.20.1
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.20.2 // indirect
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000
	github.com/go-swagger/go-swagger v0.27.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/hashicorp/vault/api v1.1.0 // indirect
	github.com/lestrrat-go/libxml2 v0.0.0-20201123224832-e6d9de61b80d // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/viper v1.8.1 // indirect
	github.com/stellar/go v0.0.0-00010101000000-000000000000
	github.com/urfave/negroni v1.0.0
	github.ibm.com/gftn/iso20022 v0.0.0-00010101000000-000000000000
	github.ibm.com/gftn/world-wire-services v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.5.4 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/tools v0.1.4 // indirect
	google.golang.org/api v0.44.0 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
	gopkg.in/xmlpath.v1 v1.0.0-20140413065638-a146725ea6e7 // indirect
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
	launchpad.net/xmlpath v0.0.0-20130614043138-000000000004 // indirect
)
