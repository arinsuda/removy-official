//go:build integration

package integration

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// seedUser inserts an active user. password "" stores NULL; verified sets verified_email_at.
func seedUser(t *testing.T, db *gorm.DB, email string, verified bool, password string) uint {
	t.Helper()
	var pw *string
	if password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(password), 4)
		if err != nil {
			t.Fatal(err)
		}
		s := string(h)
		pw = &s
	}
	var ver *time.Time
	if verified {
		now := time.Now().UTC()
		ver = &now
	}
	username := "u" + randomSuffix()
	var id uint
	err := db.Raw(`INSERT INTO users (created_at, updated_at, username, email, password, verified_email_at, role_id, is_active)
		VALUES (now(), now(), ?, ?, ?, ?, 2, true) RETURNING id`, username, email, pw, ver).Scan(&id).Error
	if err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return id
}
