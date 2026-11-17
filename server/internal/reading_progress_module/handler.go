package reading_progress_module

import (
	"errors"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/shared"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

const (
	msgInvalidMediaType = "reading progress is kept for manga only"
	msgInvalidID        = "media id must be a positive integer"
	msgInvalidBody      = "request body must be JSON with chapter and volume"
	msgOutOfRange       = "chapter or volume is out of range"
	msgInvalidInput     = "chapter or volume must be at least 1"
	msgUnauthorized     = "sign in required"
	msgInternal         = "internal server error"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetProgress answers {"progress": {...}} or {"progress": null} when the user has none.
func (h *Handler) GetProgress(c fiber.Ctx) error {
	userID, ok := viewerID(c)
	if !ok {
		return unauthorized(c)
	}
	p, err := h.svc.Get(c.Context(), userID, targetOf(c))
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"progress": p})
}

// PutProgress replaces the stored progress with the body and answers the saved row.
func (h *Handler) PutProgress(c fiber.Ctx) error {
	userID, ok := viewerID(c)
	if !ok {
		return unauthorized(c)
	}
	// Path first, so a bad media type reports as such even when the body is also bad.
	t := targetOf(c)
	if err := validateTarget(t); err != nil {
		return handleError(c, err)
	}
	var req UpsertRequest
	if err := c.Bind().JSON(&req); err != nil {
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidBody, msgInvalidBody)
	}
	p, err := h.svc.Upsert(c.Context(), userID, t, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"progress": p})
}

// DeleteProgress is idempotent: it answers 204 whether or not a row existed.
func (h *Handler) DeleteProgress(c fiber.Ctx) error {
	userID, ok := viewerID(c)
	if !ok {
		return unauthorized(c)
	}
	if err := h.svc.Delete(c.Context(), userID, targetOf(c)); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// targetOf reads the path. A non-numeric id maps to 0, which validateTarget refuses as INVALID_ID.
func targetOf(c fiber.Ctx) Target {
	id, err := strconv.Atoi(c.Params("mediaId"))
	if err != nil {
		id = 0
	}
	return Target{MediaID: id, MediaType: shared.MediaType(c.Params("mediaType"))}
}

func viewerID(c fiber.Ctx) (uint, bool) {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return 0, false
	}
	return claims.UserID, true
}

func unauthorized(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, msgUnauthorized)
}

func handleError(c fiber.Ctx, err error) error {
	if handled, werr := shared.WriteMediaAccessError(c, err); handled {
		return werr
	}
	switch {
	case errors.Is(err, ErrInvalidMediaType):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidMediaType, msgInvalidMediaType)
	case errors.Is(err, ErrInvalidID):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidID, msgInvalidID)
	case errors.Is(err, ErrOutOfRange):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeProgressOutOfRange, msgOutOfRange)
	case errors.Is(err, ErrInvalidInput):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidInput, msgInvalidInput)
	default:
		return shared.WriteAPIError(c, fiber.StatusInternalServerError, shared.ErrorCodeInternal, msgInternal)
	}
}
