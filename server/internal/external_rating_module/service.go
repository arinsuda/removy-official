package external_rating_module

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	freshTTL       = 7 * 24 * time.Hour
	notFoundTTL    = 24 * time.Hour
	requestTimeout = 3 * time.Second
	coldMissWait   = 1500 * time.Millisecond
	dailyBudget    = 900
	// Prewarm never spends the last prewarmReserve units; those stay for on-demand detail views.
	prewarmReserve     = 300
	prewarmMaxIDs      = 20
	prewarmConcurrency = 3
	prewarmTimeout     = 2 * time.Minute
	defaultOMDbURL     = "https://www.omdbapi.com"
	MediaTypeMovie     = "movie"
	MediaTypeSeries    = "tv"
)

var (
	errBudgetExhausted = errors.New("daily OMDb budget exhausted")
	imdbIDPattern      = regexp.MustCompile(`^tt\d{1,10}$`)
)

// Store persists cached external ratings.
type Store interface {
	Find(ctx context.Context, mediaType string, tmdbID int) (*ExternalRating, error)
	Save(ctx context.Context, row *ExternalRating) error
	// FindMany returns existing rows keyed by tmdb id (one query).
	FindMany(ctx context.Context, mediaType string, tmdbIDs []int) (map[int]*ExternalRating, error)
}

// IMDbIDResolver finds the IMDb id for a TMDB title (used when the caller does not know it).
type IMDbIDResolver func(ctx context.Context, mediaType string, tmdbID int) (string, error)

// Ratings is what callers get back. Nil fields are unavailable.
type Ratings struct {
	IMDbRating *float64
	IMDbVotes  int
	RTScore    *int
	// Pending means the lookup was still running when we gave up waiting; ask again shortly.
	Pending bool
}

// budget is an in-memory per-day request counter (UTC day).
type budget struct {
	mu    sync.Mutex
	limit int
	day   string
	used  int
	now   func() time.Time
}

func (b *budget) take() bool { return b.takeKeeping(0) }

// takeKeeping takes one unit only while more than keep units remain.
func (b *budget) takeKeeping(keep int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	day := b.now().UTC().Format("2006-01-02")
	if day != b.day {
		b.day, b.used = day, 0
	}
	if b.limit-b.used <= keep || b.used >= b.limit {
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
	timeout  time.Duration // upstream fetch deadline
	wait     time.Duration // max time a cold-miss request waits for the fetch
	now      func() time.Time
}

// NewService returns a service. An empty apiKey disables OMDb lookups entirely.
func NewService(store Store, apiKey string, resolver IMDbIDResolver) *Service {
	s := &Service{
		store:    store,
		resolver: resolver,
		timeout:  requestTimeout,
		wait:     coldMissWait,
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
// knownIMDbID may be empty. A stale row is returned immediately and refreshed in
// the background; a cold miss waits at most coldMissWait.
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
	key := fmt.Sprintf("%s:%d", mediaType, tmdbID)
	// Fetches run detached from the request so a slow upstream still fills the cache.
	ch := s.group.DoChan(key, func() (any, error) {
		// The budget is taken inside the singleflight call so concurrent
		// requests for one title spend a single unit. It is in-memory, per instance.
		if !s.budget.take() {
			return nil, errBudgetExhausted
		}
		return s.refresh(mediaType, tmdbID, knownIMDbID, row)
	})

	if row != nil {
		// Stale: serve it now, the refresh continues in the background.
		return toRatings(row)
	}

	// Cold miss: wait briefly, then report unavailable while the fetch keeps going.
	timer := time.NewTimer(s.wait)
	defer timer.Stop()
	select {
	case res := <-ch:
		if res.Err != nil {
			if !errors.Is(res.Err, errBudgetExhausted) {
				log.Printf("external ratings: refresh %s failed: %v", key, res.Err)
			}
			return Ratings{}
		}
		return toRatings(res.Val.(*ExternalRating))
	case <-timer.C:
		return Ratings{Pending: true}
	case <-ctx.Done():
		return Ratings{}
	}
}

// Prewarm fills the cache for titles that have no fresh row, in the background.
// It handles at most prewarmMaxIDs ids, runs prewarmConcurrency at a time, shares
// singleflight with Get, and stops spending once the budget is down to
// prewarmReserve. It never blocks the caller.
func (s *Service) Prewarm(mediaType string, ids []int) {
	if !s.Enabled() || len(ids) == 0 {
		return
	}
	seen := make(map[int]bool, len(ids))
	uniq := make([]int, 0, prewarmMaxIDs)
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		if uniq = append(uniq, id); len(uniq) == prewarmMaxIDs {
			break
		}
	}
	if len(uniq) == 0 {
		return
	}
	go s.prewarm(mediaType, uniq)
}

func (s *Service) prewarm(mediaType string, ids []int) {
	defer func() { _ = recover() }() // background work must never crash the server
	ctx, cancel := context.WithTimeout(context.Background(), prewarmTimeout)
	defer cancel()

	rows, err := s.store.FindMany(ctx, mediaType, ids)
	if err != nil {
		log.Printf("external ratings: prewarm cache read failed: %v", err)
		return
	}
	todo := make([]int, 0, len(ids))
	for _, id := range ids {
		if r := rows[id]; r == nil || !s.fresh(r) {
			todo = append(todo, id)
		}
	}

	sem := make(chan struct{}, prewarmConcurrency)
	var wg sync.WaitGroup
	for _, id := range todo {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()
			defer func() { _ = recover() }()
			key := fmt.Sprintf("%s:%d", mediaType, id)
			prev := rows[id]
			ch := s.group.DoChan(key, func() (any, error) {
				if !s.budget.takeKeeping(prewarmReserve) {
					return nil, errBudgetExhausted
				}
				return s.refresh(mediaType, id, "", prev)
			})
			select {
			case <-ch:
			case <-ctx.Done():
			}
		}(id)
	}
	wg.Wait()
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

	if !imdbIDPattern.MatchString(imdbID) {
		row.IMDbID = "" // never persist or send anything but a well-formed id
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

func (g *GormStore) FindMany(ctx context.Context, mediaType string, tmdbIDs []int) (map[int]*ExternalRating, error) {
	var rows []ExternalRating
	err := g.db.WithContext(ctx).Where("media_type = ? AND tmdb_id IN ?", mediaType, tmdbIDs).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int]*ExternalRating, len(rows))
	for i := range rows {
		out[rows[i].TMDBID] = &rows[i]
	}
	return out, nil
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
