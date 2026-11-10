package admin_module

import (
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB, hub *notification_module.Hub, mc *storage.MinIOClient, requireAdmin fiber.Handler) {
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
}
