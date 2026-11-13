package achievementsmodule

import (
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type Module struct {
	Service Service
}

func New(db *gorm.DB, policy privacy_policy.UserAccessPolicy) *Module {
	repo := newRepository(db)
	svc := newService(repo, policy)
	return &Module{Service: svc}
}

func (m *Module) RegisterPublicRoutes(router fiber.Router, guards ...fiber.Handler) {
	h := newHandler(m.Service)
	mw.GetRoute(router, "/users/:userID<int>/achievements", h.listUserAchievements, guards...)
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	h := newHandler(m.Service)

	router.Get("/achievements", h.listAllAchievements)
}
