package achievementsmodule

import (
	"errors"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

const (
	msgForbidden    = "forbidden"
	msgUserNotFound = "user not found"
	msgInternal     = "internal server error"
)

type handler struct {
	svc Service
}

func newHandler(svc Service) *handler {
	return &handler{svc: svc}
}

func (h *handler) listAllAchievements(c fiber.Ctx) error {
	var filter PaginationQuery
	if err := c.Bind().Query(&filter); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}

	res, err := h.svc.ListAllAchievements(filter)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(res)
}

func (h *handler) getAchievement(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	filter := PaginationQuery{Page: 1, Limit: 1}
	res, err := h.svc.ListAllAchievements(filter)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	_ = id
	_ = res

	return fiber.NewError(fiber.StatusNotImplemented, "use service.GetByID")
}

func (h *handler) listUserAchievements(c fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid userID")
	}

	var filter UserAchievementFilter
	if err := c.Bind().Query(&filter); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}

	if raw := c.Query("unlocked"); raw != "" {
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "unlocked must be true or false")
		}
		filter.Unlocked = &b
	}

	var requesterID uint
	if claims := middleware.GetClaims(c); claims != nil {
		requesterID = claims.UserID
	}

	res, err := h.svc.ListUserAchievements(c.Context(), uint(userID), requesterID, filter)
	if err != nil {
		switch {
		case errors.Is(err, privacy_policy.ErrForbidden):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": msgForbidden})
		case errors.Is(err, privacy_policy.ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": msgUserNotFound})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msgInternal})
		}
	}
	return c.JSON(res)
}
