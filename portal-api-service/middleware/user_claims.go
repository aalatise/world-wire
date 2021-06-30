package middleware

import "github.com/dgrijalva/jwt-go"

type UserClaims struct {
	Uid   string `json:"uid"`
	Email string `json:"email"`
	jwt.StandardClaims
}
