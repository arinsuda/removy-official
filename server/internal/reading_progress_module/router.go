package reading_progress_module

import (
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// progressPath is owner-only, so it mounts on the protected group (RequireAuth runs first).
const progressPath = "/reading-progress/:mediaType/:mediaId<int>"

func New(db *gorm.DB) *Handler {
	return NewHandler(NewService(db))
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	router.Get(progressPath, h.GetProgress)
	router.Put(progressPath, h.PutProgress)
	router.Delete(progressPath, h.DeleteProgress)
}
