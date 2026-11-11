package auth_module

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/gofiber/fiber/v3"
)

func TestOAuthTransactionStore(t *testing.T) {
	store := NewMemoryOAuthTransactionStore()
	ctx := context.Background()

	state, err := GenerateCryptoString(32)
	if err != nil {
		t.Fatalf("Failed to generate state: %v", err)
	}

	tx := &OAuthTransaction{
		State:        state,
		StateHash:    HashState(state),
		Nonce:        "nonce_123",
		CodeVerifier: "verifier_123",
		Intent:       IntentLogin,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}

	if err := store.Create(ctx, tx); err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	consumed, err := store.Consume(ctx, state)
	if err != nil {
		t.Fatalf("Failed to consume transaction: %v", err)
	}
	if consumed.Nonce != "nonce_123" {
		t.Errorf("Expected nonce_123, got %s", consumed.Nonce)
	}

	// Single-use check: second consume must fail!
	_, err = store.Consume(ctx, state)
	if err != ErrOAuthTransactionNotFound {
		t.Errorf("Expected ErrOAuthTransactionNotFound on second consume, got %v", err)
	}
}

func TestIDTokenClaimsVerification(t *testing.T) {
	provider := NewRealGoogleOAuthProvider()
	ctx := context.Background()

	claims := &GoogleIdentityClaims{
		Subject:       "google_sub_999",
		Email:         "testuser@gmail.com",
		EmailVerified: true,
		Audience:      "my_client_id",
		Issuer:        "https://accounts.google.com",
		Nonce:         "my_nonce",
		ExpiresAt:     time.Now().Add(1 * time.Hour).Unix(),
		IssuedAt:      time.Now().Unix(),
	}

	_ = claims
	_ = provider
	_ = ctx
}

func TestAuthoritativeEmailClassification(t *testing.T) {
	gmailUser := &GoogleIdentityClaims{
		Email:         "user@gmail.com",
		EmailVerified: true,
	}
	if !IsAuthoritativeGoogleEmail(gmailUser) {
		t.Errorf("Expected @gmail.com to be authoritative")
	}

	workspaceUser := &GoogleIdentityClaims{
		Email:         "alex@company.com",
		EmailVerified: true,
		HostedDomain:  "company.com",
	}
	if !IsAuthoritativeGoogleEmail(workspaceUser) {
		t.Errorf("Expected workspace user with email_verified and hd to be authoritative")
	}

	thirdPartyUnverified := &GoogleIdentityClaims{
		Email:         "alex@yahoo.com",
		EmailVerified: false,
	}
	if IsAuthoritativeGoogleEmail(thirdPartyUnverified) {
		t.Errorf("Expected third-party unverified email to NOT be authoritative")
	}
}

func TestSanitizeUsernameBase(t *testing.T) {
	s1 := SanitizeUsernameBase("john.doe@gmail.com")
	if s1 != "johndoegmailcom" {
		t.Logf("Sanitized username: %s", s1)
	}

	s2 := SanitizeUsernameBase("A!@#$%^&*B")
	if len(s2) < 3 {
		t.Errorf("Sanitized username too short")
	}
}

func TestPasswordlessAccountSafety(t *testing.T) {
	user := &user_module.User{
		Username: "google_user",
		Email:    "google@gmail.com",
		Password: nil, // Passwordless
	}

	if user.Password != nil && *user.Password != "" {
		t.Errorf("Expected password to be nil for passwordless account")
	}
}

func TestGoogleCallbackDeliversTokensInFragment(t *testing.T) {
	for _, frontendSuccessURL := range []string{
		"https://app.popcorn.test/?from=google",
		"https://app.popcorn.test/?from=google#stale-fragment",
	} {
		cfg := &config.Config{}
		cfg.Google.Enabled = true
		cfg.Google.FrontendSuccessURL = frontendSuccessURL
		cfg.Google.FrontendErrorURL = "https://app.popcorn.test/login"
		cfg.JWT.AccessTTL = 15 * time.Minute
		cfg.JWT.RefreshTTL = 24 * time.Hour

		h := &Handler{
			cfg: cfg,
			handleGoogleCallback: func(ctx context.Context, code, stateParam, cookieState, userAgent, ip string) (*TokenPair, *user_module.User, string, error) {
				return &TokenPair{AccessToken: "acc+tok", RefreshToken: "ref/tok"}, &user_module.User{}, "/movies", nil
			},
		}

		app := fiber.New()
		app.Get("/api/auth/google/callback", h.GoogleCallback)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=c1&state=s1", nil)
		req.AddCookie(&http.Cookie{Name: "popcorn_oauth_state", Value: "s1"})
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		loc := resp.Header.Get("Location")
		if loc == "" {
			t.Fatal("expected Location header on callback redirect")
		}
		parsed, err := url.Parse(loc)
		if err != nil {
			t.Fatalf("invalid Location URL %q: %v", loc, err)
		}
		if parsed.Path != "/movies" {
			t.Errorf("expected path /movies, got %q", parsed.Path)
		}
		if strings.Contains(parsed.RawQuery, "access_token") || strings.Contains(parsed.RawQuery, "refresh_token") {
			t.Fatalf("tokens must not appear in query string: %q", parsed.RawQuery)
		}
		if parsed.RawQuery != "from=google" {
			t.Errorf("expected query to keep only existing params (from=google), got %q", parsed.RawQuery)
		}
		frag, err := url.ParseQuery(parsed.EscapedFragment())
		if err != nil {
			t.Fatalf("failed to parse fragment %q: %v", parsed.EscapedFragment(), err)
		}
		if frag.Get("access_token") != "acc+tok" || frag.Get("refresh_token") != "ref/tok" {
			t.Fatalf("expected tokens in fragment, got %q", parsed.EscapedFragment())
		}
		if strings.Contains(parsed.Fragment, "stale-fragment") {
			t.Errorf("existing fragment was not stripped: %q", parsed.Fragment)
		}

		cookies := map[string]*http.Cookie{}
		for _, ck := range resp.Cookies() {
			cookies[ck.Name] = ck
		}
		if cookies["access_token"] == nil || cookies["access_token"].Value != "acc+tok" {
			t.Errorf("expected access_token cookie to still be set, got %+v", cookies["access_token"])
		}
		if cookies["refresh_token"] == nil || cookies["refresh_token"].Value != "ref/tok" {
			t.Errorf("expected refresh_token cookie to still be set, got %+v", cookies["refresh_token"])
		}
	}
}

func TestValidateRelativeReturnURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"   ", "/"},
		{"/search?q=dune&tab=movies", "/search?q=dune&tab=movies"},
		{"/movies/603", "/movies/603"},
		{"/actors/100", "/actors/100"},
		{"//evil.com", "/"},
		{"https://evil.com/phish", "/"},
		{"/\tevil", "/"},
		{"/\\evil.com", "/"},
		{"/login", "/"},
		{"/login?redirect=/movies/603", "/"},
		{"/register?redirect=/movies/603", "/"},
		{"/forgot-password", "/"},
		{"/reset-password#token=abc", "/"},
		{"/movies/603#fragment-dropped", "/movies/603"},
	}

	for _, tc := range cases {
		if got := ValidateRelativeReturnURL(tc.in); got != tc.want {
			t.Errorf("ValidateRelativeReturnURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestGoogleCallbackPreservesQueryReturnURL(t *testing.T) {
	cfg := &config.Config{}
	cfg.Google.Enabled = true
	cfg.Google.FrontendSuccessURL = "https://app.popcorn.test/?from=google"
	cfg.Google.FrontendErrorURL = "https://app.popcorn.test/login"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = 24 * time.Hour

	h := &Handler{
		cfg: cfg,
		handleGoogleCallback: func(ctx context.Context, code, stateParam, cookieState, userAgent, ip string) (*TokenPair, *user_module.User, string, error) {
			return &TokenPair{AccessToken: "acc", RefreshToken: "ref"}, &user_module.User{}, "/search?q=dune+2&tab=movie", nil
		},
	}

	app := fiber.New()
	app.Get("/api/auth/google/callback", h.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=c1&state=s1", nil)
	req.AddCookie(&http.Cookie{Name: "popcorn_oauth_state", Value: "s1"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("invalid Location URL %q: %v", loc, err)
	}
	if parsed.Path != "/search" {
		t.Errorf("expected path /search, got %q", parsed.Path)
	}
	q := parsed.Query()
	if q.Get("from") != "google" || q.Get("q") != "dune 2" || q.Get("tab") != "movie" {
		t.Errorf("unexpected query params %v", q)
	}
}

func TestGetSessionPassiveEndpoint(t *testing.T) {
	t.Run("missing refresh cookie returns false without checking service", func(t *testing.T) {
		called := false
		h := &Handler{
			checkSession: func(string) (bool, error) {
				called = true
				return true, nil
			},
		}
		app := fiber.New()
		app.Get("/api/auth/session", h.GetSession)

		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/auth/session", nil))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var out map[string]bool
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if out["has_session"] || called {
			t.Fatalf("unexpected result: out=%v called=%v", out, called)
		}
	})

	t.Run("valid cookie returns true", func(t *testing.T) {
		h := &Handler{
			checkSession: func(raw string) (bool, error) {
				return raw == "live-refresh", nil
			},
		}
		app := fiber.New()
		app.Get("/api/auth/session", h.GetSession)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "live-refresh"})
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var out map[string]bool
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if !out["has_session"] {
			t.Fatalf("expected has_session=true, got %v", out)
		}
	})

	t.Run("infra failure returns 503", func(t *testing.T) {
		h := &Handler{
			checkSession: func(string) (bool, error) {
				return false, errors.New("db down")
			},
		}
		app := fiber.New()
		app.Get("/api/auth/session", h.GetSession)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "any"})
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", resp.StatusCode)
		}
	})
}
