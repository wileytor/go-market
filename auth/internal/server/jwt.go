package server

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	j "github.com/lahnasti/go-market/common/jwt"
)

func CreateJWTToken(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, j.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
		},
		UserID: id,
	})
	key := []byte(j.SecretKey)
	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

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
