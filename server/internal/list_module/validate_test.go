package list_module

import (
	"errors"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
)

func TestValidName(t *testing.T) {
	if n, err := validName("  Best of 90s "); err != nil || n != "Best of 90s" {
		t.Fatalf("trim: %q %v", n, err)
	}
	for _, bad := range []string{"", "   ", strings.Repeat("a", MaxListNameRunes+1)} {
		if _, err := validName(bad); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q accepted", bad)
		}
	}
	// Limits count runes, not bytes.
	if _, err := validName(strings.Repeat("ก", MaxListNameRunes)); err != nil {
		t.Fatalf("max-length thai name rejected: %v", err)
	}
}

func TestValidDescription(t *testing.T) {
	if _, err := validDescription(""); err != nil {
		t.Fatal("empty description must be allowed")
	}
	if _, err := validDescription(strings.Repeat("ก", MaxListDescRunes+1)); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("over-long description accepted")
	}
}

func TestValidVisibilityAndMedia(t *testing.T) {
	if validVisibility(privacy_policy.ListFollowers) != nil || !errors.Is(validVisibility("friends"), ErrInvalidVisibility) {
		t.Fatal("visibility validation")
	}
	if !errors.Is(validMedia(AddItemRequest{MediaID: 1, MediaType: "book"}), ErrInvalidMediaType) {
		t.Fatal("bad media type accepted")
	}
	if !errors.Is(validMedia(AddItemRequest{MediaID: 0, MediaType: shared.MediaTypeTV}), ErrInvalidInput) {
		t.Fatal("zero media id accepted")
	}
}

func TestSamePermutation(t *testing.T) {
	cases := []struct {
		want, got []uint
		ok        bool
	}{
		{[]uint{1, 2, 3}, []uint{3, 1, 2}, true},
		{[]uint{1, 2, 3}, []uint{1, 2}, false},
		{[]uint{1, 2, 3}, []uint{1, 2, 2}, false},
		{[]uint{1, 2, 3}, []uint{1, 2, 4}, false},
		{nil, []uint{}, true},
	}
	for i, c := range cases {
		if got := samePermutation(c.want, c.got); got != c.ok {
			t.Errorf("case %d: got %v", i, got)
		}
	}
}
