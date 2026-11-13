package follow_module

import (
	"errors"
	"strconv"

	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Follow(c fiber.Ctx) error {
	targetID, err := parseID(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	claims := mw.GetClaims(c)

	resp, err := h.svc.Follow(c.Context(), claims.UserID, targetID)
	if err != nil {
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) Unfollow(c fiber.Ctx) error {
	targetID, err := parseID(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.Unfollow(c.Context(), claims.UserID, targetID); err != nil {
		return handleErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetFollowers(c fiber.Ctx) error {
	userID, err := parseID(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	var requesterID uint
	if claims := mw.GetClaims(c); claims != nil {
		requesterID = claims.UserID
	}

	list, err := h.svc.GetFollowers(c.Context(), requesterID, userID)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(fiber.Map{"followers": list})
}

func (h *Handler) GetFollowing(c fiber.Ctx) error {
	userID, err := parseID(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	var requesterID uint
	if claims := mw.GetClaims(c); claims != nil {
		requesterID = claims.UserID
	}

	list, err := h.svc.GetFollowing(c.Context(), requesterID, userID)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(fiber.Map{"following": list})
}

func (h *Handler) GetPendingRequests(c fiber.Ctx) error {
	userID, err := parseID(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	claims := mw.GetClaims(c)

	list, err := h.svc.GetPendingRequests(c.Context(), claims.UserID, userID)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(fiber.Map{"requests": list})
}

func (h *Handler) AcceptRequest(c fiber.Ctx) error {
	followerID, err := parseID(c, "followerId")
	if err != nil {
		return badRequest(c, "invalid follower id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.AcceptFollow(c.Context(), claims.UserID, followerID); err != nil {
		return handleErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RejectRequest(c fiber.Ctx) error {
	followerID, err := parseID(c, "followerId")
	if err != nil {
		return badRequest(c, "invalid follower id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.RejectFollow(c.Context(), claims.UserID, followerID); err != nil {
		return handleErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetFollowStatus(c fiber.Ctx) error {
	targetID, err := parseID(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	claims := mw.GetClaims(c)

	status, err := h.svc.GetRelationshipStatus(c.Context(), claims.UserID, targetID)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(status)
}

func (h *Handler) GetFollowStats(c fiber.Ctx) error {
	targetID, err := parseID(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	claims := mw.GetClaims(c)

	stats, err := h.svc.GetFollowStats(c.Context(), claims.UserID, targetID)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(stats)
}

func parseID(c fiber.Ctx, param string) (uint, error) {
	id, err := strconv.Atoi(c.Params(param))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

func badRequest(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func handleErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrAlreadyFollowing):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already following"})
	case errors.Is(err, ErrNotFollowing):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not following"})
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	case errors.Is(err, ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
	case errors.Is(err, ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
