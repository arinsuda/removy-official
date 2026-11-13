package shared

import (
	"github.com/arinsuda/movie-hub/internal/shared/apierr"
	"github.com/gofiber/fiber/v3"
)

// Thin aliases over apierr so existing callers keep compiling.
type (
	ErrorCode = apierr.ErrorCode
	APIError  = apierr.APIError
)

const (
	ErrorCodeDuplicateReview   = apierr.ErrorCodeDuplicateReview
	ErrorCodeInvalidRating     = apierr.ErrorCodeInvalidRating
	ErrorCodeInvalidVisibility = apierr.ErrorCodeInvalidVisibility
	ErrorCodeWatchLogNotFound  = apierr.ErrorCodeWatchLogNotFound
	ErrorCodeInvalidDate       = apierr.ErrorCodeInvalidDate
	ErrorCodeFutureDate        = apierr.ErrorCodeFutureDate
	ErrorCodeInvalidInput      = apierr.ErrorCodeInvalidInput
	ErrorCodeInvalidMediaType  = apierr.ErrorCodeInvalidMediaType
	ErrorCodeListNotFound      = apierr.ErrorCodeListNotFound
	ErrorCodeListLimit         = apierr.ErrorCodeListLimit
	ErrorCodeListFull          = apierr.ErrorCodeListFull
	ErrorCodeItemExists        = apierr.ErrorCodeItemExists
	ErrorCodeItemNotFound      = apierr.ErrorCodeItemNotFound
	ErrorCodeOrderMismatch     = apierr.ErrorCodeOrderMismatch
	ErrorCodeForbidden         = apierr.ErrorCodeForbidden
	ErrorCodeUnauthorized      = apierr.ErrorCodeUnauthorized
	ErrorCodeUserNotFound      = apierr.ErrorCodeUserNotFound
	ErrorCodeReviewNotFound    = apierr.ErrorCodeReviewNotFound
	ErrorCodeInvalidRole       = apierr.ErrorCodeInvalidRole
	ErrorCodeSelfDeactivation  = apierr.ErrorCodeSelfDeactivation
	ErrorCodeSelfDeletion      = apierr.ErrorCodeSelfDeletion
	ErrorCodeAlreadyInStatus   = apierr.ErrorCodeAlreadyInStatus
	ErrorCodeFinalAdmin        = apierr.ErrorCodeFinalAdmin
	ErrorCodeSelfErasure       = apierr.ErrorCodeSelfErasure
	ErrorCodeUserNotRemoved    = apierr.ErrorCodeUserNotRemoved
	ErrorCodeConfirmMismatch   = apierr.ErrorCodeConfirmMismatch
	ErrorCodeInactiveUserRole  = apierr.ErrorCodeInactiveUserRole
	ErrorCodeInternal          = apierr.ErrorCodeInternal
	ErrorCodeInvalidID         = apierr.ErrorCodeInvalidID
	ErrorCodeInvalidQuery      = apierr.ErrorCodeInvalidQuery
	ErrorCodeInvalidGenre      = apierr.ErrorCodeInvalidGenre
	ErrorCodeInvalidSearch     = apierr.ErrorCodeInvalidSearch
	ErrorCodeInvalidScope      = apierr.ErrorCodeInvalidScope
	ErrorCodeNotFound          = apierr.ErrorCodeNotFound
	ErrorCodeUpstream          = apierr.ErrorCodeUpstream
	ErrorCodeWrongPassword     = apierr.ErrorCodeWrongPassword
	ErrorCodePasswordNotSet    = apierr.ErrorCodePasswordNotSet
	ErrorCodeInvalidBody       = apierr.ErrorCodeInvalidBody
	ErrorCodeConfirmUsername   = apierr.ErrorCodeConfirmUsername
	ErrorCodeEmailInvalid      = apierr.ErrorCodeEmailInvalid
	ErrorCodeEmailSame         = apierr.ErrorCodeEmailSame
	ErrorCodeEmailInUse        = apierr.ErrorCodeEmailInUse
)

func WriteAPIError(c fiber.Ctx, status int, code ErrorCode, message string) error {
	return apierr.WriteAPIError(c, status, code, message)
}
