//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/link_module"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// ---- mailbox: captures Brevo calls instead of sending ----

type mailMsg struct{ To, Subject, HTML string }

type mailbox struct {
	mu    sync.Mutex
	msgs  []mailMsg
	local http.RoundTripper // only loopback hosts (the fake PLANZI) pass through
}

func (m *mailbox) RoundTrip(req *http.Request) (*http.Response, error) {
	if h := req.URL.Hostname(); h == "127.0.0.1" || h == "localhost" {
		return m.local.RoundTrip(req)
	}
	if req.URL.Host != "api.brevo.com" {
		return nil, fmt.Errorf("integration tests may only call the fake mail host, got %s", req.URL.Host)
	}
	raw, _ := io.ReadAll(req.Body)
	var p struct {
		To      []struct{ Email string } `json:"to"`
		Subject string                   `json:"subject"`
		HTML    string                   `json:"htmlContent"`
	}
	_ = json.Unmarshal(raw, &p)
	if len(p.To) > 0 {
		m.mu.Lock()
		m.msgs = append(m.msgs, mailMsg{To: strings.ToLower(p.To[0].Email), Subject: p.Subject, HTML: p.HTML})
		m.mu.Unlock()
	}
	return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader("{}")), Header: http.Header{}, Request: req}, nil
}

// last returns the newest mail to addr whose HTML matches re (submatch 1), waiting briefly for async senders.
func (m *mailbox) last(t *testing.T, addr string, re *regexp.Regexp) string {
	t.Helper()
	got, ok := m.find(addr, re)
	if !ok {
		t.Fatalf("no mail to %s matching %s (mails: %d)", addr, re, m.count())
	}
	return got
}

func (m *mailbox) find(addr string, re *regexp.Regexp) (string, bool) {
	addr = strings.ToLower(addr)
	deadline := time.Now().Add(2 * time.Second)
	for {
		m.mu.Lock()
		for i := len(m.msgs) - 1; i >= 0; i-- {
			if m.msgs[i].To != addr {
				continue
			}
			if sm := re.FindStringSubmatch(m.msgs[i].HTML); sm != nil {
				m.mu.Unlock()
				return sm[1], true
			}
		}
		m.mu.Unlock()
		if time.Now().After(deadline) {
			return "", false
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (m *mailbox) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.msgs)
}

var (
	reVerifyToken = regexp.MustCompile(`verify-email\?token=([0-9a-f]{64})`)
	reResetToken  = regexp.MustCompile(`reset-password\?token=([0-9a-f]{64})`)
	reCode        = regexp.MustCompile(`>\s*(\d{6})\s*<`)
)

// ---- fake PLANZI remote ----

type fakeRemote struct {
	mu     sync.Mutex
	tokens map[string]link_module.RemoteUser // handoff token -> user
	users  map[string]link_module.RemoteUser // email -> user (password == "pw")
	srv    *httptest.Server
}

func newFakeRemote(t *testing.T) *fakeRemote {
	f := &fakeRemote{tokens: map[string]link_module.RemoteUser{}, users: map[string]link_module.RemoteUser{}}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		f.mu.Lock()
		defer f.mu.Unlock()
		switch r.URL.Path {
		case "/api/v1/internal/link/introspect":
			var in struct{ Token string }
			_ = json.Unmarshal(body, &in)
			ru, ok := f.tokens[in.Token]
			if !ok {
				w.WriteHeader(401)
				return
			}
			delete(f.tokens, in.Token)
			_ = json.NewEncoder(w).Encode(ru)
		case "/api/v1/internal/link/verify-password":
			var in struct{ Email, Password string }
			_ = json.Unmarshal(body, &in)
			ru, ok := f.users[strings.ToLower(in.Email)]
			if !ok || in.Password != "remote-pw" {
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": false})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": ru})
		default:
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeRemote) addUser(ru link_module.RemoteUser) {
	f.mu.Lock()
	f.users[strings.ToLower(ru.Email)] = ru
	f.mu.Unlock()
}

func (f *fakeRemote) mintToken(token string, ru link_module.RemoteUser) {
	f.mu.Lock()
	f.tokens[token] = ru
	f.mu.Unlock()
}

// ---- fake Google ----

type fakeGoogle struct {
	mu     sync.Mutex
	claims *auth_module.GoogleIdentityClaims
}

func (g *fakeGoogle) set(c auth_module.GoogleIdentityClaims) {
	g.mu.Lock()
	g.claims = &c
	g.mu.Unlock()
}

func (g *fakeGoogle) AuthorizationURL(auth_module.GoogleAuthorizationParams) string {
	return "https://google.test/auth"
}

func (g *fakeGoogle) ExchangeCode(context.Context, string, string, string, string, string) (*auth_module.GoogleTokenResult, error) {
	return &auth_module.GoogleTokenResult{IDToken: "id-token"}, nil
}

func (g *fakeGoogle) VerifyIDToken(context.Context, string, string, string) (*auth_module.GoogleIdentityClaims, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	c := *g.claims
	return &c, nil
}

// ---- environment ----

type env struct {
	t      *testing.T
	db     *gorm.DB
	cfg    *config.Config
	mail   *mailbox
	remote *fakeRemote
	google *fakeGoogle
	auth   *auth_module.Service
	users  *user_module.Service
	link   *link_module.Service
	app    *fiber.App
}

type noStats struct{}

func (noStats) GetLevel(uint) int { return 1 }

func newEnv(t *testing.T) *env {
	t.Helper()
	db := newTestDB(t)
	e := &env{t: t, db: db, cfg: testConfig(), mail: &mailbox{}, remote: newFakeRemote(t), google: &fakeGoogle{}}

	prev := http.DefaultTransport
	e.mail.local = prev
	http.DefaultTransport = e.mail
	t.Cleanup(func() { http.DefaultTransport = prev })

	e.cfg.Brevo = config.BrevoConfig{APIKey: "test", SenderEmail: "noreply@popcorn.test", SenderName: "POPCORN", Enabled: true}
	e.cfg.Google = config.GoogleConfig{Enabled: true, ClientID: "cid", ClientSecret: "cs", RedirectURL: "http://app.test/cb"}
	e.cfg.Link = config.LinkConfig{Secrets: []string{"link-secret"}, RemoteAPIURL: e.remote.srv.URL, RemoteWebURL: "http://planzi.test", SelfPlatform: "popcorn"}

	m := mailer.New(e.cfg.Brevo)
	e.app = fiber.New()
	api := e.app.Group("/api")
	e.auth = auth_module.RegisterRoutes(api, db, e.cfg, m, nil, nil)
	e.auth.SetGoogleOAuthProvider(e.google)
	e.link = link_module.RegisterRoutes(api, db, e.cfg, m, e.auth)

	mw := middleware.NewAuthMiddleware(e.cfg)
	protected := api.Group("/", mw.RequireAuth)
	protected.Get("/ping", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"uid": middleware.GetClaims(c).UserID})
	})
	e.users = user_module.RegisterRoutes(protected, db, nil, noStats{}, e.auth,
		user_module.NewBrevoPasswordResetMailer(e.cfg.Brevo), privacy_policy.NewUserAccessPolicy(db), e.cfg.Brevo)
	e.auth.SetUserService(e.users)
	return e
}

// do sends a request through the Fiber app. body may be nil, a string or any JSON value.
func (e *env) do(method, path string, body any, token string) (int, map[string]any) {
	e.t.Helper()
	var rd io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		rd = strings.NewReader(b)
	default:
		raw, _ := json.Marshal(b)
		rd = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := e.app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	if err != nil {
		e.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func (e *env) userRow(id uint) (email string, password *string, verified bool, version int) {
	e.t.Helper()
	var r struct {
		Email           string
		Password        *string
		VerifiedEmailAt *time.Time
		TokenVersion    int
	}
	if err := e.db.Raw(`SELECT email, password, verified_email_at, token_version FROM users WHERE id = ?`, id).Scan(&r).Error; err != nil {
		e.t.Fatal(err)
	}
	return r.Email, r.Password, r.VerifiedEmailAt != nil, r.TokenVersion
}

func (e *env) count(sql string, args ...any) int64 {
	e.t.Helper()
	var n int64
	if err := e.db.Raw(sql, args...).Scan(&n).Error; err != nil {
		e.t.Fatal(err)
	}
	return n
}

func (e *env) seedGoogleIdentity(uid uint, sub, email string) {
	e.t.Helper()
	if err := e.db.Exec(`INSERT INTO user_identities (user_id, provider, provider_subject, provider_email) VALUES (?, 'google', ?, ?)`, uid, sub, email).Error; err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) seedRefreshToken(uid uint, hash string) {
	e.t.Helper()
	if err := e.db.Exec(`INSERT INTO refresh_tokens (user_id, hashed_token, expires_at, created_at) VALUES (?, ?, now() + interval '1 day', now())`, uid, hash).Error; err != nil {
		e.t.Fatal(err)
	}
}

// googleLogin runs a full login-intent Google callback against the fake provider.
func (e *env) googleLogin() (*auth_module.TokenPair, *user_module.User, error) {
	e.t.Helper()
	ctx := context.Background()
	_, state, err := e.auth.InitiateGoogleLogin(ctx, "/")
	if err != nil {
		return nil, nil, err
	}
	pair, u, _, err := e.auth.HandleGoogleCallback(ctx, "code", state, state, "ua", "1.2.3.4")
	return pair, u, err
}
