package bmol_module

import (
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func New(db *gorm.DB, policy privacy_policy.UserAccessPolicy) *Handler {
	svc := NewService(db, policy)
	return NewHandler(svc)
}

func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/bmol/user/:userId<int>", h.GetUserBMOL, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	bmol := router.Group("/bmol")
	bmol.Post("", h.AddItem)
	bmol.Put("/:itemId", h.UpdateItem)
	bmol.Delete("/:itemId", h.RemoveItem)
}
