package middleware

import (
	"github.com/golang-jwt/jwt/v5"
)

// ParseToken validates a JWT token string and populates claims.
func ParseToken(tokenString, secret string, claims *JWTClaims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
