package admin_module

import (
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// RegisterRoutes mounts the admin routes under /admin guarded by requireAdmin and returns that
// guarded group. Modules that add admin routes (event_module.RegisterAdminRoutes) register them on
// the returned group so they inherit the guard structurally, not by mount order.
func RegisterRoutes(router fiber.Router, db *gorm.DB, hub *notification_module.Hub, mc *storage.MinIOClient, requireAdmin fiber.Handler) fiber.Router {
	repo := NewRepository(db)
	svc := NewService(repo, hub, mc)
	h := NewHandler(svc)

	admin := router.Group("/admin", requireAdmin)

	admin.Get("/overview", h.GetOverview)
	admin.Get("/growth", h.GetGrowth)
	admin.Get("/users", h.ListUsers)
	admin.Patch("/users/:userId/role", h.UpdateUserRole)
	admin.Patch("/users/:userId/status", h.UpdateUserStatus)
	admin.Delete("/users/:userId", h.DeleteUser)
	admin.Post("/users/:userId/erasure", h.EraseUser)
	admin.Get("/reviews", h.ListReviews)
	admin.Delete("/reviews/:reviewId", h.DeleteReview)
	admin.Get("/audit-logs", h.ListAuditLogs)

	return admin
}
