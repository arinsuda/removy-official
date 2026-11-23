package auth_module

import (
	"errors"
	"testing"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/user_module"
)

func TestRegisterDuplicateEmailStillHashes(t *testing.T) {
	calls := 0
	orig := hashSecret
	hashSecret = func(string) ([]byte, error) { calls++; return []byte("x"), nil }
	defer func() { hashSecret = orig }()

	svc := &Service{
		cfg:         &config.Config{},
		mailer:      mailer.New(config.BrevoConfig{}),
		findByEmail: func(string) (*user_module.User, error) { return &user_module.User{Email: "a@b.test"}, nil },
	}
	_, err := svc.Register(RegisterRequest{Username: "dupuser", Email: "a@b.test", Password: "Vq7mTx2LpZ9k", ConfirmPassword: "Vq7mTx2LpZ9k"})
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("err = %v, want ErrEmailTaken", err)
	}
	if calls != 1 {
		t.Fatalf("hasher calls = %d, want 1 (equal work for duplicate email)", calls)
	}
}
