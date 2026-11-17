package shared

import (
	"fmt"
	"testing"
)

func TestNewMangaMetaCapsGenresAndStoresArray(t *testing.T) {
	many := make([]string, MangaMetaMaxGenres+3)
	for i := range many {
		many[i] = fmt.Sprintf("G%d", i)
	}
	row := newMangaMeta(1, MediaInfo{Title: "T", StartYear: 2019, GenreNames: many})
	if len(row.Genres) != MangaMetaMaxGenres || row.StartYear != 2019 {
		t.Fatalf("capped genres/year: %d %d", len(row.Genres), row.StartYear)
	}

	empty := newMangaMeta(2, MediaInfo{Title: "T"})
	if empty.Genres == nil || len(empty.Genres) != 0 {
		t.Fatalf("nil genres must store an empty array, got %#v", empty.Genres)
	}
	if marker := notFoundMangaMeta(3); marker.Genres == nil {
		t.Fatalf("not-found marker must store an empty array")
	}
}
