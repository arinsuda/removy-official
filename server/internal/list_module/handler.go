package list_module

import (
	"errors"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"log"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func callerID(c fiber.Ctx) (uint, bool) {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return 0, false
	}
	return claims.UserID, true
}

func unauthorized(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}

func (h *Handler) ListOwn(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	var mediaType shared.MediaType
	var mediaID int
	if q := c.Query("media_type"); q != "" {
		mediaType = shared.MediaType(q)
		id, err := strconv.Atoi(c.Query("media_id"))
		if err != nil {
			return handleError(c, ErrInvalidInput)
		}
		mediaID = id
		if err := validMedia(AddItemRequest{MediaID: mediaID, MediaType: mediaType}); err != nil {
			return handleError(c, err)
		}
	}
	lists, err := h.svc.ListOwn(uid, mediaType, mediaID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"lists": lists})
}

func (h *Handler) Create(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	var req CreateListRequest
	if err := c.Bind().JSON(&req); err != nil {
		return handleError(c, ErrInvalidInput)
	}
	list, err := h.svc.Create(uid, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"list": list})
}

func (h *Handler) Get(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	listID, err := paramID(c, "listId")
	if err != nil {
		return handleError(c, ErrNotFound)
	}
	page, _ := strconv.Atoi(c.Query("page"))
	res, err := h.svc.Get(c.Context(), uid, listID, page)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(res)
}

func (h *Handler) Update(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	listID, err := paramID(c, "listId")
	if err != nil {
		return handleError(c, ErrNotFound)
	}
	var req UpdateListRequest
	if err := c.Bind().JSON(&req); err != nil {
		return handleError(c, ErrInvalidInput)
	}
	list, err := h.svc.Update(uid, listID, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"list": list})
}

func (h *Handler) Delete(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	listID, err := paramID(c, "listId")
	if err != nil {
		return handleError(c, ErrNotFound)
	}
	if err := h.svc.Delete(uid, listID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) AddItem(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	listID, err := paramID(c, "listId")
	if err != nil {
		return handleError(c, ErrNotFound)
	}
	var req AddItemRequest
	if err := c.Bind().JSON(&req); err != nil {
		return handleError(c, ErrInvalidInput)
	}
	item, err := h.svc.AddItem(c.Context(), uid, listID, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"item": item})
}

func (h *Handler) RemoveItem(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	listID, err := paramID(c, "listId")
	if err != nil {
		return handleError(c, ErrNotFound)
	}
	itemID, err := paramID(c, "itemId")
	if err != nil {
		return handleError(c, ErrItemNotFound)
	}
	if err := h.svc.RemoveItem(uid, listID, itemID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Reorder(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	listID, err := paramID(c, "listId")
	if err != nil {
		return handleError(c, ErrNotFound)
	}
	var req ReorderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return handleError(c, ErrInvalidInput)
	}
	if err := h.svc.Reorder(uid, listID, req); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListForUser(c fiber.Ctx) error {
	uid, ok := callerID(c)
	if !ok {
		return unauthorized(c)
	}
	ownerID, err := paramID(c, "userId")
	if err != nil {
		return handleError(c, ErrNotFound)
	}
	lists, err := h.svc.ListForUser(c.Context(), uid, ownerID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"lists": lists})
}

func paramID(c fiber.Ctx, name string) (uint, error) {
	id, err := strconv.Atoi(c.Params(name))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

func handleError(c fiber.Ctx, err error) error {
	write := func(status int, code shared.ErrorCode) error {
		return shared.WriteAPIError(c, status, code, err.Error())
	}
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, privacy_policy.ErrNotFound):
		return write(fiber.StatusNotFound, shared.ErrorCodeListNotFound)
	case errors.Is(err, ErrItemNotFound):
		return write(fiber.StatusNotFound, shared.ErrorCodeItemNotFound)
	case errors.Is(err, ErrForbidden):
		return write(fiber.StatusForbidden, shared.ErrorCodeForbidden)
	case errors.Is(err, ErrInvalidInput):
		return write(fiber.StatusBadRequest, shared.ErrorCodeInvalidInput)
	case errors.Is(err, ErrInvalidVisibility):
		return write(fiber.StatusBadRequest, shared.ErrorCodeInvalidVisibility)
	case errors.Is(err, ErrInvalidMediaType):
		return write(fiber.StatusBadRequest, shared.ErrorCodeInvalidMediaType)
	case errors.Is(err, ErrOrderMismatch):
		return write(fiber.StatusBadRequest, shared.ErrorCodeOrderMismatch)
	case errors.Is(err, ErrItemExists):
		return write(fiber.StatusConflict, shared.ErrorCodeItemExists)
	case errors.Is(err, ErrListLimit):
		return write(fiber.StatusUnprocessableEntity, shared.ErrorCodeListLimit)
	case errors.Is(err, ErrListFull):
		return write(fiber.StatusUnprocessableEntity, shared.ErrorCodeListFull)
	default:
		log.Printf("ERROR [%s %s] list_module: %v", c.Method(), c.Path(), err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
