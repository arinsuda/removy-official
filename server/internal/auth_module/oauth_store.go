package auth_module

import (
	"context"
	"crypto/rand"
	"crypto/sha256"

	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type OAuthTransactionIntent string

const (
	IntentLogin OAuthTransactionIntent = "login"
	IntentLink  OAuthTransactionIntent = "link"
)

var (
	ErrOAuthTransactionNotFound = errors.New("oauth transaction not found or expired")
	ErrOAuthTransactionConsumed = errors.New("oauth transaction already consumed")
)

type OAuthTransaction struct {
	State         string                 `json:"state"`
	StateHash     string                 `json:"state_hash"`
	Nonce         string                 `json:"nonce"`
	CodeVerifier  string                 `json:"code_verifier"`
	CodeChallenge string                 `json:"code_challenge"`
	Intent        OAuthTransactionIntent `json:"intent"`
	UserID        uint                   `json:"user_id,omitempty"`
	// UserVersion is the initiating session's epoch; the callback re-checks it under lock.
	UserVersion int       `json:"user_version,omitempty"`
	ReturnURL   string    `json:"return_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type OAuthTransactionStore interface {
	Create(ctx context.Context, tx *OAuthTransaction) error
	Consume(ctx context.Context, state string) (*OAuthTransaction, error)
	Delete(ctx context.Context, state string) error
}

func HashState(state string) string {
	h := sha256.Sum256([]byte(state))
	return hex.EncodeToString(h[:])
}

func GenerateCryptoString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func ComputePKCEChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

type MemoryOAuthTransactionStore struct {
	mu           sync.RWMutex
	transactions map[string]*OAuthTransaction
}

func NewMemoryOAuthTransactionStore() *MemoryOAuthTransactionStore {
	store := &MemoryOAuthTransactionStore{
		transactions: make(map[string]*OAuthTransaction),
	}
	go store.backgroundCleanup(5 * time.Minute)
	return store
}

func (s *MemoryOAuthTransactionStore) Create(ctx context.Context, tx *OAuthTransaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tx.StateHash == "" {
		tx.StateHash = HashState(tx.State)
	}
	key := "oauth:google:transaction:" + tx.StateHash
	s.transactions[key] = tx
	return nil
}

func (s *MemoryOAuthTransactionStore) Consume(ctx context.Context, state string) (*OAuthTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := HashState(state)
	key := "oauth:google:transaction:" + hash

	tx, ok := s.transactions[key]
	if !ok {
		return nil, ErrOAuthTransactionNotFound
	}

	delete(s.transactions, key)

	if time.Now().After(tx.ExpiresAt) {
		return nil, ErrOAuthTransactionNotFound
	}

	return tx, nil
}

func (s *MemoryOAuthTransactionStore) Delete(ctx context.Context, state string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := HashState(state)
	key := "oauth:google:transaction:" + hash
	delete(s.transactions, key)
	return nil
}

func (s *MemoryOAuthTransactionStore) backgroundCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.transactions {
			if now.After(v.ExpiresAt) {
				delete(s.transactions, k)
			}
		}
		s.mu.Unlock()
	}
}

func ValidateRelativeReturnURL(url string) string {
	if url == "" {
		return "/"
	}
	url = strings.TrimSpace(url)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "//") || strings.HasPrefix(url, "\\") {
		return "/"
	}
	if !strings.HasPrefix(url, "/") {
		return "/"
	}
	if url == "/login" || url == "/register" || url == "/q" || strings.HasPrefix(url, "/q") || strings.HasPrefix(url, "/login?") || strings.HasPrefix(url, "/register?") {
		return "/"
	}
	return url
}
