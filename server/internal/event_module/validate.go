package event_module

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/arinsuda/movie-hub/internal/shared"
)

const (
	maxTitleRunes       = 120
	maxBodyRunes        = 2000
	maxLinkURLLength    = 500
	maxNomineeTitleRune = 255
	minPopupPriority    = -100
	maxPopupPriority    = 100
)

var (
	monthlyPeriodRe = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)
	yearlyPeriodRe  = regexp.MustCompile(`^\d{4}$`)
)

var ErrInvalidInput = errors.New("invalid event input")

// ValidationError names the field that failed; the message is for logs/devs, the client maps the code.
type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return "invalid field: " + e.Field }
func (e *ValidationError) Unwrap() error { return ErrInvalidInput }

func invalid(field string) error { return &ValidationError{Field: field} }

// normalize trims text fields in place and validates the input.
func (in *EventInput) normalize() error {
	if in.Type == "" {
		in.Type = TypeAnnouncement
	}
	if !in.Type.Valid() {
		return invalid("type")
	}

	if in.Type.IsVoting() {
		if in.MediaType == nil || !validMediaType(*in.MediaType) {
			return invalid("media_type")
		}
		if in.Period == nil {
			return invalid("period")
		}
		period := strings.TrimSpace(*in.Period)
		if in.Type == TypeVoteMonthly && !monthlyPeriodRe.MatchString(period) {
			return invalid("period")
		}
		if in.Type == TypeVoteYearly && !yearlyPeriodRe.MatchString(period) {
			return invalid("period")
		}
		in.Period = &period
	} else {
		in.MediaType = nil
		in.Period = nil
	}

	in.TitleEN = strings.TrimSpace(in.TitleEN)
	in.TitleTH = strings.TrimSpace(in.TitleTH)
	in.BodyEN = strings.TrimSpace(in.BodyEN)
	in.BodyTH = strings.TrimSpace(in.BodyTH)
	in.LinkURL = strings.TrimSpace(in.LinkURL)

	// Generated monthly rounds may be saved without copy (the copy is filled in at display time);
	// every other kind, including yearly rounds, needs titles.
	requireTitles := in.Type != TypeVoteMonthly
	if (requireTitles && in.TitleEN == "") || utf8.RuneCountInString(in.TitleEN) > maxTitleRunes {
		return invalid("title_en")
	}
	if (requireTitles && in.TitleTH == "") || utf8.RuneCountInString(in.TitleTH) > maxTitleRunes {
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

	if len(in.Nominees) > 0 {
		if !in.Type.IsVoting() {
			return invalid("nominees")
		}
		if err := normalizeNominees(in.Nominees, *in.MediaType); err != nil {
			return err
		}
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

func validMediaType(v string) bool {
	return v == string(shared.MediaTypeMovie) || v == string(shared.MediaTypeTV)
}

// normalizeNominees trims and validates a nominee list for a round of mediaType.
// Positions default to list order; a title may appear only once per round.
func normalizeNominees(noms []NomineeInput, mediaType string) error {
	if len(noms) > maxNomineesPerRound {
		return invalid("nominees")
	}
	seen := make(map[int]bool, len(noms))
	for i := range noms {
		nom := &noms[i]
		nom.MediaTitle = strings.TrimSpace(nom.MediaTitle)
		if nom.MediaID <= 0 || seen[nom.MediaID] {
			return invalid("nominees.media_id")
		}
		seen[nom.MediaID] = true
		if nom.MediaType != mediaType {
			return invalid("nominees.media_type")
		}
		if nom.MediaTitle == "" || utf8.RuneCountInString(nom.MediaTitle) > maxNomineeTitleRune {
			return invalid("nominees.media_title")
		}
		if nom.Position <= 0 {
			nom.Position = i + 1
		}
	}
	return nil
}
