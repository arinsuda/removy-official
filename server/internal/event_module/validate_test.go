package event_module

import (
	"strings"
	"testing"
	"time"
)

func validTestInput() EventInput {
	now := time.Now().UTC()
	return EventInput{
		TitleEN:       "Summer Movie Festival",
		TitleTH:       "เทศกาลภาพยนตร์ฤดูร้อน",
		BodyEN:        "Join our annual festival!",
		BodyTH:        "ร่วมสนุกกับเทศกาลประจำปี!",
		LinkURL:       "/movies",
		StartsAt:      now.Add(-time.Hour),
		EndsAt:        now.Add(24 * time.Hour),
		ShowPopup:     true,
		PopupPriority: 10,
		Status:        StatusPublished,
	}
}

func TestEventInputNormalize_Valid(t *testing.T) {
	in := validTestInput()
	if err := in.normalize(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
	if in.TitleEN != "Summer Movie Festival" {
		t.Errorf("unexpected TitleEN: %s", in.TitleEN)
	}
}

func TestEventInputNormalize_TitleValidation(t *testing.T) {
	// Empty English title
	in := validTestInput()
	in.TitleEN = "   "
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_en") {
		t.Errorf("expected title_en error, got: %v", err)
	}

	// Empty Thai title
	in = validTestInput()
	in.TitleTH = ""
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_th") {
		t.Errorf("expected title_th error, got: %v", err)
	}

	// Too long English title (>120 runes)
	in = validTestInput()
	in.TitleEN = strings.Repeat("A", 121)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_en") {
		t.Errorf("expected title_en length error, got: %v", err)
	}

	// 120 Thai runes is fine
	in = validTestInput()
	in.TitleTH = strings.Repeat("ก", 120)
	if err := in.normalize(); err != nil {
		t.Errorf("expected 120 runes valid, got: %v", err)
	}

	// 121 Thai runes fails
	in.TitleTH = strings.Repeat("ก", 121)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_th") {
		t.Errorf("expected title_th length error, got: %v", err)
	}
}

func TestEventInputNormalize_BodyValidation(t *testing.T) {
	in := validTestInput()
	in.BodyEN = strings.Repeat("B", 2001)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "body_en") {
		t.Errorf("expected body_en error, got: %v", err)
	}

	in = validTestInput()
	in.BodyTH = strings.Repeat("ข", 2001)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "body_th") {
		t.Errorf("expected body_th error, got: %v", err)
	}
}

func TestEventInputNormalize_LinkURL(t *testing.T) {
	cases := []struct {
		url   string
		valid bool
	}{
		{"", true},
		{"/events", true},
		{"/movies/123", true},
		{"https://example.com/details", true},
		{"http://example.com", true},
		{"//evil.com", false},
		{"javascript:alert(1)", false},
		{"ftp://files.example.com", false},
		{"/with space", false},
		{"https://user:pass@example.com", false},
	}

	for _, c := range cases {
		in := validTestInput()
		in.LinkURL = c.url
		err := in.normalize()
		if c.valid && err != nil {
			t.Errorf("expected url %q to be valid, got: %v", c.url, err)
		}
		if !c.valid && err == nil {
			t.Errorf("expected url %q to be invalid, but passed", c.url)
		}
	}
}

func TestEventInputNormalize_Dates(t *testing.T) {
	now := time.Now().UTC()
	// ends_at before starts_at
	in := validTestInput()
	in.StartsAt = now.Add(time.Hour)
	in.EndsAt = now
	if err := in.normalize(); err == nil {
		t.Error("expected ends_at before starts_at to fail")
	}

	// starts_at == ends_at
	in = validTestInput()
	in.StartsAt = now
	in.EndsAt = now
	if err := in.normalize(); err == nil {
		t.Error("expected equal dates to fail")
	}

	// Zero dates
	in = validTestInput()
	in.StartsAt = time.Time{}
	if err := in.normalize(); err == nil {
		t.Error("expected zero starts_at to fail")
	}
}

func TestEventInputNormalize_PriorityAndStatus(t *testing.T) {
	in := validTestInput()
	in.PopupPriority = 101
	if err := in.normalize(); err == nil {
		t.Error("expected priority > 100 to fail")
	}

	in = validTestInput()
	in.PopupPriority = -101
	if err := in.normalize(); err == nil {
		t.Error("expected priority < -100 to fail")
	}

	in = validTestInput()
	in.Status = "unknown"
	if err := in.normalize(); err == nil {
		t.Error("expected invalid status to fail")
	}

	in = validTestInput()
	in.Status = ""
	if err := in.normalize(); err != nil || in.Status != StatusDraft {
		t.Errorf("expected empty status to default to draft, got: %v, status: %s", err, in.Status)
	}
}

func TestEventStateAt(t *testing.T) {
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	ev := Event{
		StartsAt: time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC),
	}

	// During
	if ev.StateAt(now) != StateActive {
		t.Errorf("expected active, got %s", ev.StateAt(now))
	}

	// Before
	if ev.StateAt(now.Add(-10*24*time.Hour)) != StateUpcoming {
		t.Errorf("expected upcoming, got %s", ev.StateAt(now.Add(-10*24*time.Hour)))
	}

	// After
	if ev.StateAt(now.Add(10*24*time.Hour)) != StatePast {
		t.Errorf("expected past, got %s", ev.StateAt(now.Add(10*24*time.Hour)))
	}
}
