package middleware

import "github.com/golang-jwt/jwt"

type UserClaims struct {
	Uid   string `json:"uid"`
	Email string `json:"email"`
	jwt.StandardClaims
}
