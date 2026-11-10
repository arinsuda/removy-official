package user_module

import (
	"strings"
	"testing"
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
