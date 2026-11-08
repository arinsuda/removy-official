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

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))

	m := mailer.New(cfg.Brevo)
	notifHub := router.Register(app, database.DB, cfg, m)

	fiberHandler := adaptor.FiberApp(app)

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/socket.io") {
			origin := r.Header.Get("Origin")
			if cfg.CORS.IsAllowed(origin) {
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
