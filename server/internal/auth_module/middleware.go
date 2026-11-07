package auth_module

import (
	"github.com/arinsuda/movie-hub/config"
	mw "github.com/arinsuda/movie-hub/middleware"
	"gorm.io/gorm"
)

type Middleware = mw.AuthMiddleware

func NewMiddleware(cfg *config.Config, db *gorm.DB) *mw.AuthMiddleware {
	return mw.NewAuthMiddleware(cfg, db)
}

var GetClaims = mw.GetClaims
