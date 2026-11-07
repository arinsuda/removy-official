package user_module

import (
	"strconv"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared/ratelimit"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterRoutes(
	router fiber.Router,
	db *gorm.DB,
	mc *storage.MinIOClient,
	statsSvc StatsProvider,
	emailVerifier EmailVerificationSender,
	passwordResetMailer PasswordResetMailer,
	policy privacy_policy.UserAccessPolicy,
	brevoCfg config.BrevoConfig,
) *Service {
	mailer := NewBrevoMailer(brevoCfg)
	svc := NewService(db, mc, statsSvc, mailer, emailVerifier, passwordResetMailer, policy)
	h := NewHandler(svc)

	emailChangeKey := func(c fiber.Ctx) string {
		claims := mw.GetClaims(c)
		if claims == nil {
			return ""
		}
		return "user:" + strconv.FormatUint(uint64(claims.UserID), 10)
	}

	users := router.Group("/users")
	users.Get("/:userId", h.GetProfile)
	users.Patch("/:userId", h.UpdateProfile)
	users.Delete("/:userId", h.DeleteUser)
	users.Patch("/:userId/genres", h.UpdateFavoriteGenres)
	users.Post("/:userId/email", ratelimit.New(5, time.Hour, emailChangeKey), h.RequestEmailChange)
	users.Put("/:userId/email", h.VerifyEmailChange)
	users.Patch("/:userId/password", h.ChangePassword)

	return svc
}
