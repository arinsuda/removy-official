package auth_module

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
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
		cfg:            &config.Config{},
		mailer:         mailer.New(config.BrevoConfig{}),
		findByEmail:    func(string) (*user_module.User, error) { return withID(1, "a@b.test"), nil },
		findByUsername: noUser,
	}
	_, err := svc.Register(RegisterRequest{Username: "dupuser", Email: "a@b.test", Password: "Vq7mTx2LpZ9k", ConfirmPassword: "Vq7mTx2LpZ9k"})
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("err = %v, want ErrEmailTaken", err)
	}
	if calls != 1 {
		t.Fatalf("hasher calls = %d, want 1 (equal work for duplicate email)", calls)
	}
}

func noUser(string) (*user_module.User, error) { return nil, ErrUserNotFound }

const testPassword = "Vq7mTx2LpZ9k"

func TestRegisterTakenUsernameWinsBeforeEmailLookup(t *testing.T) {
	emailLooked := false
	svc := &Service{
		cfg:    &config.Config{},
		mailer: mailer.New(config.BrevoConfig{}),
		findByUsername: func(string) (*user_module.User, error) {
			return withID(2, ""), nil
		},
		findByEmail: func(string) (*user_module.User, error) {
			emailLooked = true
			return withID(1, ""), nil
		},
	}
	_, err := svc.Register(RegisterRequest{Username: "johndoe", Email: "a@b.test", Password: testPassword, ConfirmPassword: testPassword})
	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("err = %v, want ErrUsernameTaken", err)
	}
	if emailLooked {
		t.Fatal("email must not be looked up when the username is taken")
	}
}

func TestShouldNotifyThrottlesPerRecipient(t *testing.T) {
	var s Service
	if !s.shouldNotify(1) || s.shouldNotify(1) {
		t.Fatal("second notice within the interval must be suppressed")
	}
	if !s.shouldNotify(2) {
		t.Fatal("other recipient unaffected")
	}
}

func TestLoginWrongPasswordOnLegacyHashBurns(t *testing.T) {
	burns := 0
	orig := burnPassword
	burnPassword = func(string) bool { burns++; return false }
	defer func() { burnPassword = orig }()

	h, _ := bcrypt.GenerateFromPassword([]byte("other-password-1"), bcrypt.MinCost)
	hs := string(h)
	svc := &Service{findByEmail: func(string) (*user_module.User, error) {
		return &user_module.User{IsActive: true, Password: &hs}, nil
	}}
	_, _, err := svc.Login(LoginRequest{Identifier: "a@b.test", Password: testPassword}, "", "")
	if !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("err = %v, want ErrWrongPassword", err)
	}
	if burns != 1 {
		t.Fatalf("burn calls = %d, want 1 for a low-cost hash", burns)
	}
}

func withID(id uint, email string) *user_module.User {
	u := &user_module.User{Email: email}
	u.ID = id
	return u
}
