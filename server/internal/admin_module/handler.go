package admin_module

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

func (h *Handler) GetOverview(c fiber.Ctx) error {
	stats, err := h.svc.GetOverview()
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"overview": stats})
}

func (h *Handler) GetGrowth(c fiber.Ctx) error {
	growth, err := h.svc.GetGrowth()
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"growth": growth})
}

func (h *Handler) ListUsers(c fiber.Ctx) error {
	var filter UserFilter
	if err := c.Bind().Query(&filter); err != nil {
		return badRequest(c, "invalid query parameters")
	}
	res, err := h.svc.ListUsers(filter)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(res)
}

func (h *Handler) UpdateUserRole(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, "unauthorized")
	}

	targetUserID, err := parseIDParam(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}

	var req UpdateRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	if err := h.svc.UpdateUserRole(claims.UserID, targetUserID, req); err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "user role updated successfully",
		"user_id": targetUserID,
		"role":    req.Role,
	})
}

func (h *Handler) UpdateUserStatus(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, "unauthorized")
	}

	targetUserID, err := parseIDParam(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}

	var req UpdateStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	if err := h.svc.UpdateUserStatus(claims.UserID, targetUserID, req); err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"message":   "user status updated successfully",
		"user_id":   targetUserID,
		"is_active": req.IsActive,
	})
}

func (h *Handler) DeleteUser(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, "unauthorized")
	}

	targetUserID, err := parseIDParam(c, "userId")
	if err != nil {
		return badRequest(c, "invalid user id")
	}

	var req DeleteUserRequest
	_ = c.Bind().JSON(&req)

	if err := h.svc.DeleteUser(claims.UserID, targetUserID, req); err != nil {
		return handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListReviews(c fiber.Ctx) error {
	var filter ReviewFilter
	if err := c.Bind().Query(&filter); err != nil {
		return badRequest(c, "invalid query parameters")
	}
	res, err := h.svc.ListReviews(filter)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(res)
}

func (h *Handler) DeleteReview(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return shared.WriteAPIError(c, fiber.StatusUnauthorized, shared.ErrorCodeUnauthorized, "unauthorized")
	}

	reviewID, err := parseIDParam(c, "reviewId")
	if err != nil {
		return badRequest(c, "invalid review id")
	}

	var req DeleteReviewRequest
	_ = c.Bind().JSON(&req)

	if err := h.svc.DeleteReview(claims.UserID, reviewID, req); err != nil {
		return handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListAuditLogs(c fiber.Ctx) error {
	var filter AuditLogFilter
	if err := c.Bind().Query(&filter); err != nil {
		return badRequest(c, "invalid query parameters")
	}
	res, err := h.svc.ListAuditLogs(filter)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(res)
}

func parseIDParam(c fiber.Ctx, param string) (uint, error) {
	id, err := strconv.Atoi(c.Params(param))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

func handleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return shared.WriteAPIError(c, fiber.StatusNotFound, shared.ErrorCodeUserNotFound, "user not found")
	case errors.Is(err, ErrReviewNotFound):
		return shared.WriteAPIError(c, fiber.StatusNotFound, shared.ErrorCodeReviewNotFound, "review not found")
	case errors.Is(err, ErrRoleNotFound):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidRole, "invalid role: must be 'admin' or 'user'")
	case errors.Is(err, ErrSelfDeactivation):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeSelfDeactivation, "cannot deactivate your own account")
	case errors.Is(err, ErrSelfDeletion):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeSelfDeletion, "cannot delete your own account")
	case errors.Is(err, ErrUserAlreadyInStatus):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeAlreadyInStatus, "user is already in requested status")
	case errors.Is(err, ErrFinalAdminProtection):
		return shared.WriteAPIError(c, fiber.StatusConflict, shared.ErrorCodeFinalAdmin, "cannot demote or deactivate final active admin")
	case errors.Is(err, ErrInactiveUserRoleChange):
		return shared.WriteAPIError(c, fiber.StatusConflict, shared.ErrorCodeInactiveUserRole, "cannot change role of inactive user")
	default:
		return shared.WriteAPIError(c, fiber.StatusInternalServerError, shared.ErrorCodeInternal, "internal server error")
	}
}

func badRequest(c fiber.Ctx, msg string) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidInput, msg)
}
