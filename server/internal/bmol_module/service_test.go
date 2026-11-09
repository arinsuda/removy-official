package bmol_module

import (
	"errors"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/internal/movie_module"
)

func ptr[T any](v T) *T { return &v }

func TestValidateCreate(t *testing.T) {
	ok := CreateBMOLRequest{MediaID: 603, MediaType: movie_module.MediaMovie, Rank: 1}
	if err := validateCreate(ok); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	cases := map[string]struct {
		mut  func(*CreateBMOLRequest)
		want error
	}{
		"rank zero":     {func(r *CreateBMOLRequest) { r.Rank = 0 }, ErrInvalidInput},
		"rank negative": {func(r *CreateBMOLRequest) { r.Rank = -3 }, ErrInvalidInput},
		"rank too big":  {func(r *CreateBMOLRequest) { r.Rank = 1000 }, ErrInvalidInput},
		"media zero":    {func(r *CreateBMOLRequest) { r.MediaID = 0 }, ErrInvalidInput},
		"bad type":      {func(r *CreateBMOLRequest) { r.MediaType = "book" }, ErrInvalidMediaType},
	}
	for name, c := range cases {
		r := ok
		c.mut(&r)
		if err := validateCreate(r); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v want %v", name, err, c.want)
		}
	}
}

func TestValidateUpdate(t *testing.T) {
	if _, err := validateUpdate(UpdateBMOLRequest{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty update must fail: %v", err)
	}
	if _, err := validateUpdate(UpdateBMOLRequest{Rank: ptr(0)}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("rank 0 must fail: %v", err)
	}
	n, err := validateUpdate(UpdateBMOLRequest{Note: ptr("  hi  ")})
	if err != nil || n == nil || *n != "hi" {
		t.Fatalf("note not trimmed: %v %v", n, err)
	}
	// 280 multi-byte runes are fine, 281 are not.
	if _, err := validateUpdate(UpdateBMOLRequest{Note: ptr(strings.Repeat("ก", 280))}); err != nil {
		t.Fatalf("280 runes rejected: %v", err)
	}
	if _, err := validateUpdate(UpdateBMOLRequest{Note: ptr(strings.Repeat("ก", 281))}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("281 runes accepted: %v", err)
	}
	if n, err := validateUpdate(UpdateBMOLRequest{Rank: ptr(5)}); err != nil || n != nil {
		t.Fatalf("rank-only update: %v %v", n, err)
	}
}
