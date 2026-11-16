package manga_module

import (
	"context"
	"errors"
	"strconv"

	anilist "github.com/arinsuda/movie-hub/internal/anilist_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

// LookupMedia serves the shared media resolver: one AniList title with its adult flag. The
// viewer gate is not applied here; shared.CheckMediaAccess and MediaVisibility decide per request.
func (s *MangaService) LookupMedia(ctx context.Context, id int) (shared.MediaInfo, error) {
	m, err := s.media(ctx, detailQuery, idVariables(id), true)
	if err != nil {
		return shared.MediaInfo{}, mapLookupErr(err)
	}
	return mediaInfo(*m), nil
}

// mediaInfo maps one AniList title to the shared media record that manga_meta stores.
func mediaInfo(m rawMedia) shared.MediaInfo {
	item := toItem(m)
	info := shared.MediaInfo{
		Title:        item.DisplayTitle,
		EnglishTitle: item.Titles.English,
		PosterURL:    item.CoverURL,
		Kind:         item.Kind,
		IsAdult:      item.IsAdult,
	}
	if item.StartYear != nil {
		info.ReleaseDate = strconv.Itoa(*item.StartYear)
	}
	return info
}

func mapLookupErr(err error) error {
	if errors.Is(err, anilist.ErrNotFound) {
		return shared.ErrMediaNotFound
	}
	return shared.ErrMediaUnavailable
}

// InstallMediaSupport hands the AniList lookup, the adult gate and the manga_meta store to the shared media layer.
// The router calls it once at boot.
func (h *Handler) InstallMediaSupport(db *gorm.DB) {
	shared.SetMangaSupport(h.svc, h.gate, db)
}
