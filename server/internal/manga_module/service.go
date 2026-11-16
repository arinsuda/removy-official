package manga_module

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/arinsuda/movie-hub/config"
	anilist "github.com/arinsuda/movie-hub/internal/anilist_module"
)

// ErrAgeRestricted means the title is adult and the viewer is not eligible to see it.
var ErrAgeRestricted = errors.New("manga: age restricted")

// MangaService turns AniList responses into API DTOs. It has no database access.
type MangaService struct {
	client *anilist.Client
}

func NewMangaService(client *anilist.Client) *MangaService {
	return &MangaService{client: client}
}

// rawPage is the Page selection shared by browse and search.
type rawPage struct {
	PageInfo rawPageInfo `json:"pageInfo"`
	Media    []rawMedia  `json:"media"`
}

// List returns one browse page for the validated filter; f.Adult must already reflect the viewer gate.
func (s *MangaService) List(ctx context.Context, f ListFilter) (*PaginatedResult[MangaItem], error) {
	return s.page(ctx, listQuery, listVariables(f), f.Page, f.Adult)
}

// Search returns one page of title matches for q.
func (s *MangaService) Search(ctx context.Context, q string, page int, adult bool) (*PaginatedResult[MangaItem], error) {
	return s.page(ctx, listQuery, searchVariables(q, page, adult), page, adult)
}

func (s *MangaService) page(ctx context.Context, query string, vars map[string]any, page int, adult bool) (*PaginatedResult[MangaItem], error) {
	raw, err := s.client.Query(ctx, scopeOf(adult), query, vars, config.AniListListCacheTTL)
	if err != nil {
		return nil, err
	}
	var data struct {
		Page rawPage `json:"Page"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, anilist.ErrUnavailable
	}

	items := make([]MangaItem, 0, len(data.Page.Media))
	for _, m := range data.Page.Media {
		if isVisible(m, adult) {
			items = append(items, toItem(m))
		}
	}
	return &PaginatedResult[MangaItem]{
		Page:         page,
		Results:      items,
		TotalPages:   data.Page.PageInfo.LastPage,
		TotalResults: data.Page.PageInfo.Total,
	}, nil
}

// GetByID loads one title with its long-form fields. eligible is the viewer's adult eligibility.
func (s *MangaService) GetByID(ctx context.Context, id int, eligible bool) (*MangaDetail, error) {
	m, err := s.media(ctx, detailQuery, idVariables(id), eligible)
	if err != nil {
		return nil, err
	}
	d := toDetail(*m)
	return &d, nil
}

// Similar returns AniList recommendations for the title. eligible gates the title itself (an adult title
// the viewer may not open is ErrAgeRestricted); showAdult decides whether adult recommendations are listed.
func (s *MangaService) Similar(ctx context.Context, id int, eligible, showAdult bool) ([]MangaItem, error) {
	m, err := s.media(ctx, similarQuery, similarVariables(id), eligible)
	if err != nil {
		return nil, err
	}
	return toSimilar(m.Recommendations, showAdult), nil
}

// media runs a single-title query. A missing or non-manga title is anilist.ErrNotFound; an adult
// title the viewer may not see is ErrAgeRestricted. The cached payload is viewer-independent
// (detailScope), so the gate runs here, after the cache.
func (s *MangaService) media(ctx context.Context, query string, vars map[string]any, eligible bool) (*rawMedia, error) {
	raw, err := s.client.Query(ctx, detailScope, query, vars, config.AniListDetailCacheTTL)
	if err != nil {
		return nil, err
	}
	var data struct {
		Media *rawMedia `json:"Media"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, anilist.ErrUnavailable
	}
	if data.Media == nil || data.Media.Type != aniListManga {
		return nil, anilist.ErrNotFound
	}
	if data.Media.IsAdult && !eligible {
		return nil, ErrAgeRestricted
	}
	return data.Media, nil
}
