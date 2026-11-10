package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/database"
	"github.com/arinsuda/movie-hub/internal/mailer"
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/arinsuda/movie-hub/router"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {

	if err := config.LoadDotEnv(); err != nil {
		log.Fatalf("❌ Env file error: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Config error: %v", err)
	}

	tmdb.Init(cfg)
	database.Connect(cfg)

	app := fiber.New(fiber.Config{
		AppName:      "movie-hub v1.0",
		ErrorHandler: customErrorHandler,
	})

	app.Use(cors.New(corsConfig(cfg.CORS.AllowedOrigins)))

	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()

	m := mailer.New(cfg.Brevo)
	notifHub := router.Register(workerCtx, app, database.DB, cfg, m)

	fiberHandler := adaptor.FiberApp(app)

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/socket.io") {
			origin := r.Header.Get("Origin")
			if cfg.CORS.IsAllowed(origin) {
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsAllowHeaders, ", "))
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsAllowMethods, ", "))
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			notifHub.Handler().ServeHTTP(w, r)
			return
		}
		mergeForwardedFor(r.Header)
		fiberHandler.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mainHandler,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("🛑 Shutting down server...")
		stopWorkers()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("❌ Server shutdown error: %v", err)
		}
	}()

	log.Printf("🚀 Server running on port %s (HTTP API & Socket.io WebSockets)", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server error: %v", err)
	}

	log.Println("👋 Server exited cleanly")
}

// CORS lists shared by Fiber and the /socket.io handler. X-Requested-With must stay allowed: the
// SPA sends it on every unsafe request (CSRF guard), so the preflight fails without it.
var (
	corsAllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", mw.HeaderRequestedWith}
	corsAllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
)

// corsConfig builds the Fiber CORS middleware config from the exact-origin allowlist.
func corsConfig(origins []string) cors.Config {
	return cors.Config{
		AllowOrigins:     origins,
		AllowCredentials: true,
		AllowHeaders:     corsAllowHeaders,
		AllowMethods:     corsAllowMethods,
	}
}

// mergeForwardedFor joins repeated X-Forwarded-For lines into one comma-joined
// value. The adaptor keeps only the last line, which would hide the entries
// ClientIP counts from the right.
func mergeForwardedFor(h http.Header) {
	if v := h.Values("X-Forwarded-For"); len(v) > 1 {
		h.Set("X-Forwarded-For", strings.Join(v, ", "))
	}
}

func customErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
