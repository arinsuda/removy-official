//go:build integration

package integration

import (
	"bytes"
	"database/sql"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Bound values (emails, tokens, hashes) must never reach the logs, not even on a failing query.
func TestGormLoggerNeverPrintsBoundValues(t *testing.T) {
	db := newTestDB(t)
	var buf bytes.Buffer
	l := database.NewGormLogger(log.New(&buf, "", 0))
	probe, err := gorm.Open(postgres.New(postgres.Config{Conn: mustSQLDB(t, db)}), &gorm.Config{Logger: l, NowFunc: func() time.Time { return time.Now().UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	const secret = "s3cret-token-value"
	if err := probe.Exec(`SELECT no_such_column FROM users WHERE email = ? AND password = ?`, "leak@x.test", secret).Error; err == nil {
		t.Fatal("expected the probe query to fail")
	}
	out := buf.String()
	if out == "" {
		t.Fatal("a failing query should still be logged")
	}
	if strings.Contains(out, secret) || strings.Contains(out, "leak@x.test") {
		t.Fatalf("bound values leaked into the log: %s", out)
	}
	// and successful queries are not logged at all at Warn
	buf.Reset()
	var n int
	probe.Raw(`SELECT 1 WHERE ? = ?`, "leak@x.test", "leak@x.test").Scan(&n)
	if buf.Len() != 0 {
		t.Fatalf("fast successful query was logged: %s", buf.String())
	}
}

func mustSQLDB(t *testing.T, db *gorm.DB) *sql.DB {
	t.Helper()
	s, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	return s
}
