package middleware

import (
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"gorm.io/gorm"
)

// TokenVerifier authenticates a bare access token (the socket handshake) with
// the same rules as RequireAuth: a valid signature is not enough, the token must
// match the user's current session epoch and the account must be active and verified.
type TokenVerifier struct {
	secret string
	db     *gorm.DB
}

func NewTokenVerifier(accessSecret string, db *gorm.DB) *TokenVerifier {
	return &TokenVerifier{secret: accessSecret, db: db}
}

func (v *TokenVerifier) VerifyToken(token string) (uint, error) {
	claims, err := ParseAccess(token, v.secret)
	if err != nil {
		return 0, err
	}
	st, found, err := identity_policy.ReadSession(v.db, claims.UserID)
	if err != nil {
		return 0, err
	}
	if !found || !st.Active || !st.Verified || st.Version != claims.TokenVersion {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}
