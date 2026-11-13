package event_module

import (
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func New(db *gorm.DB, mc *storage.MinIOClient) *Handler {
	svc := NewService(db, mc)
	return NewHandler(svc)
}

func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/events", h.ListPublic, guards...)
	mw.GetRoute(router, "/events/:id<int>", h.GetPublic, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	events := router.Group("/events")
	events.Get("/popup", h.Popup)
}

func RegisterAdminRoutes(router fiber.Router, h *Handler, requireAdmin fiber.Handler) {
	admin := router.Group("/admin/events", requireAdmin)
	admin.Get("", h.ListAdmin)
	admin.Post("", h.Create)
	admin.Put("/:id<int>", h.Update)
	admin.Delete("/:id<int>", h.Delete)
	admin.Post("/:id<int>/image", h.SetImage)
	admin.Delete("/:id<int>/image", h.RemoveImage)
}
