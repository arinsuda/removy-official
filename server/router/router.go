package router

import (
	"log"

	"github.com/arinsuda/movie-hub/config"
	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/analytics_module"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/bmol_module"
	"github.com/arinsuda/movie-hub/internal/external_rating_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	"github.com/arinsuda/movie-hub/internal/follow_module"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/library_module"
	"github.com/arinsuda/movie-hub/internal/like_module"
	"github.com/arinsuda/movie-hub/internal/link_module"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/media_stats_module"
	"github.com/arinsuda/movie-hub/internal/movie_module"
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/review_module"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/arinsuda/movie-hub/internal/user_stats_module"
	"github.com/arinsuda/movie-hub/internal/watch_log_module"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func Register(app *fiber.App, db *gorm.DB, cfg *config.Config, m *mailer.Mailer) *notification_module.Hub {

	mc, err := storage.NewMinIOClient(config.MinIOConfig{
		Endpoint:   cfg.MinIO.Endpoint,
		AccessKey:  cfg.MinIO.AccessKey,
		SecretKey:  cfg.MinIO.SecretKey,
		BucketName: cfg.MinIO.BucketName,
		UseSSL:     cfg.MinIO.UseSSL,
	})
	if err != nil {
		log.Fatalf("❌ Cannot connect to MinIO: %v", err)
	}
	log.Println("✅ MinIO connected")

	statsSvc := user_stats_module.NewService(db)
	passwordResetMailer := user_module.NewBrevoPasswordResetMailer(cfg.Brevo)

	userNotifAdapter := user_module.NewNotificationUserAdapter(db)

	verifier := middleware.NewTokenVerifier(cfg.JWT.AccessSecret, db)

	notifHub := notification_module.NewHub(verifier, cfg.CORS.AllowedOrigins)
	identity_policy.SetSessionsRevokedHook(notifHub.DisconnectUser)
	notifSvc := notification_module.NewService(db, userNotifAdapter, notifHub)

	policy := privacy_policy.NewUserAccessPolicy(db)

	achieveModule := achievementsmodule.New(db, policy)
	feedModule := feed_module.New(db, notifHub, mc)

	app.Get("/", welcomeHandler)
	app.Get("/health", healthHandler)

	api := app.Group("/api")
	api.Get("/share/proxy-image", shareImageProxyHandler(cfg.TrustedProxyHops))

	authSvc := auth_module.RegisterRoutes(api, db, cfg, m, mc, notifSvc)

	link_module.RegisterRoutes(api, db, cfg, m, authSvc)

	mw := middleware.NewAuthMiddleware(cfg, db)
	protected := api.Group("/", mw.RequireAuth)

	userSvc := user_module.RegisterRoutes(protected, db, mc, statsSvc, authSvc, passwordResetMailer, policy, cfg.Brevo)
	authSvc.SetUserService(userSvc)
	userSvc.SetSessionReissuer(authSvc)

	achieveModule.RegisterRoutes(api, mw.RequireAuth)
	feedModule.RegisterRoutes(protected)

	notification_module.RegisterRoutes(protected, notifSvc, notifHub)

	follow_module.RegisterRoutes(api, db, achieveModule.Service, notifSvc, feedModule.Service)
	review_module.RegisterRoutes(protected, db, mc, statsSvc, achieveModule.Service, notifSvc, feedModule.Service, policy)
	libSvc := library_module.RegisterRoutes(protected, db, statsSvc, achieveModule.Service, notifSvc, feedModule.Service, policy)
	watch_log_module.RegisterRoutes(protected, db, feedModule.Service, libSvc, policy)
	like_module.RegisterRoutes(protected, db, achieveModule.Service, notifSvc, feedModule.Service, policy)
	bmol_module.RegisterRoutes(protected, db)

	analytics_module.RegisterRoutes(protected, db)
	ratingRepo := review_module.NewRatingStatsReader(db)
	extRatings := external_rating_module.NewService(
		external_rating_module.NewGormStore(db), cfg.OMDb.APIKey, movie_module.SeriesIMDbResolver)
	movie_module.RegisterRoutes(protected, ratingRepo, extRatings)
	media_stats_module.RegisterRoutes(protected, db)
	user_stats_module.RegisterRoutes(protected, db)
	admin_module.RegisterRoutes(protected, db, notifHub, mc, mw.RequireCurrentAdmin(db))

	return notifHub
}

func welcomeHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "welcome to Movie Hub API"})
}

func healthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

