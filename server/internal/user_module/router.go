package user_module

import (
	"strconv"
	"time"

	"github.com/arinsuda/movie-hub/config"
	mailerpkg "github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared/ratelimit"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

const (
	deleteFailMax    = 5
	deleteFailWindow = 15 * time.Minute
)

func New(
	db *gorm.DB,
	mc *storage.MinIOClient,
	statsSvc StatsProvider,
	emailVerifier EmailVerificationSender,
	passwordResetMailer PasswordResetMailer,
	policy privacy_policy.UserAccessPolicy,
	brevoCfg config.BrevoConfig,
) (*Service, *Handler) {
	mailer := NewBrevoMailer(brevoCfg)
	svc := NewService(db, mc, statsSvc, mailer, emailVerifier, passwordResetMailer, policy)
	svc.SetErasureMailer(mailerpkg.New(brevoCfg))
	h := NewHandler(svc)
	return svc, h
}

func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/users/:userId<int>", h.GetProfile, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	userKey := func(c fiber.Ctx) string {
		claims := mw.GetClaims(c)
		if claims == nil {
			return ""
		}
		return "user:" + strconv.FormatUint(uint64(claims.UserID), 10)
	}

	// One limiter for both self-removal routes: attempts on either count together.
	selfRemovalLimit := ratelimit.NewFailedOnly(deleteFailMax, deleteFailWindow, userKey)

	users := router.Group("/users")
	users.Patch("/:userId", h.UpdateProfile)
	users.Delete("/:userId", selfRemovalLimit, h.DeleteUser)
	users.Post("/me/erasure-request", requireJSON, selfRemovalLimit, h.RequestErasure)
	users.Patch("/:userId/genres", h.UpdateFavoriteGenres)
	users.Post("/:userId/email", ratelimit.New(5, time.Hour, userKey), h.RequestEmailChange)
	users.Put("/:userId/email", h.VerifyEmailChange)
	users.Patch("/:userId/password", h.ChangePassword)
}
