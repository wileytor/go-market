package jwt

import "github.com/golang-jwt/jwt/v4"

const SecretKey = "marketPrivateKey"

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}
