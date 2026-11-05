package link_module

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHMACRoundTripAndRotation(t *testing.T) {
	body := []byte(`{"token":"abc"}`)
	now := time.Now()
	ts := strconv.FormatInt(now.Unix(), 10)
	sig := Sign("old-secret", ts, body)

	if err := VerifySignature([]string{"new-secret", "old-secret"}, ts, body, sig, now); err != nil {
		t.Fatalf("rotation secret should verify: %v", err)
	}
	if VerifySignature([]string{"new-secret"}, ts, body, sig, now) == nil {
		t.Fatal("wrong secret must fail")
	}
	if VerifySignature([]string{"old-secret"}, ts, []byte(`{"token":"x"}`), sig, now) == nil {
		t.Fatal("tampered body must fail")
	}
	if VerifySignature([]string{"old-secret"}, ts, body, "zz", now) == nil {
		t.Fatal("bad hex must fail")
	}
	if VerifySignature([]string{"old-secret"}, ts, body, sig, now.Add(31*time.Second)) == nil {
		t.Fatal("stale timestamp must fail")
	}
	if VerifySignature([]string{"old-secret"}, "nope", body, sig, now) == nil {
		t.Fatal("bad timestamp must fail")
	}
}

func TestDecide(t *testing.T) {
	cases := []struct {
		linked, exists, lv, rv bool
		want                   Decision
	}{
		{true, true, true, true, DecisionLoginLinked},
		{false, false, false, false, DecisionProvision},
		{false, true, true, true, DecisionAutoLink},
		{false, true, true, false, DecisionConfirm},
		{false, true, false, true, DecisionConfirm},
		{false, true, false, false, DecisionConfirm},
	}
	for i, c := range cases {
		if got := Decide(c.linked, c.exists, c.lv, c.rv); got != c.want {
			t.Errorf("case %d: got %v want %v", i, got, c.want)
		}
	}
}

func TestConfirmMethods(t *testing.T) {
	if m := ConfirmMethods(true, true); len(m) != 2 || m[0] != MethodPasswd {
		t.Errorf("unexpected %v", m)
	}
	if m := ConfirmMethods(true, false); len(m) != 1 || m[0] != MethodEmail {
		t.Errorf("unverified local must only allow email code: %v", m)
	}
	if m := ConfirmMethods(false, true); len(m) != 1 || m[0] != MethodEmail {
		t.Errorf("no password must only allow email code: %v", m)
	}
}

func TestGenerateUsername(t *testing.T) {
	none := func(string) bool { return false }
	if got := GenerateUsername("John.Doe+x@example.com", none); got != "johndoex" {
		t.Errorf("got %q", got)
	}
	if got := GenerateUsername("a@x.com", none); got != "auser" {
		t.Errorf("short local part: got %q", got)
	}
	if got := GenerateUsername("!!!@x.com", none); got != "user" {
		t.Errorf("empty local part: got %q", got)
	}
	long := strings.Repeat("a", 80) + "@x.com"
	if got := GenerateUsername(long, none); len(got) > 30 {
		t.Errorf("too long: %d", len(got))
	}
	taken := map[string]bool{"john": true}
	got := GenerateUsername("john@x.com", func(s string) bool { return taken[s] })
	if got == "john" || !strings.HasPrefix(got, "john") || len(got) > 50 {
		t.Errorf("collision: got %q", got)
	}
	always := func(s string) bool { return s != "" && len(s) < 20 }
	if got := GenerateUsername("john@x.com", always); len(got) > 50 {
		t.Errorf("fallback too long %q", got)
	}
}
