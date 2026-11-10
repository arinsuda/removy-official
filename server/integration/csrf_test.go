//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/router"
	"github.com/gofiber/fiber/v3"
)

// Cookie-authenticated unsafe requests need X-Requested-With; the PLANZI HMAC routes do not
// (they carry no cookie), but still fail on their own signature check.
func TestCSRFHeaderGuard(t *testing.T) {
	e := newEnv(t)
	body := `{"email":"a@x.test","password":"whatever-pass1"}`

	if code, res := e.doWith("POST", "/api/auth/login", body, "", nil, "text/plain"); code != 403 || res["code"] != "XHR_HEADER_REQUIRED" {
		t.Fatalf("login without header: %d %v", code, res)
	}
	seedUser(t, e.db, "csrf@x.test", true, "right-pass1")
	wrong := `{"email":"csrf@x.test","password":"wrong-pass1"}`
	if code, _ := e.doWith("POST", "/api/auth/login", wrong, "", map[string]string{"X-Requested-With": "XMLHttpRequest"}, "application/json"); code != 401 {
		t.Fatalf("login with header and wrong password: want 401, got %d", code)
	}
	if code, res := e.doWith("POST", "/api/internal/link/introspect", `{"token":"x"}`, "", nil, "application/json"); code != 401 {
		t.Fatalf("hmac route without header must reach the signature check (401), got %d %v", code, res)
	}
}

// router.Register itself must install the guard ahead of every route: a protected route without
// the header answers 403 XHR_HEADER_REQUIRED, not 401 from RequireAuth.
func TestRouterRegisterInstallsCSRFGuard(t *testing.T) {
	e := newEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app := fiber.New()
	cfg := *e.cfg
	// Fake S3: Register checks the bucket on boot (location lookup + HEAD answer 200 = exists).
	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<?xml version="1.0"?><LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">us-east-1</LocationConstraint>`))
	}))
	defer s3.Close()
	cfg.MinIO.Endpoint = strings.TrimPrefix(s3.URL, "http://")
	cfg.MinIO.UseSSL = false
	cfg.MinIO.BucketName = "csrf-test"
	router.Register(ctx, app, e.db, &cfg, mailer.New(cfg.Brevo))
	e.app = app

	for _, path := range []string{"/api/reviews", "/api/auth/logout", "/api/library"} {
		if code, res := e.doWith("POST", path, "{}", "", nil, "application/json"); code != 403 || res["code"] != "XHR_HEADER_REQUIRED" {
			t.Fatalf("%s: %d %v", path, code, res)
		}
	}
}
