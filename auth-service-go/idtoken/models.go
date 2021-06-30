package idtoken

import (
	"github.com/dgrijalva/jwt-go"
)

type TokenSecure struct {
	Secret string `json:"secret" bson:"secret"`
	JTI    string `json:"jti" bson:"jti"`
	Number int    `json:"number" bson:"number"`
	UID    string `json:"uid" bson:"uid"`
}

type ClaimsIDToken struct {
	jwt.StandardClaims
	UID   string `json:"uid"`
	Email string `json:"email"`
}
