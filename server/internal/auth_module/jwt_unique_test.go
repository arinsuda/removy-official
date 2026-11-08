package auth_module

import (
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
)

func TestIssueGivesUniqueRefreshTokensWithinSameSecond(t *testing.T) {
	j := &jwtManager{cfg: config.JWTConfig{
		AccessSecret: "access-secret", RefreshSecret: "refresh-secret",
		AccessTTL: time.Minute, RefreshTTL: time.Hour,
	}}
	first, err := j.Issue(1, "user", 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := j.Issue(1, "user", 0)
	if err != nil {
		t.Fatal(err)
	}
	if first.RefreshToken == second.RefreshToken {
		t.Fatal("refresh tokens must differ")
	}
}
