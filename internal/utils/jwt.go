package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IsJwtExpired(jwtToken string) bool {
	token, _, err := jwt.NewParser().ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenUnverifiable) {
			return true
		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			// if the expiration time is before now, the token is expired
			return expirationTime.Before(time.Now())
		}
	}
	return false
}

func GetTokenExpiration(jwtToken string) time.Time {
	token, _, err := jwt.NewParser().ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenUnverifiable) {
			return time.Time{}
		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			return time.Unix(int64(exp), 0)
		}
	}
	return time.Time{}
}
