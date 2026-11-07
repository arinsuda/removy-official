package identity_policy

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

// ErrSessionStale means the account's session epoch moved on (or the account
// stopped being active/verified) since the caller's token was issued.
var ErrSessionStale = errors.New("identity: session is no longer valid")

// ErrUnverified means the account has no proven email, so it may not hold a session.
var ErrUnverified = errors.New("identity: email not verified")

// SessionState is what a request needs to know about its user.
type SessionState struct {
	Version  int
	Active   bool
	Verified bool
}

// ReadSession is one primary-key read, no lock. found is false when the user
// does not exist or is soft-deleted.
func ReadSession(db *gorm.DB, userID uint) (st SessionState, found bool, err error) {
	return readSession(db, userID, "")
}

// LockSession reads the same state under a row lock inside tx. FOR SHARE lets
// concurrent readers through but waits for a running ProveEmail; forUpdate
// serializes writers (password change, email change) against it.
func LockSession(tx *gorm.DB, userID uint, forUpdate bool) (SessionState, bool, error) {
	if forUpdate {
		return readSession(tx, userID, " FOR UPDATE")
	}
	return readSession(tx, userID, " FOR SHARE")
}

func readSession(db *gorm.DB, userID uint, lock string) (SessionState, bool, error) {
	var row struct {
		TokenVersion int
		IsActive     bool
		Verified     bool
	}
	res := db.Raw(`SELECT token_version, is_active, (verified_email_at IS NOT NULL) AS verified
		FROM users WHERE id = ? AND deleted_at IS NULL`+lock, userID).Scan(&row)
	if res.Error != nil {
		return SessionState{}, false, res.Error
	}
	if res.RowsAffected == 0 {
		return SessionState{}, false, nil
	}
	return SessionState{Version: row.TokenVersion, Active: row.IsActive, Verified: row.Verified}, true, nil
}

// EnsureSession locks the user and confirms the caller's token still matches
// the account. Sensitive writes call it first, inside their transaction, so a
// request that passed RequireAuth just before a claim cannot write after it.
func EnsureSession(tx *gorm.DB, userID uint, claimedVersion int, forUpdate bool) error {
	st, found, err := LockSession(tx, userID, forUpdate)
	if err != nil {
		return err
	}
	if !found || !st.Active || !st.Verified || st.Version != claimedVersion {
		return ErrSessionStale
	}
	return nil
}

var (
	hookMu          sync.RWMutex
	sessionsRevoked func(userID uint)
)

// SetSessionsRevokedHook registers what runs after a transaction that revoked
// a user's sessions has committed (the realtime hub drops that user's sockets).
func SetSessionsRevokedHook(f func(userID uint)) {
	hookMu.Lock()
	sessionsRevoked = f
	hookMu.Unlock()
}

// Publish must be called after the transaction that produced o has committed.
func (o Outcome) Publish(userID uint) {
	if !o.SessionsRevoked {
		return
	}
	hookMu.RLock()
	f := sessionsRevoked
	hookMu.RUnlock()
	if f != nil {
		f(userID)
	}
}
