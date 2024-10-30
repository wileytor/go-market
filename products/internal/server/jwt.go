package server

import (
	"fmt"

	"github.com/golang-jwt/jwt"
	j "github.com/lahnasti/go-market/common/jwt"
)

func CheckToken(tokenStr string) (int, error) {
	claims := &j.Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil {
		return -1, err
	}
	if !token.Valid {
		return -1, fmt.Errorf("invalid token")
	}
	return claims.UserID, nil
}
