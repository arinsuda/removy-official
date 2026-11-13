package event_module

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	maxTitleRunes    = 120
	maxBodyRunes     = 2000
	maxLinkURLLength = 500
	minPopupPriority = -100
	maxPopupPriority = 100
)

var ErrInvalidInput = errors.New("invalid event input")

// ValidationError names the field that failed; the message is for logs/devs, the client maps the code.
type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return "invalid field: " + e.Field }
func (e *ValidationError) Unwrap() error { return ErrInvalidInput }

func invalid(field string) error { return &ValidationError{Field: field} }

// normalize trims text fields in place and validates the input.
func (in *EventInput) normalize() error {
	in.TitleEN = strings.TrimSpace(in.TitleEN)
	in.TitleTH = strings.TrimSpace(in.TitleTH)
	in.BodyEN = strings.TrimSpace(in.BodyEN)
	in.BodyTH = strings.TrimSpace(in.BodyTH)
	in.LinkURL = strings.TrimSpace(in.LinkURL)

	if in.TitleEN == "" || utf8.RuneCountInString(in.TitleEN) > maxTitleRunes {
		return invalid("title_en")
	}
	if in.TitleTH == "" || utf8.RuneCountInString(in.TitleTH) > maxTitleRunes {
		return invalid("title_th")
	}
	if utf8.RuneCountInString(in.BodyEN) > maxBodyRunes {
		return invalid("body_en")
	}
	if utf8.RuneCountInString(in.BodyTH) > maxBodyRunes {
		return invalid("body_th")
	}
	if !validLinkURL(in.LinkURL) {
		return invalid("link_url")
	}
	if in.StartsAt.IsZero() || in.EndsAt.IsZero() || !in.StartsAt.Before(in.EndsAt) {
		return invalid("ends_at")
	}
	if in.PopupPriority < minPopupPriority || in.PopupPriority > maxPopupPriority {
		return invalid("popup_priority")
	}
	if in.Status == "" {
		in.Status = StatusDraft
	}
	if !in.Status.Valid() {
		return invalid("status")
	}
	in.StartsAt = in.StartsAt.UTC()
	in.EndsAt = in.EndsAt.UTC()
	return nil
}

// validLinkURL accepts "" , an internal path ("/events", not "//host"), or an absolute http(s) URL.
func validLinkURL(v string) bool {
	if v == "" {
		return true
	}
	if len(v) > maxLinkURLLength || strings.ContainsAny(v, " \t\r\n\\") {
		return false
	}
	if strings.HasPrefix(v, "/") {
		return !strings.HasPrefix(v, "//")
	}
	u, err := url.Parse(v)
	if err != nil || u.Host == "" || u.User != nil {
		return false
	}
	return u.Scheme == "https" || u.Scheme == "http"
}
