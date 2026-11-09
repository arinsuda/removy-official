package bmol_module

import (
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB, policy privacy_policy.UserAccessPolicy) {
	svc := NewService(db, policy)
	h := NewHandler(svc)

	bmol := router.Group("/bmol")
	bmol.Post("", h.AddItem)
	bmol.Get("/user/:userId", h.GetUserBMOL)
	bmol.Put("/:itemId", h.UpdateItem)
	bmol.Delete("/:itemId", h.RemoveItem)
}
