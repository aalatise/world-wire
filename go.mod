module github.com/IBM/world-wire

go 1.16

replace github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0

replace github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb

replace github.ibm.com/gftn/iso20022 => github.com/IBM/world-wire/iso20022 v0.0.0-20210708201302-0a511f5187f1 // indirect

replace github.ibm.com/gftn/world-wire-services => github.com/IBM/world-wire v0.0.0-20210708201302-0a511f5187f1 // indirect

require (
	cloud.google.com/go/firestore v1.5.0 // indirect
	firebase.google.com/go v3.13.0+incompatible
	fknsrs.biz/p/xml v0.0.0-20141012122126-75d8d9641d0d
	github.com/BurntSushi/toml v0.3.1
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/aws/aws-sdk-go v1.39.3
	github.com/beevik/etree v1.1.0
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-errors/errors v1.4.0
	github.com/go-openapi/analysis v0.20.1 // indirect
	github.com/go-openapi/errors v0.20.0
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/runtime v0.19.29 // indirect
	github.com/go-openapi/strfmt v0.20.1
	github.com/go-openapi/swag v0.19.15
	github.com/go-openapi/validate v0.20.2
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000
	github.com/go-swagger/go-swagger v0.27.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	github.com/hashicorp/vault/api v1.1.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/justinas/alice v1.2.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/libxml2 v0.0.0-20201123224832-e6d9de61b80d
	github.com/lib/pq v1.10.2
	github.com/miekg/pkcs11 v1.0.3
	github.com/neo4j/neo4j-go-driver v1.8.3
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.9.3
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/viper v1.8.1 // indirect
	github.com/stellar/go v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	github.com/urfave/negroni v1.0.0
	github.ibm.com/gftn/iso20022 v0.0.0-00010101000000-000000000000
	github.ibm.com/gftn/world-wire-services v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.5.4
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/api v0.50.0
	googlemaps.github.io/maps v1.3.2
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/resty.v1 v1.12.0 // indirect
	gopkg.in/xmlpath.v1 v1.0.0-20140413065638-a146725ea6e7 // indirect
)
