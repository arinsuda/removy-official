package user_module

import "github.com/gofiber/fiber/v3"

type StatsProvider interface {
	GetLevel(userID uint) int
}

type EmailVerificationSender interface {
	SendVerification(userID uint, email string) error
}

// SessionReissuer starts a fresh session for the caller after their sessions were
// revoked (password change), writing the new cookies on c and returning the tokens
// for the response body.
type SessionReissuer interface {
	ReissueSession(c fiber.Ctx, userID uint) (accessToken, refreshToken string, err error)
}
