package auth_module

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/arinsuda/movie-hub/config"
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

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type jwtManager struct {
	cfg config.JWTConfig
}

func newJWTManager(cfg config.JWTConfig) *jwtManager {
	return &jwtManager{cfg: cfg}
}

func (j *jwtManager) Issue(userID uint, role string, version int) (*TokenPair, error) {
	access, err := j.sign(userID, role, version, AccessToken, j.cfg.AccessTTL, j.cfg.AccessSecret)
	if err != nil {
		return nil, fmt.Errorf("jwt: issue access: %w", err)
	}

	refresh, err := j.sign(userID, role, version, RefreshToken, j.cfg.RefreshTTL, j.cfg.RefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("jwt: issue refresh: %w", err)
	}

	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (j *jwtManager) ParseAccess(tokenStr string) (*Claims, error) {
	return j.parse(tokenStr, j.cfg.AccessSecret, AccessToken)
}

func (j *jwtManager) ParseRefresh(tokenStr string) (*Claims, error) {
	return j.parse(tokenStr, j.cfg.RefreshSecret, RefreshToken)
}

func (j *jwtManager) sign(userID uint, role string, version int, tokenType TokenType, ttl time.Duration, secret string) (string, error) {
	// A random ID keeps tokens unique when one user gets several in the same second;
	// refresh tokens are stored under a unique hash.
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("jwt: generate id: %w", err)
	}
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		Role:         role,
		Type:         tokenType,
		TokenVersion: version,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        hex.EncodeToString(idBytes),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (j *jwtManager) parse(tokenStr, secret string, expectedType TokenType) (*Claims, error) {
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

func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
