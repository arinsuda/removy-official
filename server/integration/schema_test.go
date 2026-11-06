//go:build integration

package integration

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/arinsuda/movie-hub/database"
)

func TestHostGuardRefusesRemoteHosts(t *testing.T) {
	for _, raw := range []string{
		"postgres://u:p@db.neon.tech/neondb",
		"postgres://u:p@10.0.0.5:5432/x",
		"postgres://u:p@localhost.evil.com/x",
	} {
		if _, err := localHostOnly(raw); err == nil {
			t.Errorf("host guard accepted %q", raw)
		}
	}
	for _, raw := range []string{"postgres://u:p@localhost:5433/x", "postgres://u:p@127.0.0.1/x"} {
		if _, err := localHostOnly(raw); err != nil {
			t.Errorf("host guard rejected %q: %v", raw, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

// T10: uq_user_provider lives only in raw SQL, so only a real Postgres proves it.
func TestUserProviderUniqueConstraint(t *testing.T) {
	db := newTestDB(t)
	uid := seedUser(t, db, "a@x.test", true, "pw-aaaaaaaa")

	ins := `INSERT INTO user_identities (user_id, provider, provider_subject, provider_email) VALUES (?, 'google', ?, 'a@x.test')`
	if err := db.Exec(ins, uid, "sub-1").Error; err != nil {
		t.Fatalf("first identity: %v", err)
	}
	err := db.Exec(ins, uid, "sub-2").Error
	if err == nil || !strings.Contains(err.Error(), "uq_user_provider") {
		t.Fatalf("expected uq_user_provider violation, got %v", err)
	}
	other := seedUser(t, db, "b@x.test", true, "pw-bbbbbbbb")
	if err := db.Exec(ins, other, "sub-1").Error; err == nil || !strings.Contains(err.Error(), "uq_provider_subject") {
		t.Fatalf("expected uq_provider_subject violation, got %v", err)
	}
}

// Two concurrent inserts of the same identity: exactly one wins.
func TestConcurrentDoubleIdentityInsert(t *testing.T) {
	db := newTestDB(t)
	uid := seedUser(t, db, "a@x.test", true, "pw-aaaaaaaa")

	var wins, losses atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			err := db.Exec(`INSERT INTO user_identities (user_id, provider, provider_subject, provider_email) VALUES (?, 'google', ?, 'a@x.test')`,
				uid, "sub-"+string(rune('a'+i))).Error
			if err == nil {
				wins.Add(1)
			} else {
				losses.Add(1)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if wins.Load() != 1 || losses.Load() != 7 {
		t.Fatalf("wins=%d losses=%d, want 1/7", wins.Load(), losses.Load())
	}
}
