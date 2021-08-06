module github.com/IBM/world-wire/auth-service-go

go 1.16

replace (
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
	github.ibm.com/gftn/world-wire-services => github.com/IBM/world-wire v0.0.0-20210708201302-0a511f5187f1 // indirect
)

require (
	cloud.google.com/go/firestore v1.5.0 // indirect
	firebase.google.com/go v3.13.0+incompatible // indirect
	fknsrs.biz/p/xml v0.0.0-20141012122126-75d8d9641d0d // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-errors/errors v1.4.0 // indirect
	github.com/go-openapi/errors v0.20.0 // indirect
	github.com/go-openapi/strfmt v0.20.1 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.20.2 // indirect
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000 // indirect
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	github.com/hashicorp/vault/api v1.1.0 // indirect
	github.com/jmoiron/sqlx v1.3.4 // indirect
	github.com/justinas/alice v1.2.0
	github.com/lib/pq v1.10.2 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pquerna/otp v1.3.0
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/stellar/go v0.0.0-00010101000000-000000000000 // indirect
	github.ibm.com/gftn/world-wire-services v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.6.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	gopkg.in/resty.v1 v1.12.0 // indirect
)
