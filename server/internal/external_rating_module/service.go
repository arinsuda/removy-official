package external_rating_module

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	freshTTL        = 7 * 24 * time.Hour
	notFoundTTL     = 24 * time.Hour
	requestTimeout  = 3 * time.Second
	dailyBudget     = 900
	defaultOMDbURL  = "https://www.omdbapi.com"
	MediaTypeMovie  = "movie"
	MediaTypeSeries = "tv"
)

// Store persists cached external ratings.
type Store interface {
	Find(ctx context.Context, mediaType string, tmdbID int) (*ExternalRating, error)
	Save(ctx context.Context, row *ExternalRating) error
}

// IMDbIDResolver finds the IMDb id for a TMDB title (used when the caller does not know it).
type IMDbIDResolver func(ctx context.Context, mediaType string, tmdbID int) (string, error)

// Ratings is what callers get back. Nil fields are unavailable.
type Ratings struct {
	IMDbRating *float64
	IMDbVotes  int
	RTScore    *int
}

// budget is an in-memory per-day request counter (UTC day).
type budget struct {
	mu    sync.Mutex
	limit int
	day   string
	used  int
	now   func() time.Time
}

func (b *budget) take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	day := b.now().UTC().Format("2006-01-02")
	if day != b.day {
		b.day, b.used = day, 0
	}
	if b.used >= b.limit {
		return false
	}
	b.used++
	return true
}

type Service struct {
	store    Store
	client   *omdbClient // nil when disabled
	resolver IMDbIDResolver
	budget   *budget
	group    singleflight.Group
	timeout  time.Duration
	now      func() time.Time
}

// NewService returns a service. An empty apiKey disables OMDb lookups entirely.
func NewService(store Store, apiKey string, resolver IMDbIDResolver) *Service {
	s := &Service{
		store:    store,
		resolver: resolver,
		timeout:  requestTimeout,
		now:      time.Now,
	}
	s.budget = &budget{limit: dailyBudget, now: func() time.Time { return s.now() }}
	if apiKey != "" {
		s.client = &omdbClient{
			http:    &http.Client{Timeout: requestTimeout},
			baseURL: defaultOMDbURL,
			apiKey:  apiKey,
		}
	}
	return s
}

func (s *Service) Enabled() bool { return s != nil && s.client != nil }

// Get returns external ratings for a title. It never returns an error: on any
// failure it falls back to stale cache, then to an empty (unavailable) result.
// knownIMDbID may be empty. The wait is bounded by the service timeout.
func (s *Service) Get(ctx context.Context, mediaType string, tmdbID int, knownIMDbID string) Ratings {
	if !s.Enabled() {
		return Ratings{}
	}

	row, err := s.store.Find(ctx, mediaType, tmdbID)
	if err != nil {
		log.Printf("external ratings: cache read failed: %v", err)
		row = nil
	}
	if row != nil && s.fresh(row) {
		return toRatings(row)
	}
	if !s.budget.take() {
		return toRatings(row)
	}

	key := fmt.Sprintf("%s:%d", mediaType, tmdbID)
	// The fetch runs detached so a slow upstream still fills the cache for next time.
	ch := s.group.DoChan(key, func() (any, error) {
		return s.refresh(mediaType, tmdbID, knownIMDbID, row)
	})

	timer := time.NewTimer(s.timeout)
	defer timer.Stop()
	select {
	case res := <-ch:
		if res.Err != nil {
			log.Printf("external ratings: refresh %s failed: %v", key, res.Err)
			return toRatings(row)
		}
		return toRatings(res.Val.(*ExternalRating))
	case <-timer.C:
		return toRatings(row)
	case <-ctx.Done():
		return toRatings(row)
	}
}

func (s *Service) fresh(row *ExternalRating) bool {
	ttl := freshTTL
	if row.NotFound {
		ttl = notFoundTTL
	}
	return s.now().Sub(row.FetchedAt) < ttl
}

func (s *Service) refresh(mediaType string, tmdbID int, knownIMDbID string, prev *ExternalRating) (*ExternalRating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	imdbID := knownIMDbID
	if imdbID == "" && prev != nil {
		imdbID = prev.IMDbID
	}
	if imdbID == "" && s.resolver != nil {
		id, err := s.resolver(ctx, mediaType, tmdbID)
		if err != nil {
			return nil, err
		}
		imdbID = id
	}

	row := &ExternalRating{MediaType: mediaType, TMDBID: tmdbID, IMDbID: imdbID, FetchedAt: s.now()}

	if imdbID == "" {
		row.NotFound = true
	} else {
		res, err := s.client.fetch(ctx, imdbID)
		switch {
		case errors.Is(err, errOMDbNotFound):
			row.NotFound = true
		case err != nil:
			return nil, err
		default:
			row.IMDbRating, row.IMDbVotes, row.RTScore = res.IMDbRating, res.IMDbVotes, res.RTScore
		}
	}

	if err := s.store.Save(ctx, row); err != nil {
		log.Printf("external ratings: cache write failed: %v", err)
	}
	return row, nil
}

func toRatings(row *ExternalRating) Ratings {
	if row == nil || row.NotFound {
		return Ratings{}
	}
	out := Ratings{IMDbRating: row.IMDbRating, RTScore: row.RTScore}
	if row.IMDbVotes != nil {
		out.IMDbVotes = *row.IMDbVotes
	}
	return out
}

// GormStore is the PostgreSQL-backed Store.
type GormStore struct{ db *gorm.DB }

func NewGormStore(db *gorm.DB) *GormStore { return &GormStore{db: db} }

func (g *GormStore) Find(ctx context.Context, mediaType string, tmdbID int) (*ExternalRating, error) {
	var row ExternalRating
	err := g.db.WithContext(ctx).Where("media_type = ? AND tmdb_id = ?", mediaType, tmdbID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (g *GormStore) Save(ctx context.Context, row *ExternalRating) error {
	return g.db.WithContext(ctx).Clauses(onConflictUpdate()).Create(row).Error
}

func onConflictUpdate() clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{{Name: "media_type"}, {Name: "tmdb_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"imdb_id", "imdb_rating", "imdb_votes", "rt_score", "not_found", "fetched_at",
		}),
	}
}
