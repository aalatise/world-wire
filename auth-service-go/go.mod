module github.com/IBM/world-wire/auth-service-go

go 1.16

replace (
	github.com/IBM/world-wire/crypto-service-client => ../../../../github.com/IBM/world-wire/crypto-service-client
	github.com/IBM/world-wire/gftn-models => ../../../../github.com/IBM/world-wire/gftn-models
	github.com/IBM/world-wire/participant-registry-client => ../../../../github.com/IBM/world-wire/participant-registry-client
	github.com/IBM/world-wire/utility => ../../../../github.com/IBM/world-wire/utility
	github.com/IBM/world-wire/utility/common => ../../../../github.com/IBM/world-wire/utility/common
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	cloud.google.com/go/firestore v1.5.0 // indirect
	firebase.google.com/go v3.13.0+incompatible // indirect
	fknsrs.biz/p/xml v0.0.0-20141012122126-75d8d9641d0d // indirect
	github.com/IBM/world-wire/crypto-service-client v0.0.0-00010101000000-000000000000 // indirect
	github.com/IBM/world-wire/gas-service-client v0.0.0-00010101000000-000000000000 // indirect
	github.com/IBM/world-wire/participant-registry-client v0.0.0-00010101000000-000000000000 // indirect
	github.com/IBM/world-wire/utility v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/beevik/etree v1.1.0 // indirect
	github.com/go-errors/errors v1.4.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	github.com/hashicorp/vault/api v1.1.1 // indirect
	github.com/jmoiron/sqlx v1.3.4 // indirect
	github.com/justinas/alice v1.2.0
	github.com/lib/pq v1.10.2 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pquerna/otp v1.3.0
	go.mongodb.org/mongo-driver v1.7.1
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/oauth2 v0.0.0-20210805134026-6f1e6394065a
)
