module github.com/IBM/world-wire/auth-service-go

go 1.16

replace (
	github.com/IBM/world-wire/gftn-models => ./../gftn-models
	github.com/IBM/world-wire/utility => ./../utility
	github.com/IBM/world-wire/utility/common => ./../utility/common
	github.com/IBM/world-wire/utility/database => ./../utility/database
	github.com/IBM/world-wire/utility/global-environment => ./../utility/global-environment
	github.com/stellar/go => github.com/kingaj12/go v0.0.0-20210409221219-b9a73c8c53cb
)

require (
	github.com/IBM/world-wire/gftn-models v0.0.0-00010101000000-000000000000 // indirect
	github.com/IBM/world-wire/utility v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/common v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/database v0.0.0-00010101000000-000000000000
	github.com/IBM/world-wire/utility/global-environment v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	github.com/justinas/alice v1.2.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pquerna/otp v1.3.0
	go.mongodb.org/mongo-driver v1.7.1
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/oauth2 v0.0.0-20210805134026-6f1e6394065a
)
