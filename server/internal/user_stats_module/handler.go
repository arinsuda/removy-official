package user_stats_module

import (
	"errors"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc    *Service
	policy privacy_policy.UserAccessPolicy
}

func NewHandler(svc *Service, policy privacy_policy.UserAccessPolicy) *Handler {
	return &Handler{svc: svc, policy: policy}
}

// GetUserStats GET /users/:userId/stats — library counts follow the library privacy rule.
func (h *Handler) GetUserStats(c fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil || userID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var viewerID uint
	if claims := mw.GetClaims(c); claims != nil {
		viewerID = claims.UserID
	}
	canView, err := h.policy.CanViewProfileSection(c.Context(), viewerID, uint(userID), privacy_policy.SectionLibrary)
	if err != nil {
		if errors.Is(err, privacy_policy.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if !canView {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	stats, err := h.svc.GetUserStats(uint(userID))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(fiber.Map{"stats": stats})
}
