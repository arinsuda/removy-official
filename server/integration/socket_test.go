//go:build integration

package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/middleware"
)

// engine.io polling client, just enough to open a socket.io connection with a token.
type pollingSocket struct {
	t     *testing.T
	base  string
	sid   string
	token string
}

var reSID = regexp.MustCompile(`"sid":"([^"]+)"`)

func (e *env) openSocket(t *testing.T, base, token string) *pollingSocket {
	t.Helper()
	// Same credential path as the browser: the access_token cookie on the handshake.
	do := func(method, path, body string) string {
		req, _ := http.NewRequest(method, base+path, strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		return string(raw)
	}
	open := do("GET", "/socket.io/?EIO=4&transport=polling", "")
	m := reSID.FindStringSubmatch(open)
	if m == nil {
		t.Fatalf("no engine.io open packet: %s", open)
	}
	ps := &pollingSocket{t: t, base: base, sid: m[1], token: token}
	do("POST", "/socket.io/?EIO=4&transport=polling&sid="+ps.sid, "40") // connect to the default namespace
	return ps
}

// waitEvent long-polls until a socket.io event packet `42["<name>",<json>]` arrives and returns its JSON.
func (ps *pollingSocket) waitEvent(name string) []byte {
	ps.t.Helper()
	prefix := `42["` + name + `",`
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest("GET", ps.base+"/socket.io/?EIO=4&transport=polling&sid="+ps.sid, nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: ps.token})
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ps.t.Fatal(err)
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		for _, pkt := range strings.Split(string(raw), "\x1e") {
			if strings.HasPrefix(pkt, prefix) && strings.HasSuffix(pkt, "]") {
				return []byte(strings.TrimSuffix(strings.TrimPrefix(pkt, prefix), "]"))
			}
		}
	}
	ps.t.Fatalf("no %s event received", name)
	return nil
}

func waitCount(t *testing.T, hub *notification_module.Hub, want int) bool {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if hub.UniqueOnlineCount() == want {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return hub.UniqueOnlineCount() == want
}

// N2: sockets authenticate once, at connect. A claim must drop live sockets and
// a pre-claim token must not open new ones.
func TestN2_ClaimDisconnectsSocketsAndBlocksReconnect(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")

	hub := notification_module.NewHub(middleware.NewTokenVerifier(e.cfg.JWT.AccessSecret, e.db), []string{"http://app.test"})
	identity_policy.SetSessionsRevokedHook(hub.DisconnectUser)
	t.Cleanup(func() { identity_policy.SetSessionsRevokedHook(nil) })
	srv := httptest.NewServer(hub.Handler())
	t.Cleanup(srv.Close)

	atk := e.accessToken(uid, 0)
	e.openSocket(t, srv.URL, atk)
	if !waitCount(t, hub, 1) {
		t.Fatalf("socket did not register (online=%d)", hub.UniqueOnlineCount())
	}

	e.claimByReset(uid, victim)

	if !waitCount(t, hub, 0) {
		t.Fatalf("attacker socket still connected after the claim (online=%d)", hub.UniqueOnlineCount())
	}
	e.openSocket(t, srv.URL, atk)
	time.Sleep(500 * time.Millisecond)
	if hub.UniqueOnlineCount() != 0 {
		t.Fatal("a pre-claim token opened a new socket")
	}
	e.openSocket(t, srv.URL, e.accessToken(uid, 1))
	if !waitCount(t, hub, 1) {
		t.Fatal("a current token must still connect")
	}
}

// revokeAfterFirstVerify delegates to the real verifier, then bumps the epoch once, as a
// claim committing between the connect-time check and Join would.
type revokeAfterFirstVerify struct {
	inner notification_module.TokenVerifier
	once  sync.Once
	bump  func()
}

func (r *revokeAfterFirstVerify) VerifyToken(token string) (uint, error) {
	id, err := r.inner.VerifyToken(token)
	if err == nil {
		r.once.Do(r.bump)
	}
	return id, err
}

// N2b: a revoke racing the handshake must not leave a live socket behind.
func TestN2b_RevokeBetweenVerifyAndJoinDisconnects(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")
	v := &revokeAfterFirstVerify{
		inner: middleware.NewTokenVerifier(e.cfg.JWT.AccessSecret, e.db),
		bump:  func() { e.db.Exec(`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, uid) },
	}
	hub := notification_module.NewHub(v, []string{"http://app.test"})
	srv := httptest.NewServer(hub.Handler())
	t.Cleanup(srv.Close)

	e.openSocket(t, srv.URL, e.accessToken(uid, 0))
	if !waitCount(t, hub, 0) {
		t.Fatalf("socket revoked mid-handshake is still connected (online=%d)", hub.UniqueOnlineCount())
	}
	time.Sleep(300 * time.Millisecond)
	if hub.UniqueOnlineCount() != 0 {
		t.Fatal("socket came back")
	}
}
