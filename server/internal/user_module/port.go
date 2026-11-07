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
// for the response body. expectedVersion is the epoch the password change produced
// (the caller's token version + 1): if a reset or claim moved it on since, no session
// is issued.
type SessionReissuer interface {
	ReissueSession(c fiber.Ctx, userID uint, expectedVersion int) (accessToken, refreshToken string, err error)
}
