package event_module

import (
	"context"
	"sort"
	"strconv"

	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

const (
	// maxAwardsPerQuery caps one GET /events/awards response; a year holds far fewer winners.
	maxAwardsPerQuery = 200
	// minAwardYear and maxAwardYear bound the ?year filter.
	minAwardYear = 2000
	maxAwardYear = 9999
)

// AwardReader serves a title's award history to the movie and TV detail pages.
// It implements shared.AwardReader.
type AwardReader struct {
	db *gorm.DB
}

func NewAwardReader(db *gorm.DB) *AwardReader {
	return &AwardReader{db: db}
}

// AwardsFor returns the awards of every title in ids with one query, newest period first.
// Titles without awards are absent from the map. Only published, non-deleted rounds count.
func (r *AwardReader) AwardsFor(ctx context.Context, ids []shared.MediaIdentity) (map[shared.MediaKey][]shared.MediaAward, error) {
	out := make(map[shared.MediaKey][]shared.MediaAward, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	pairs := make([][]interface{}, 0, len(ids))
	for _, id := range ids {
		pairs = append(pairs, []interface{}{string(id.Type), id.ID})
	}

	var rows []struct {
		MediaID   int
		MediaType string
		Kind      string
		Period    string
	}
	err := r.db.WithContext(ctx).Table("event_awards a").
		Select("a.media_id, a.media_type, a.award_kind AS kind, a.period").
		Joins("JOIN events e ON e.id = a.event_id").
		Where("e.deleted_at IS NULL AND e.status = ?", StatusPublished).
		Where("(a.media_type, a.media_id) IN ?", pairs).
		Order("a.period DESC, a.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		key := shared.MediaKey{ID: row.MediaID, Type: shared.MediaType(row.MediaType)}
		out[key] = append(out[key], shared.MediaAward{Kind: row.Kind, Period: row.Period})
	}
	return out, nil
}

// awardRow is one winner joined to its nominee snapshot.
type awardRow struct {
	EventID      uint
	Kind         string
	Period       string
	MediaID      int
	MediaType    string
	Votes        int
	MediaTitle   string
	PosterPath   string
	BackdropPath string
}

// awardsQuery is the base of the awards listing: published, non-deleted rounds, optionally one media type.
func (s *Service) awardsQuery(ctx context.Context, mediaType string) *gorm.DB {
	q := s.db.WithContext(ctx).Table("event_awards a").
		Joins("JOIN events e ON e.id = a.event_id").
		Where("e.deleted_at IS NULL AND e.status = ?", StatusPublished)
	if mediaType != "" {
		q = q.Where("a.media_type = ?", mediaType)
	}
	return q
}

// ListAwards returns the winners of calendar year `year` (optionally one media type), newest
// period first, plus every year that has winners for that media type, newest first. Periods are
// "YYYY-MM" or "YYYY", so the year is a string range on period.
func (s *Service) ListAwards(ctx context.Context, mediaType string, year int) ([]AwardResponse, []int, error) {
	var rows []awardRow
	err := s.awardsQuery(ctx, mediaType).
		Joins("JOIN event_nominees n ON n.id = a.nominee_id").
		Select("a.event_id, a.award_kind AS kind, a.period, a.media_id, a.media_type, a.votes, "+
			"n.media_title, n.poster_path, n.backdrop_path").
		Where("a.period >= ? AND a.period < ?", strconv.Itoa(year), strconv.Itoa(year+1)).
		Order("a.period DESC, a.id ASC").
		Limit(maxAwardsPerQuery).
		Scan(&rows).Error
	if err != nil {
		return nil, nil, err
	}
	awards := make([]AwardResponse, 0, len(rows))
	for _, row := range rows {
		awards = append(awards, AwardResponse{
			EventID:      row.EventID,
			Kind:         row.Kind,
			Period:       row.Period,
			MediaID:      row.MediaID,
			MediaType:    row.MediaType,
			MediaTitle:   row.MediaTitle,
			PosterPath:   optional(row.PosterPath),
			BackdropPath: optional(row.BackdropPath),
			Votes:        row.Votes,
		})
	}

	var years []int
	if err := s.awardsQuery(ctx, mediaType).
		Distinct().
		Pluck("CAST(LEFT(a.period, 4) AS INTEGER)", &years).Error; err != nil {
		return nil, nil, err
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	if years == nil {
		years = []int{}
	}
	return awards, years, nil
}
