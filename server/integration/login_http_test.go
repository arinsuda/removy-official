//go:build integration

package integration

import (
	"reflect"
	"testing"
)

// A wrong password, an unknown user and a passwordless account must be
// indistinguishable over HTTP: same status, same body.
func TestLoginHTTP_FailuresAreIndistinguishable(t *testing.T) {
	e := newEnv(t)
	seedUser(t, e.db, "has-pw@x.test", true, "right-password1")
	seedUser(t, e.db, "no-pw@x.test", true, "")

	login := func(id, pw string) (int, map[string]any) {
		return e.do("POST", "/api/auth/login", map[string]string{"identifier": id, "password": pw}, "")
	}
	wrongCode, wrongBody := login("has-pw@x.test", "wrong-password1")
	if wrongCode != 401 {
		t.Fatalf("wrong password: want 401, got %d %v", wrongCode, wrongBody)
	}
	for name, c := range map[string][2]string{
		"unknown user": {"ghost@x.test", "wrong-password1"},
		"passwordless": {"no-pw@x.test", "wrong-password1"},
	} {
		code, body := login(c[0], c[1])
		if code != wrongCode || !reflect.DeepEqual(body, wrongBody) {
			t.Errorf("%s: got %d %v, want %d %v", name, code, body, wrongCode, wrongBody)
		}
	}
}
