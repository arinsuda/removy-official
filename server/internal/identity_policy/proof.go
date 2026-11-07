// Package identity_policy is the single place where "this email is proven"
// is recorded. Proof of an address does not prove the person: an unverified
// account may carry a password, Google identity or link planted by whoever
// registered the address first. ProveEmail therefore wipes every planted
// credential unless the caller also proves the current password.
//
// Raw SQL on purpose: auth_module imports user_module, so this package must
// not import either of them.
package identity_policy

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Kind says how the address was proven. It never changes what is wiped; it
// only decides which extra inputs are meaningful.
type Kind string

const (
	KindVerifyLink    Kind = "verify_link"    // link mailed to the address (may carry the current password)
	KindGoogle        Kind = "google"         // Google asserted the address (authoritative claim)
	KindPasswordReset Kind = "password_reset" // reset link mailed to the address (sets a new password)
	KindLinkCode      Kind = "link_code"      // PLANZI link code mailed to the address (proof only)
)

var (
	ErrUserNotFound   = errors.New("identity: user not found")
	ErrUserInactive   = errors.New("identity: user inactive")
	ErrEmailMismatch  = errors.New("identity: proven email is not the current email")
	ErrMissingNewHash = errors.New("identity: password reset needs a new password hash")
	// ErrWrongPassword: a non-empty current password was presented and does not match.
	// Nothing was changed; the caller rolls back and lets the person retry or skip.
	ErrWrongPassword = errors.New("identity: presented password does not match")
)

// Proof describes one successful email proof.
type Proof struct {
	UserID uint
	Email  string // the address that was proven
	Kind   Kind
	// PresentedPassword is the current password typed by the person proving the
	// address. A match means the same person registered the account; a non-empty
	// mismatch is ErrWrongPassword (KindVerifyLink); nil or empty means "unknown".
	PresentedPassword *string
	// NewPasswordHash is required for KindPasswordReset only.
	NewPasswordHash *string
}

// Outcome reports what ProveEmail did.
type Outcome struct {
	AlreadyVerified  bool // nothing was unproven; no credentials touched (except reset)
	SamePerson       bool // unverified account, current password matched: everything kept
	CredentialsWiped bool // planted credentials removed, session epoch bumped
	PasswordReplaced bool // a reset stored the new password
	SessionsRevoked  bool // token_version bumped; call Publish after commit
}

type lockedUser struct {
	ID              uint
	Email           string
	Password        *string
	VerifiedEmailAt *time.Time
	IsActive        bool
}

// ProveEmail runs inside the caller's transaction. It locks the user row first,
// so concurrent proofs, refreshes and sensitive writes serialize on it.
func ProveEmail(tx *gorm.DB, p Proof) (Outcome, error) {
	var out Outcome
	if p.Kind == KindPasswordReset && (p.NewPasswordHash == nil || *p.NewPasswordHash == "") {
		return out, ErrMissingNewHash
	}

	var u lockedUser
	res := tx.Raw(`SELECT id, email, password, verified_email_at, is_active
		FROM users WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, p.UserID).Scan(&u)
	if res.Error != nil {
		return out, res.Error
	}
	if res.RowsAffected == 0 {
		return out, ErrUserNotFound
	}
	if !u.IsActive {
		return out, ErrUserInactive
	}
	if !sameEmail(u.Email, p.Email) {
		return out, ErrEmailMismatch
	}

	if u.VerifiedEmailAt != nil {
		out.AlreadyVerified = true
		if p.Kind == KindPasswordReset {
			// Verified owner resets: replace the password, end every session and
			// pending token, keep identities and links.
			if err := tx.Exec(`UPDATE users SET password = ?, token_version = token_version + 1, updated_at = now() WHERE id = ?`,
				*p.NewPasswordHash, u.ID).Error; err != nil {
				return out, err
			}
			out.PasswordReplaced = true
			out.SessionsRevoked = true
			return out, deleteAll(tx, u.ID, resetScopeTables)
		}
		return out, nil
	}

	if p.Kind == KindVerifyLink && samePerson(u.Password, p.PresentedPassword) {
		out.SamePerson = true
		err := tx.Exec(`UPDATE users SET verified_email_at = now(), updated_at = now() WHERE id = ?`, u.ID).Error
		if err != nil {
			return out, err
		}
		return out, tx.Exec(`DELETE FROM email_verifications WHERE user_id = ?`, u.ID).Error
	}

	// A typed password that does not match is a typo, not consent to wipe: only an
	// absent password (the caller's explicit "I don't know it") reaches the wipe below.
	if p.Kind == KindVerifyLink && p.PresentedPassword != nil && *p.PresentedPassword != "" && u.Password != nil && *u.Password != "" {
		return out, ErrWrongPassword
	}

	// Unverified and the person is not proven: whatever is on the account was
	// put there by someone who never owned the mailbox.
	var newPassword *string
	if p.Kind == KindPasswordReset {
		newPassword = p.NewPasswordHash
		out.PasswordReplaced = true
	}
	if err := tx.Exec(`UPDATE users SET password = ?, verified_email_at = now(), identity_reset_at = now(),
		token_version = token_version + 1, updated_at = now() WHERE id = ?`, newPassword, u.ID).Error; err != nil {
		return out, err
	}
	out.CredentialsWiped = true
	out.SessionsRevoked = true
	return out, deleteAll(tx, u.ID, wipeTables)
}

type userScoped struct{ table, column string }

// Everything a squatter could have planted, plus every live session.
var wipeTables = []userScoped{
	{"refresh_tokens", "user_id"},
	{"account_links", "local_user_id"},
	{"link_pending", "local_user_id"},
	{"link_handoff_tokens", "user_id"},
	{"user_identities", "user_id"},
	{"password_reset_tokens", "user_id"},
	{"email_change_requests", "user_id"},
	{"email_verifications", "user_id"},
}

// A reset on a verified account keeps identities and links.
var resetScopeTables = []userScoped{
	{"refresh_tokens", "user_id"},
	{"link_pending", "local_user_id"},
	{"link_handoff_tokens", "user_id"},
	{"password_reset_tokens", "user_id"},
	{"email_verifications", "user_id"},
}

func deleteAll(tx *gorm.DB, userID uint, tables []userScoped) error {
	for _, t := range tables {
		if err := tx.Exec(`DELETE FROM `+t.table+` WHERE `+t.column+` = ?`, userID).Error; err != nil {
			return err
		}
	}
	return nil
}

func samePerson(stored, presented *string) bool {
	if stored == nil || *stored == "" || presented == nil || *presented == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(*stored), []byte(*presented)) == nil
}

func sameEmail(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// SameEmail is the comparison every email-bound token check must use.
func SameEmail(a, b string) bool { return sameEmail(a, b) }
