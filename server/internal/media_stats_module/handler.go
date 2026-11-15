package media_stats_module

import (
	"errors"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetStats(c fiber.Ctx) error {
	mediaID, mediaType, err := parseMediaParams(c)
	if err != nil {
		return badRequest(c, err.Error())
	}

	var requesterID uint
	if claims := mw.GetClaims(c); claims != nil {
		requesterID = claims.UserID
	}

	stats, err := h.svc.GetStats(mediaID, mediaType, requesterID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"stats": stats})
}

func (h *Handler) RecordView(c fiber.Ctx) error {
	mediaID, mediaType, err := parseMediaParams(c)
	if err != nil {
		return badRequest(c, err.Error())
	}
	claims := mw.GetClaims(c)
	if err := h.svc.RecordView(claims.UserID, mediaID, mediaType); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) LikeMedia(c fiber.Ctx) error {
	mediaID, mediaType, err := parseMediaParams(c)
	if err != nil {
		return badRequest(c, err.Error())
	}
	claims := mw.GetClaims(c)
	if err := h.svc.LikeMedia(claims.UserID, mediaID, mediaType); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) UnlikeMedia(c fiber.Ctx) error {
	mediaID, mediaType, err := parseMediaParams(c)
	if err != nil {
		return badRequest(c, err.Error())
	}
	claims := mw.GetClaims(c)
	if err := h.svc.UnlikeMedia(claims.UserID, mediaID, mediaType); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseMediaParams(c fiber.Ctx) (int, movie_module.MediaType, error) {

	mediaID, err := strconv.Atoi(c.Params("media_id"))
	if err != nil || mediaID <= 0 {
		return 0, "", errors.New("invalid media_id")
	}

	mediaType := movie_module.MediaType(c.Params("media_type"))

	if mediaType == "tv" {
		mediaType = movie_module.MediaSeries
	}

	if mediaType != movie_module.MediaMovie && mediaType != movie_module.MediaSeries {
		return 0, "", errors.New("invalid media_type, must be 'movie' or 'tv'")
	}

	return mediaID, mediaType, nil
}

func handleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrAlreadyLiked):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already liked"})
	case errors.Is(err, ErrNotLiked):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "not liked"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}

func badRequest(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

// ValidateMediaParams rejects malformed ids before later middleware (such as the view rate
// limiter) spends budget on a request the handler would refuse anyway.
func ValidateMediaParams(c fiber.Ctx) error {
	if _, _, err := parseMediaParams(c); err != nil {
		return badRequest(c, err.Error())
	}
	return c.Next()
}
