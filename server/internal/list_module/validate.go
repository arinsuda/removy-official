package list_module

import (
	"strings"
	"unicode/utf8"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
)

func validName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > MaxListNameRunes {
		return "", ErrInvalidInput
	}
	return name, nil
}

func validDescription(desc string) (string, error) {
	desc = strings.TrimSpace(desc)
	if utf8.RuneCountInString(desc) > MaxListDescRunes {
		return "", ErrInvalidInput
	}
	return desc, nil
}

func validVisibility(v privacy_policy.ListVisibility) error {
	if !v.Valid() {
		return ErrInvalidVisibility
	}
	return nil
}

func validMedia(req AddItemRequest) error {
	if req.MediaType != shared.MediaTypeMovie && req.MediaType != shared.MediaTypeTV {
		return ErrInvalidMediaType
	}
	if req.MediaID <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// samePermutation reports whether got holds exactly the ids in want, each once.
func samePermutation(want, got []uint) bool {
	if len(want) != len(got) {
		return false
	}
	seen := make(map[uint]struct{}, len(want))
	for _, id := range want {
		seen[id] = struct{}{}
	}
	for _, id := range got {
		if _, ok := seen[id]; !ok {
			return false
		}
		delete(seen, id)
	}
	return len(seen) == 0
}
