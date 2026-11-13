package event_module

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/shared"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

// Developer-facing error text. The client shows its own translation keyed by the error code.
const (
	msgInvalidState   = "invalid state: must be 'active', 'upcoming', or 'past'"
	msgInvalidStatus  = "invalid status: must be 'draft', 'published', or 'cancelled'"
	msgUnauthorized   = "unauthorized"
	msgInvalidEventID = "invalid event id"
	msgInvalidBody    = "invalid request body"
	msgMissingImage   = "missing image file"
	msgImageTooLarge  = "image too large (max 5 MB)"
	msgReadImage      = "cannot read image file"
	msgInspectImage   = "cannot inspect image"
	msgRewindImage    = "cannot rewind image"
	msgImageType      = "unsupported image type: must be JPEG, PNG, WebP, or GIF"
	msgInvalidInput   = "invalid input"
	msgEventNotFound  = "event not found"
	msgInternal       = "internal server error"

	defaultPageParam = "1"
	// imageSniffBytes is how much of an upload http.DetectContentType reads.
	imageSniffBytes = 512
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListPublic(c fiber.Ctx) error {
	stateStr := c.Query("state", string(StateActive))
	state := EventState(stateStr)
	if !state.Valid() {
		return badRequest(c, msgInvalidState)
	}
	page, _ := strconv.Atoi(c.Query("page", defaultPageParam))
	limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultPageLimit)))

	events, pagination, err := h.svc.ListPublic(c.Context(), state, page, limit)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{
		"events":     events,
		"pagination": pagination,
	})
}

func (h *Handler) GetPublic(c fiber.Ctx) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return badID(c, msgInvalidEventID)
	}

	ev, err := h.svc.GetPublic(c.Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"event": ev})
}

func (h *Handler) Popup(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, msgUnauthorized)
	}

	events, err := h.svc.Popup(c.Context())
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"events": events})
}

func (h *Handler) ListAdmin(c fiber.Ctx) error {
	statusStr := c.Query("status", "")
	status := EventStatus(statusStr)
	if statusStr != "" && !status.Valid() {
		return badRequest(c, msgInvalidStatus)
	}
	page, _ := strconv.Atoi(c.Query("page", defaultPageParam))
	limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultPageLimit)))

	events, pagination, err := h.svc.ListAdmin(c.Context(), status, page, limit)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{
		"events":     events,
		"pagination": pagination,
	})
}

func (h *Handler) Create(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, msgUnauthorized)
	}

	var in EventInput
	if err := c.Bind().JSON(&in); err != nil {
		return badRequest(c, msgInvalidBody)
	}

	ev, err := h.svc.Create(c.Context(), claims.UserID, in)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"event": ev})
}

func (h *Handler) Update(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, msgUnauthorized)
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		return badID(c, msgInvalidEventID)
	}

	var in EventInput
	if err := c.Bind().JSON(&in); err != nil {
		return badRequest(c, msgInvalidBody)
	}

	ev, err := h.svc.Update(c.Context(), claims.UserID, id, in)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"event": ev})
}

func (h *Handler) Delete(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, msgUnauthorized)
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		return badID(c, msgInvalidEventID)
	}

	if err := h.svc.Delete(c.Context(), claims.UserID, id); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) SetImage(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, msgUnauthorized)
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		return badID(c, msgInvalidEventID)
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return badRequest(c, msgMissingImage)
	}
	if fileHeader.Size > MaxImageBytes {
		return shared.WriteAPIError(c, fiber.StatusRequestEntityTooLarge, shared.ErrorCodeInvalidInput, msgImageTooLarge)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return badRequest(c, msgReadImage)
	}
	defer file.Close()

	buf := make([]byte, imageSniffBytes)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return badRequest(c, msgInspectImage)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return badRequest(c, msgRewindImage)
	}

	contentType := http.DetectContentType(buf[:n])
	if _, ok := allowedImageTypes[contentType]; !ok {
		return shared.WriteAPIError(c, fiber.StatusUnsupportedMediaType, shared.ErrorCodeUnsupportedMedia, msgImageType)
	}

	ev, err := h.svc.SetImage(c.Context(), claims.UserID, id, file, fileHeader.Size, contentType)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"event": ev})
}

func (h *Handler) RemoveImage(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, msgUnauthorized)
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		return badID(c, msgInvalidEventID)
	}

	ev, err := h.svc.RemoveImage(c.Context(), claims.UserID, id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"event": ev})
}

func parseIDParam(c fiber.Ctx, param string) (uint, error) {
	val := c.Params(param)
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

func handleError(c fiber.Ctx, err error) error {
	var valErr *ValidationError
	switch {
	case errors.As(err, &valErr):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidInput, valErr.Error())
	case errors.Is(err, ErrInvalidInput):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidInput, msgInvalidInput)
	case errors.Is(err, ErrNotFound):
		return shared.WriteAPIError(c, fiber.StatusNotFound, shared.ErrorCodeNotFound, msgEventNotFound)
	case errors.Is(err, ErrImageTooLarge):
		return shared.WriteAPIError(c, fiber.StatusRequestEntityTooLarge, shared.ErrorCodeInvalidInput, msgImageTooLarge)
	default:
		log.Printf("event_module: %v", err)
		return shared.WriteAPIError(c, fiber.StatusInternalServerError, shared.ErrorCodeInternal, msgInternal)
	}
}

func badRequest(c fiber.Ctx, msg string) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidInput, msg)
}

func badID(c fiber.Ctx, msg string) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidID, msg)
}
