package auth_module

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/gofiber/fiber/v3"
)

func cookiesFrom(t *testing.T, app *fiber.App, path string) map[string]*http.Cookie {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(http.MethodPost, path, nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out := map[string]*http.Cookie{}
	for _, ck := range resp.Cookies() {
		out[ck.Name] = ck
	}
	return out
}

// Production: API on Render, SPA on Cloudflare Pages, so the cookies are cross-site
// (SameSite=None; Secure) with a Domain. Logout must expire them with the very same
// attributes or the browser keeps the old cookies.
func TestLogoutClearsCookiesWithTheAttributesTheyWereSetWith(t *testing.T) {
	cfg := &config.Config{}
	cfg.Cookie = config.CookieConfig{Domain: "api.popcorn.test", Secure: true, SameSite: "None"}
	cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL = 15*time.Minute, 24*time.Hour
	h := &Handler{cfg: cfg}

	app := fiber.New()
	app.Post("/login", func(c fiber.Ctx) error {
		SetTokenCookies(c, cfg, &TokenPair{AccessToken: "a", RefreshToken: "r"})
		return nil
	})
	app.Post("/logout", h.Logout)

	set, cleared := cookiesFrom(t, app, "/login"), cookiesFrom(t, app, "/logout")
	for _, name := range []string{"access_token", "refresh_token"} {
		s, c := set[name], cleared[name]
		if s == nil || c == nil {
			t.Fatalf("%s: missing Set-Cookie (set=%v cleared=%v)", name, s, c)
		}
		if c.Domain != s.Domain || c.Path != s.Path || c.SameSite != s.SameSite || c.Secure != s.Secure || c.HttpOnly != s.HttpOnly {
			t.Errorf("%s: clear attributes %+v differ from set attributes %+v", name, c, s)
		}
		if c.Value != "" || c.MaxAge >= 0 {
			t.Errorf("%s: not expired: %+v", name, c)
		}
	}
	if cleared["access_token"].Domain != "api.popcorn.test" || cleared["access_token"].SameSite != http.SameSiteNoneMode || !cleared["access_token"].Secure {
		t.Errorf("attributes not applied: %+v", cleared["access_token"])
	}
}
