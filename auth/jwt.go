package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/patrickmn/go-cache"
)

// Cache for storing refresh tokens
var RefreshCache = cache.New(20*time.Minute, 30*time.Minute)

func newJwt(ttl time.Duration) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(ttl).Unix(),
	})
}

func GenerateAccessToken() (string, error) {
	claims := newJwt(15 * time.Minute)

	token, err := claims.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}

	return token, err
}

func GenerateRefreshToken() (string, error) {
	claims := newJwt(20 * time.Minute)

	token, err := claims.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return "", err
	}

	// Store the refresh token in the cache
	RefreshCache.Add(token, nil, 20*time.Minute)

	return token, err
}

func ParseToken(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})

	return claims, err
}

func VerifyToken(tokenStr string, refresh bool) (bool, error) {
	claims := jwt.MapClaims{}
	secret := os.Getenv("ACCESS_SECRET")
	if refresh { secret = os.Getenv("REFRESH_SECRET") }

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return false, err
	} else if !token.Valid {
		return false, fmt.Errorf("Invalid token")
	}

	return true, nil
}
