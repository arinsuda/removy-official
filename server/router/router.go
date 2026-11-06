package router

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/config"
	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/analytics_module"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/bmol_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	"github.com/arinsuda/movie-hub/internal/follow_module"
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

	verifier := &jwtVerifier{secret: cfg.JWT.AccessSecret}

	notifHub := notification_module.NewHub(verifier, cfg.CORS.AllowedOrigins)
	notifSvc := notification_module.NewService(db, userNotifAdapter, notifHub)

	policy := privacy_policy.NewUserAccessPolicy(db)

	achieveModule := achievementsmodule.New(db, policy)
	feedModule := feed_module.New(db, notifHub, mc)

	app.Get("/", welcomeHandler)
	app.Get("/health", healthHandler)

	api := app.Group("/api")
	api.Get("/share/proxy-image", shareImageProxyHandler(cfg))

	authSvc := auth_module.RegisterRoutes(api, db, cfg, m, mc, notifSvc)

	link_module.RegisterRoutes(api, db, cfg, m, authSvc)

	mw := middleware.NewAuthMiddleware(cfg)
	protected := api.Group("/", mw.RequireAuth)

	userSvc := user_module.RegisterRoutes(protected, db, mc, statsSvc, authSvc, passwordResetMailer, policy, cfg.Brevo)
	authSvc.SetUserService(userSvc)

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
	movie_module.RegisterRoutes(protected, ratingRepo)
	media_stats_module.RegisterRoutes(protected, db)
	user_stats_module.RegisterRoutes(protected, db)
	admin_module.RegisterRoutes(protected, db, notifHub, mc, mw.RequireCurrentAdmin(db))

	return notifHub
}

type jwtVerifier struct {
	secret string
}

func (v *jwtVerifier) VerifyToken(token string) (uint, error) {
	claims, err := middleware.ParseAccess(token, v.secret)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func welcomeHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "welcome to Movie Hub API"})
}

func healthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func isPrivateHost(host string) bool {
	hostOnly := strings.ToLower(host)
	if h, _, err := net.SplitHostPort(hostOnly); err == nil {
		hostOnly = h
	}

	if hostOnly == "localhost" || hostOnly == "0.0.0.0" || strings.HasSuffix(hostOnly, ".local") || strings.HasSuffix(hostOnly, ".internal") {
		return true
	}

	ip := net.ParseIP(hostOnly)
	if ip == nil {
		return false
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	return false
}

func shareImageProxyHandler(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		rawURL := c.Query("url")
		if rawURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url parameter is required"})
		}

		parsed, err := url.Parse(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
		}

		// Security check: restrict internal/loopback/private IP addresses to prevent SSRF
		if isPrivateHost(parsed.Host) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "restricted host"})
		}

		// Fetch image from target
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(rawURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch remote image"})
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		c.Set("Content-Type", contentType)
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Cache-Control", "public, max-age=86400")

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read image stream"})
		}

		return c.Status(fiber.StatusOK).Send(bodyBytes)
	}
}
