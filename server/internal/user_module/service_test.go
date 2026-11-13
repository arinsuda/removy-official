package user_module

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestResetLinkUsesConfiguredFrontendBaseURL(t *testing.T) {
	const base = "https://app.example.test"
	svc := &Service{}
	svc.SetFrontendBaseURL(base)

	link := svc.resetLink("tok123", 42)

	if !strings.HasPrefix(link, base+ResetPasswordPath) {
		t.Fatalf("link %q should start with %q", link, base+ResetPasswordPath)
	}
	if !strings.Contains(link, "#") || !strings.Contains(link, "tok123") {
		t.Fatalf("link %q should carry the token in the fragment", link)
	}
}

func TestRedactProfileForViewerOmitsOwnerFieldsForAllNonOwners(t *testing.T) {
	now := time.Now().UTC()
	makeProfile := func() *UserProfileResponse {
		return &UserProfileResponse{
			ID:              10,
			Username:        "alice",
			Email:           "alice@example.com",
			VerifiedEmailAt: &now,
			DateOfBirth:     &now,
			Gender:          GenderFemale,
			Role:            "admin",
			OnboardedAt:     &now,
			CreatedAt:       now,
		}
	}

	omittedKeys := []string{"email", "verified_email_at", "date_of_birth", "gender", "role", "onboarded_at"}

	for _, tc := range []struct {
		name        string
		requesterID uint
		wantOmitted bool
	}{
		{name: "guest", requesterID: 0, wantOmitted: true},
		{name: "signed-in non-owner", requesterID: 99, wantOmitted: true},
		{name: "owner", requesterID: 10, wantOmitted: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := makeProfile()
			redactProfileForViewer(p, tc.requesterID, 10)
			raw, err := json.Marshal(p)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			for _, key := range omittedKeys {
				_, exists := m[key]
				if tc.wantOmitted && exists {
					t.Errorf("key %q should be omitted for %s, got %v", key, tc.name, m[key])
				}
				if !tc.wantOmitted && !exists {
					t.Errorf("key %q should be present for owner", key)
				}
			}
		})
	}
}
