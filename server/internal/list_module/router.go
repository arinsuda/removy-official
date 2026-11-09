package list_module

import (
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// RegisterRoutes mounts the list endpoints on the authenticated router.
func RegisterRoutes(router fiber.Router, db *gorm.DB, policy privacy_policy.UserAccessPolicy) *Service {
	svc := NewService(db, policy)
	h := NewHandler(svc)

	lists := router.Group("/lists")
	lists.Get("", h.ListOwn)
	lists.Post("", h.Create)
	lists.Get("/:listId", h.Get)
	lists.Patch("/:listId", h.Update)
	lists.Delete("/:listId", h.Delete)
	lists.Post("/:listId/items", h.AddItem)
	lists.Put("/:listId/items/order", h.Reorder)
	lists.Delete("/:listId/items/:itemId", h.RemoveItem)

	router.Get("/users/:userId/lists", h.ListForUser)
	return svc
}
