package middleware

import (
	"errors"
	"log"
	"strings"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type AuthMiddleware struct {
	cfg *config.Config
	db  *gorm.DB
}

// NewAuthMiddleware needs the database: a valid signature is not enough, the
// token must still match the user's session epoch (see identity_policy).
func NewAuthMiddleware(cfg *config.Config, db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg, db: db}
}

func (m *AuthMiddleware) RequireAuth(c fiber.Ctx) error {
	// Already authenticated by an outer group: one DB read per request is enough.
	if existing, ok := c.Locals("claims").(*Claims); ok && existing != nil {
		return c.Next()
	}

	token := c.Cookies("access_token")
	if token == "" {
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	claims, err := ParseAccess(token, m.cfg.JWT.AccessSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	st, found, err := identity_policy.ReadSession(m.db, claims.UserID)
	if err != nil {
		// A database failure is not proof the token is bad: 503 keeps the client logged in.
		log.Printf("ERROR auth: read session for user %d: %v", claims.UserID, err)
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "service unavailable"})
	}
	if !found || !st.Active || !st.Verified || st.Version != claims.TokenVersion {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	c.Locals("claims", claims)
	return c.Next()
}

func (m *AuthMiddleware) RequireRole(role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*Claims)
		if !ok || claims.Role != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}

func (m *AuthMiddleware) RequireCurrentAdmin(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*Claims)
		if !ok || claims.UserID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		var count int64
		err := db.Table("users").
			Joins("JOIN roles ON roles.id = users.role_id").
			Where("users.id = ? AND users.deleted_at IS NULL AND users.is_active = true AND roles.role_name = 'admin'", claims.UserID).
			Count(&count).Error

		if err != nil || count == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin authorization lost"})
		}

		return c.Next()
	}
}
