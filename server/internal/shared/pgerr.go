package shared

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

const PgUniqueViolation = "23505"

func IsPgUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == PgUniqueViolation
	}
	return false
}

func IsPgUniqueViolationOn(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == PgUniqueViolation && pgErr.ConstraintName == constraint
	}
	return false
}

const (
	PgQueryCanceled = "57014" // statement_timeout
	PgLockNotAvail  = "55P03" // lock_timeout
	PgDeadlock      = "40P01"
)

// IsPgRetryable reports a statement timeout, lock timeout or deadlock: transient
// failures where a later attempt can succeed.
func IsPgRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	switch pgErr.Code {
	case PgQueryCanceled, PgLockNotAvail, PgDeadlock:
		return true
	}
	return false
}
