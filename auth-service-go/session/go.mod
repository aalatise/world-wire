module github.com/IBM/world-wire/auth-service-go/session

go 1.16

replace github.com/IBM/world-wire/auth-service-go/jwt => ../../../../../github.com/IBM/world-wire/auth-service-go/jwt

require (
	github.com/IBM/world-wire/auth-service-go/jwt v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/context v1.1.1
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
)
