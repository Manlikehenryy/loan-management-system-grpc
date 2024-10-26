package helpers

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/manlikehenryy/loan-management-system-grpc/userService/configs"
)

var secretKey string

func Initialize() {
	if configs.Env == nil {
		panic("configs.Env is not initialized")
	}
	secretKey = configs.Env.JWT_SECRET
	if secretKey == "" {
		panic("JWT_SECRET is not set in the environment")
	}
}

func GenerateJwt(issuer string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer: issuer,
		ExpiresAt: time.Now().Add(time.Hour*24).Unix(),
	})

	return claims.SignedString([]byte(secretKey))
}

func ParseJwt(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
        return []byte(secretKey), nil
    })
    if err != nil || !token.Valid {
        return "", errors.New("invalid token")
    }

    claims, ok := token.Claims.(*jwt.StandardClaims)
    if !ok {
        return "", errors.New("invalid claims")
    }

    return claims.Issuer, nil
}