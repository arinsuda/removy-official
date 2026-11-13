package like_module

import (
	"errors"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetLikes(c fiber.Ctx) error {
	ownerID, err := parseUserID(c)
	if err != nil {
		return badRequest(c, "invalid user id")
	}

	var requesterID uint
	if claims := mw.GetClaims(c); claims != nil {
		requesterID = claims.UserID
	}

	likes, hasMore, err := h.svc.GetLikes(c.Context(), ownerID, requesterID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{"likes": likes, "has_more": hasMore})
}

func (h *Handler) Like(c fiber.Ctx) error {
	mediaID, mediaType, err := parseMediaParams(c)
	if err != nil {
		return badRequest(c, err.Error())
	}

	claims := mw.GetClaims(c)
	if err := h.svc.Like(c.Context(), claims.UserID, mediaID, mediaType); err != nil {
		return handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Unlike(c fiber.Ctx) error {
	mediaID, mediaType, err := parseMediaParams(c)
	if err != nil {
		return badRequest(c, err.Error())
	}

	claims := mw.GetClaims(c)
	if err := h.svc.Unlike(c.Context(), claims.UserID, mediaID, mediaType); err != nil {
		return handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func parseUserID(c fiber.Ctx) (uint, error) {
	id, err := strconv.Atoi(c.Params("userId"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid user id")
	}
	return uint(id), nil
}

func parseMediaParams(c fiber.Ctx) (int, movie_module.MediaType, error) {
	mediaID, err := strconv.Atoi(c.Params("mediaId"))
	if err != nil || mediaID <= 0 {
		return 0, "", errors.New("invalid media_id")
	}
	mediaType := movie_module.MediaType(c.Params("mediaType"))
	if mediaType != movie_module.MediaMovie && mediaType != movie_module.MediaSeries {
		return 0, "", errors.New("invalid media_type, must be 'movie' or 'tv'")
	}
	return mediaID, mediaType, nil
}

const errMsgUserNotFound = "user not found"

func handleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, privacy_policy.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": errMsgUserNotFound})
	case errors.Is(err, ErrAlreadyLiked):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already liked"})
	case errors.Is(err, ErrNotLiked):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "not liked"})
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}

func badRequest(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}
