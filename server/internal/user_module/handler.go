package user_module

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/arinsuda/movie-hub/internal/shared/apierr"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetProfile(c fiber.Ctx) error {
	targetID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	claims := mw.GetClaims(c)

	profile, err := h.svc.GetProfile(targetID, claims.UserID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{"user": profile})
}

func (h *Handler) UpdateProfile(c fiber.Ctx) error {
	targetID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	claims := mw.GetClaims(c)

	req, err := parseUpdateProfileForm(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var avatarFile *multipartFile
	fh, formErr := c.FormFile("avatar")
	if formErr == nil && fh != nil {
		if err := validateImageFile(fh); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		f, err := fh.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot read avatar file"})
		}
		defer f.Close()
		avatarFile = &multipartFile{file: f, header: fh}
	}

	var profile *UserProfileResponse
	if avatarFile != nil {
		profile, err = h.svc.UpdateProfile(targetID, claims.UserID, req, avatarFile.file, avatarFile.header)
	} else {
		profile, err = h.svc.UpdateProfile(targetID, claims.UserID, req, nil, nil)
	}
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{"user": profile})
}

func (h *Handler) RequestEmailChange(c fiber.Ctx) error {
	targetID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	claims := mw.GetClaims(c)

	var req RequestEmailChangeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeInvalidBody, "error": msgInvalidBody})
	}

	if strings.TrimSpace(req.NewEmail) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new_email is required"})
	}

	if err := h.svc.RequestEmailChange(targetID, claims.UserID, req.NewEmail); err != nil {
		return handleEmailChangeError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "OTPs have been sent to your current and new email addresses",
	})
}

// DeleteUser lets a user remove their own account. It is a deactivation (is_active=false,
// sessions revoked), not the admin soft delete: reviews stay public and the email stays reserved.
func (h *Handler) DeleteUser(c fiber.Ctx) error {
	targetID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	claims := mw.GetClaims(c)

	var req DeleteAccountRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeInvalidBody, "error": msgInvalidBody})
		}
	}

	if err := h.svc.DeleteUser(targetID, claims.UserID, claims.TokenVersion, req); err != nil {
		return handleError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RequestErasure schedules the caller's permanent erasure. 202: the account is deactivated now
// and removed by the purge worker when erasure_due_at passes.
// Messages and header values used by requireXHRJSON.
const (
	msgUnsupportedMedia = "content type must be application/json"
	msgXHRRequired      = "missing X-Requested-With header"
	headerRequestedWith = "X-Requested-With"
	requestedWithXHR    = "XMLHttpRequest"
)

// requireXHRJSON blocks cross-site form posts on a destructive route. The JSON binder ignores
// Content-Type, so a text/plain form from another site would otherwise carry the SameSite=None
// cookie. A browser cannot add a custom header cross-site without a CORS preflight, which only
// the configured origins pass.
func requireXHRJSON(c fiber.Ctx) error {
	if ct := strings.ToLower(c.Get(fiber.HeaderContentType)); !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{"code": apierr.ErrorCodeUnsupportedMedia, "error": msgUnsupportedMedia})
	}
	if c.Get(headerRequestedWith) != requestedWithXHR {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"code": apierr.ErrorCodeXHRRequired, "error": msgXHRRequired})
	}
	return c.Next()
}

func (h *Handler) RequestErasure(c fiber.Ctx) error {
	claims := mw.GetClaims(c)

	var req DeleteAccountRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeInvalidBody, "error": msgInvalidBody})
		}
	}

	dueAt, err := h.svc.RequestErasure(claims.UserID, claims.TokenVersion, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"erasure_due_at": dueAt})
}

func (h *Handler) UpdateFavoriteGenres(c fiber.Ctx) error {
	targetID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	claims := mw.GetClaims(c)

	var req UpdateFavoriteGenresRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeInvalidBody, "error": msgInvalidBody})
	}

	profile, err := h.svc.UpdateFavoriteGenres(targetID, claims.UserID, req.FavoriteGenres)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{"user": profile})
}

func (h *Handler) VerifyEmailChange(c fiber.Ctx) error {
	targetID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	claims := mw.GetClaims(c)

	var req VerifyEmailChangeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeInvalidBody, "error": msgInvalidBody})
	}

	if req.OTP == "" || req.NewEmailOTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "otp and new_email_otp are required"})
	}

	profile, err := h.svc.VerifyEmailChange(targetID, claims.UserID, claims.TokenVersion, req.OTP, req.NewEmailOTP)
	if err != nil {
		return handleEmailChangeError(c, err)
	}

	return c.JSON(fiber.Map{"user": profile})
}

func (h *Handler) ChangePassword(c fiber.Ctx) error {
	targetID, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	claims := mw.GetClaims(c)

	var req ChangePasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeInvalidBody, "error": msgInvalidBody})
	}
	if req.OldPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "old_password is required"})
	}
	if req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new_password is required"})
	}
	if req.ConfirmPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "confirm_password is required"})
	}

	if err := h.svc.ChangePassword(targetID, claims.UserID, claims.TokenVersion, req); err != nil {
		return handlePasswordError(c, err)
	}

	// Every session was revoked by the change; give this device a fresh one so it stays signed in.
	resp := fiber.Map{"message": "password changed successfully"}
	if h.svc.reissuer != nil {
		access, refresh, err := h.svc.reissuer.ReissueSession(c, claims.UserID, claims.TokenVersion+1)
		if err != nil {
			log.Printf("ERROR reissue session after password change for user %d: %v", claims.UserID, err)
		} else {
			resp["access_token"] = access
			resp["refresh_token"] = refresh
		}
	}
	return c.JSON(resp)
}

func handlePasswordError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	case errors.Is(err, ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "incorrect current password"})
	case errors.Is(err, ErrSessionStale):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	case errors.Is(err, ErrPasswordNotSet):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error(), "code": apierr.ErrorCodePasswordNotSet})
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
}

func handleEmailChangeError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	case errors.Is(err, storage.ErrInvalidImage):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "avatar must be a JPEG, PNG, WebP or GIF image"})
	case errors.Is(err, ErrInvalidEmailFormat):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid email format"})
	case errors.Is(err, ErrEmailSameAsCurrent):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new email must be different from current email"})
	case errors.Is(err, ErrEmailAlreadyInUse):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already in use"})
	case errors.Is(err, ErrPendingEmailMissing):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "pending new email missing"})
	case errors.Is(err, ErrOTPNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no pending email change request"})
	case errors.Is(err, ErrOTPExpired):
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "otp has expired, please request a new one"})
	case errors.Is(err, ErrOTPInvalid):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "invalid otp"})
	case errors.Is(err, ErrSessionStale):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	case errors.Is(err, ErrOTPMaxAttempts):
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many failed attempts, please request a new otp"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

const (
	maxDisplayNameRunes = 60
	maxBioRunes         = 200
	maxGenderOtherRunes = 100
)

type multipartFile struct {
	file   multipart.File
	header *multipart.FileHeader
}

func parseUpdateProfileForm(c fiber.Ctx) (UpdateProfileRequest, error) {
	allowed := map[string]bool{
		"display_name":    true,
		"bio":             true,
		"gender":          true,
		"gender_other":    true,
		"favorite_genres": true,
		"date_of_birth":   true,
		"is_private":      true,
	}

	form, err := c.MultipartForm()
	if err != nil {
		return UpdateProfileRequest{}, errors.New("invalid multipart form")
	}

	var unknownFields []string
	for key := range form.Value {
		if !allowed[key] {
			unknownFields = append(unknownFields, key)
		}
	}
	if len(unknownFields) > 0 {
		return UpdateProfileRequest{}, fmt.Errorf("unknown fields: %s", strings.Join(unknownFields, ", "))
	}

	var req UpdateProfileRequest
	// A key that is present but empty clears the field.
	if _, ok := form.Value["display_name"]; ok {
		v := strings.TrimSpace(c.FormValue("display_name"))
		if utf8.RuneCountInString(v) > maxDisplayNameRunes {
			return req, fmt.Errorf("display_name must be at most %d characters", maxDisplayNameRunes)
		}
		req.DisplayName = &v
	}
	if _, ok := form.Value["bio"]; ok {
		v := strings.TrimSpace(c.FormValue("bio"))
		if utf8.RuneCountInString(v) > maxBioRunes {
			return req, fmt.Errorf("bio must be at most %d characters", maxBioRunes)
		}
		req.Bio = &v
	}
	if v := c.FormValue("gender"); v != "" {
		req.Gender = GenderType(v)
	}
	if _, ok := form.Value["gender_other"]; ok {
		v := strings.TrimSpace(c.FormValue("gender_other"))
		if utf8.RuneCountInString(v) > maxGenderOtherRunes {
			return req, fmt.Errorf("gender_other must be at most %d characters", maxGenderOtherRunes)
		}
		req.GenderOther = &v
	}
	if v := c.FormValue("favorite_genres"); v != "" {
		req.FavoriteGenres = &v
	}
	if _, ok := form.Value["date_of_birth"]; ok {
		v := c.FormValue("date_of_birth")
		if v == "" {
			req.ClearDateOfBirth = true
		} else {
			t, err := time.Parse("2006-01-02", v)
			if err != nil {
				return req, errors.New("date_of_birth must be YYYY-MM-DD")
			}
			if t.Before(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)) || t.After(time.Now().UTC()) {
				return req, errors.New("date_of_birth must be between 1900-01-01 and today")
			}
			req.DateOfBirth = &t
		}
	}
	if v := c.FormValue("is_private"); v != "" {
		b := v == "true"
		req.IsPrivate = &b
	}

	return req, nil
}

func validateImageFile(fh *multipart.FileHeader) error {
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	if !allowed[fh.Header.Get("Content-Type")] {
		return errors.New("avatar must be jpeg, png, webp, or gif")
	}
	const maxSize = 5 << 20 // 5 MB
	if fh.Size > maxSize {
		return errors.New("avatar must be smaller than 5 MB")
	}
	return nil
}

func parseUserID(c fiber.Ctx) (uint, error) {
	id, err := strconv.Atoi(c.Params("userId"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

const msgInvalidBody = "invalid request body"

func handleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	case errors.Is(err, ErrSessionStale):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	case errors.Is(err, ErrWrongPassword):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeWrongPassword, "error": ErrWrongPassword.Error()})
	case errors.Is(err, ErrConfirmMismatch):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": apierr.ErrorCodeConfirmUsername, "error": ErrConfirmMismatch.Error()})
	case errors.Is(err, ErrFinalAdmin):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"code": apierr.ErrorCodeFinalAdmin, "error": ErrFinalAdmin.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
