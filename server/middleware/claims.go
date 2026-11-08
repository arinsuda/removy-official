package middleware

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID uint      `json:"uid"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	// TokenVersion is users.token_version at issue time; a mismatch revokes the token.
	// Tokens issued before the field existed decode as 0, which matches every untouched user.
	TokenVersion int `json:"ver"`
	jwt.RegisteredClaims
}

func GetClaims(c fiber.Ctx) *Claims {
	claims, _ := c.Locals("claims").(*Claims)
	return claims
}

func ParseAccess(tokenStr, secret string) (*Claims, error) {
	return parseToken(tokenStr, secret, AccessToken)
}

func ParseRefresh(tokenStr, secret string) (*Claims, error) {
	return parseToken(tokenStr, secret, RefreshToken)
}

func parseToken(tokenStr, secret string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: parse: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("jwt: invalid token")
	}
	if claims.Type != expectedType {
		return nil, fmt.Errorf("jwt: wrong token type: expected %s got %s", expectedType, claims.Type)
	}
	return claims, nil
}
