package shared

import "github.com/gofiber/fiber/v3"

type ErrorCode string

const (
	ErrorCodeDuplicateReview   ErrorCode = "DUPLICATE_REVIEW"
	ErrorCodeInvalidRating     ErrorCode = "INVALID_RATING"
	ErrorCodeInvalidVisibility ErrorCode = "INVALID_VISIBILITY"
	ErrorCodeWatchLogNotFound  ErrorCode = "WATCH_LOG_NOT_FOUND"
	ErrorCodeInvalidDate       ErrorCode = "INVALID_DATE"
	ErrorCodeFutureDate        ErrorCode = "FUTURE_DATE"
	ErrorCodeInvalidInput      ErrorCode = "INVALID_INPUT"
	ErrorCodeInvalidMediaType  ErrorCode = "INVALID_MEDIA_TYPE"
	ErrorCodeListNotFound      ErrorCode = "LIST_NOT_FOUND"
	ErrorCodeListLimit         ErrorCode = "LIST_LIMIT"
	ErrorCodeListFull          ErrorCode = "LIST_FULL"
	ErrorCodeItemExists        ErrorCode = "ITEM_EXISTS"
	ErrorCodeItemNotFound      ErrorCode = "ITEM_NOT_FOUND"
	ErrorCodeOrderMismatch     ErrorCode = "ORDER_MISMATCH"
	ErrorCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrorCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrorCodeUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrorCodeReviewNotFound    ErrorCode = "REVIEW_NOT_FOUND"
	ErrorCodeInvalidRole       ErrorCode = "INVALID_ROLE"
	ErrorCodeSelfDeactivation  ErrorCode = "SELF_DEACTIVATION"
	ErrorCodeSelfDeletion      ErrorCode = "SELF_DELETION"
	ErrorCodeAlreadyInStatus   ErrorCode = "ALREADY_IN_STATUS"
	ErrorCodeFinalAdmin        ErrorCode = "FINAL_ADMIN"
	ErrorCodeInactiveUserRole  ErrorCode = "INACTIVE_USER_ROLE_CHANGE"
	ErrorCodeInternal          ErrorCode = "INTERNAL_ERROR"
	ErrorCodeInvalidID         ErrorCode = "INVALID_ID"
	ErrorCodeInvalidQuery      ErrorCode = "INVALID_QUERY"
	ErrorCodeInvalidGenre      ErrorCode = "INVALID_GENRE_FILTER"
	ErrorCodeInvalidSearch     ErrorCode = "INVALID_SEARCH_FILTER"
	ErrorCodeInvalidScope      ErrorCode = "INVALID_FEED_SCOPE"
	ErrorCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrorCodeUpstream          ErrorCode = "UPSTREAM_ERROR"
)

type APIError struct {
	Code  ErrorCode `json:"code"`
	Error string    `json:"error"`
}

func WriteAPIError(c fiber.Ctx, status int, code ErrorCode, message string) error {
	return c.Status(status).JSON(APIError{Code: code, Error: message})
}
