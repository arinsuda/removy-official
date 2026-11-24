package authpolicy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"
)

// These hashes are pinned in LINKING_SPEC.md ("Auth parity v2") and in the sibling repo's test.
// Change them only together, in both repos and the SPEC.
const (
	conformanceSHA256 = "176f03f80b6cc688aa4f5eb02af549eea28a2ebf0d6ba6077c292398f59576ab"
	blocklistSHA256   = "5ede5e13198f18de4214c5e6dc524ef0568143e19362c303f27e710e81e81249"
)

func sha(t *testing.T, b []byte) string {
	t.Helper()
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

type confCase struct {
	Input      string `json:"input"`
	Valid      bool   `json:"valid"`
	Normalized string `json:"normalized"`
}

type confPassword struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Code     string `json:"code"`
}

func TestSharedFilesMatchSpecHashes(t *testing.T) {
	raw, err := os.ReadFile("testdata/auth_conformance.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := sha(t, raw); got != conformanceSHA256 {
		t.Fatalf("auth_conformance.json sha256 = %s, spec pins %s", got, conformanceSHA256)
	}
	if got := sha(t, []byte(blocklistRaw)); got != blocklistSHA256 {
		t.Fatalf("common_passwords.txt sha256 = %s, spec pins %s", got, blocklistSHA256)
	}
}

func loadConformance(t *testing.T) (c struct {
	Email    []confCase     `json:"email"`
	Password []confPassword `json:"password"`
	Username []confCase     `json:"username"`
}) {
	t.Helper()
	raw, err := os.ReadFile("testdata/auth_conformance.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestConformanceEmail(t *testing.T) {
	for _, tc := range loadConformance(t).Email {
		got, err := ValidateEmail(tc.Input)
		if (err == nil) != tc.Valid {
			t.Errorf("email %q: valid=%v want %v (%v)", tc.Input, err == nil, tc.Valid, err)
			continue
		}
		if tc.Valid && got != tc.Normalized {
			t.Errorf("email %q normalized %q want %q", tc.Input, got, tc.Normalized)
		}
		if !tc.Valid && !errors.Is(err, ErrInvalidEmail) {
			t.Errorf("email %q: wrong error %v", tc.Input, err)
		}
	}
}

func TestConformanceUsername(t *testing.T) {
	for _, tc := range loadConformance(t).Username {
		got, err := ValidateUsername(tc.Input)
		if (err == nil) != tc.Valid {
			t.Errorf("username %q: valid=%v want %v", tc.Input, err == nil, tc.Valid)
			continue
		}
		if tc.Valid && got != tc.Normalized {
			t.Errorf("username %q normalized %q want %q", tc.Input, got, tc.Normalized)
		}
	}
}

func TestConformancePassword(t *testing.T) {
	for _, tc := range loadConformance(t).Password {
		err := ValidatePassword(tc.Password, tc.Email, tc.Username)
		got := ""
		if err != nil {
			var pe *PolicyError
			if !errors.As(err, &pe) {
				t.Fatalf("password %q: non-policy error %v", tc.Password, err)
			}
			got = pe.Code
		}
		if got != tc.Code {
			t.Errorf("password %q (email=%q user=%q): code %q want %q", tc.Password, tc.Email, tc.Username, got, tc.Code)
		}
	}
}

func TestNeedsRehashAndDummy(t *testing.T) {
	h, err := HashPassword("x7Kp2mQz-test")
	if err != nil {
		t.Fatal(err)
	}
	if NeedsRehash(h) {
		t.Error("cost-12 hash should not need rehash")
	}
	if !NeedsRehash("$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234") {
		t.Error("cost-10 hash should need rehash")
	}
	if NeedsRehash("not-a-hash") {
		t.Error("garbage must not trigger rehash")
	}
	if BurnPassword("anything") {
		t.Error("BurnPassword must be false")
	}
}
