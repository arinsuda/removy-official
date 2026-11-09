package bmol_module

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/arinsuda/movie-hub/internal/movie_module"
)

const (
	MinRank      = 1
	MaxRank      = 999
	MaxNoteRunes = 280
)

var ErrInvalidInput = errors.New("invalid input")

func validRank(rank int) bool { return rank >= MinRank && rank <= MaxRank }

// normalizeNote trims the note and enforces the rune limit.
func normalizeNote(note string) (string, error) {
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) > MaxNoteRunes {
		return "", ErrInvalidInput
	}
	return note, nil
}

func validateCreate(req CreateBMOLRequest) error {
	if req.MediaType != movie_module.MediaMovie && req.MediaType != movie_module.MediaSeries {
		return ErrInvalidMediaType
	}
	if req.MediaID <= 0 || !validRank(req.Rank) {
		return ErrInvalidInput
	}
	return nil
}

// validateUpdate returns the normalized note (when present).
func validateUpdate(req UpdateBMOLRequest) (*string, error) {
	if req.Rank == nil && req.Note == nil {
		return nil, ErrInvalidInput
	}
	if req.Rank != nil && !validRank(*req.Rank) {
		return nil, ErrInvalidInput
	}
	if req.Note == nil {
		return nil, nil
	}
	n, err := normalizeNote(*req.Note)
	if err != nil {
		return nil, err
	}
	return &n, nil
}
