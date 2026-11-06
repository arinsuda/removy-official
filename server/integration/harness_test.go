//go:build integration

// Package integration runs account-security scenarios against a real Postgres.
//
// Run: TEST_DATABASE_URL=postgres://user:pass@localhost:5432/postgres \
//
//	go test -tags integration ./integration/...
//
// The URL only identifies the server; every test creates its own throwaway
// database (popcorn_it_*) and drops it afterwards. Hosts other than
// localhost/127.0.0.1/::1 are refused, so this can never touch a hosted DB.
package integration

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const dbNamePrefix = "popcorn_it_"

// localHostOnly returns an error unless the URL points at the local machine.
func localHostOnly(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("TEST_DATABASE_URL is not a valid URL")
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return u, nil
	}
	return nil, fmt.Errorf("refusing to run: TEST_DATABASE_URL host %q is not localhost/127.0.0.1", u.Hostname())
}

func openDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
}

func randomSuffix() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// newTestDB creates a throwaway database, migrates it with the production
// migration entry point and drops it on cleanup.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if raw == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	base, err := localHostOnly(raw)
	if err != nil {
		t.Fatal(err)
	}

	admin, err := openDB(base.String())
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	name := dbNamePrefix + randomSuffix()
	if err := admin.Exec(`CREATE DATABASE ` + name).Error; err != nil {
		t.Fatalf("create database: %v", err)
	}

	dbURL := *base
	dbURL.Path = "/" + name
	db, err := openDB(dbURL.String())
	if err != nil {
		_ = admin.Exec(`DROP DATABASE IF EXISTS ` + name).Error
		t.Fatalf("connect test db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(30)

	t.Cleanup(func() {
		_ = sqlDB.Close()
		if strings.HasPrefix(name, dbNamePrefix) {
			_ = admin.Exec(`DROP DATABASE IF EXISTS ` + name + ` WITH (FORCE)`).Error
		}
		if adminSQL, err := admin.DB(); err == nil {
			_ = adminSQL.Close()
		}
	})

	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func testConfig() *config.Config {
	return &config.Config{
		AppBaseURL: "http://app.test",
		DB:         config.DBConfig{MigrationLockTimeoutMs: 5000, MigrationStatementTimeoutMs: 60000},
		JWT: config.JWTConfig{
			AccessSecret:  "test-access-secret",
			RefreshSecret: "test-refresh-secret",
			AccessTTL:     15 * time.Minute,
			RefreshTTL:    time.Hour,
		},
		Cookie: config.CookieConfig{SameSite: "Lax"},
	}
}
