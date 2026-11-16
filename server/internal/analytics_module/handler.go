package analytics_module

import (
	"errors"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/shared"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetMediaAnalytics(c fiber.Ctx) error {
	mediaID, err := strconv.Atoi(c.Params("mediaId"))
	if err != nil || mediaID <= 0 {
		return badRequest(c, "invalid media id")
	}

	mt := routeToMediaType(c.Params("mediaType"))
	if mt == "" {
		return badRequest(c, "invalid media type: use 'movies', 'series' or 'manga'")
	}

	viewer := mw.GetClaims(c)
	result, err := h.svc.GetMediaAnalytics(c.Context(), viewer.UserID, mediaID, mt)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"analytics": result})
}

func (h *Handler) GetTrending(c fiber.Ctx) error {
	mt := routeToMediaType(c.Params("mediaType"))
	if mt == "" {
		return badRequest(c, "invalid media type: use 'movies', 'series' or 'manga'")
	}

	viewer := mw.GetClaims(c)
	items, err := h.svc.GetTrending(c.Context(), viewer.UserID, mt)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"trending": items})
}

func routeToMediaType(route string) string {
	switch route {
	case "movies", "movie":
		return "movie"
	case "series", "tv":
		return "tv"
	case "manga":
		return "manga"
	default:
		return ""
	}
}

func handleError(c fiber.Ctx, err error) error {
	if handled, werr := shared.WriteMediaAccessError(c, err); handled {
		return werr
	}
	switch {
	case errors.Is(err, ErrInvalidMediaType):
		return badRequest(c, "media_type must be 'movie', 'tv' or 'manga'")
	case errors.Is(err, ErrInvalidMediaID):
		return badRequest(c, "invalid media_id")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}

func badRequest(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}
