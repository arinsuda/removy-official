package shared

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// MangaMetaFallbackLimit caps the AniList lookups one request may make for manga that are missing or stale in manga_meta.
	// Titles past the cap keep their stored row, or stay unknown: hidden from viewers who are not eligible, never hidden from their owner.
	MangaMetaFallbackLimit = 8
	// mangaMetaFallbackWorkers bounds the concurrent AniList lookups of one request.
	mangaMetaFallbackWorkers = 4
	// MangaMetaRefreshAfter is the age after which a stored title is re-resolved on read. Adult flags can change
	// upstream, so a stale row is refreshed; while the refresh fails, its stored flags still apply.
	MangaMetaRefreshAfter = 24 * time.Hour
	// MangaMetaNotFoundRecheckAfter is how long an AniList not-found marker is trusted before the id is looked up again.
	MangaMetaNotFoundRecheckAfter = 6 * time.Hour
	// MangaMetaMaxGenres caps the genre names stored per title; AniList can return more.
	MangaMetaMaxGenres = 8
	// tmdbSummaryConcurrency bounds the concurrent TMDB lookups of one MediaSummaries call (matches bmol_module).
	tmdbSummaryConcurrency = 6
)

// MangaMeta is the stored title data of one AniList manga. Every manga write that resolves the title
// stores it, and lists, feeds, counts and notifications read it instead of calling AniList. It has no user FK.
type MangaMeta struct {
	MediaID   int       `gorm:"primaryKey;autoIncrement:false;column:media_id"`
	Title     string    `gorm:"type:text;not null;default:''"`
	CoverURL  string    `gorm:"type:text;not null;default:''"`
	Kind      string    `gorm:"type:varchar(10);not null;default:''"`
	IsAdult   bool      `gorm:"not null;default:false"`
	StartYear int       `gorm:"not null;default:0"`
	Genres    []string  `gorm:"type:jsonb;serializer:json;not null;default:'[]'"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (MangaMeta) TableName() string { return "manga_meta" }

// isNotFound reports whether the row is the marker for an id AniList does not know: adult and untitled,
// so every read hides it and every write needs an eligible viewer.
func (m MangaMeta) isNotFound() bool { return m.Title == "" && m.IsAdult }

// isStale reports whether the row is old enough that its flags should be read from AniList again.
func (m MangaMeta) isStale(now time.Time) bool {
	ttl := MangaMetaRefreshAfter
	if m.isNotFound() {
		ttl = MangaMetaNotFoundRecheckAfter
	}
	return now.Sub(m.UpdatedAt) > ttl
}

// mangaDB is the store behind manga_meta. Nil (tests without a database) turns every read into a miss and every write into a no-op.
var mangaDB *gorm.DB

// VisibleRef names the title one item shows and the user the item belongs to. OwnerID 0 means the item is
// not one user's own (for example a notification about someone else's action), so no owner exemption applies.
type VisibleRef struct {
	Media   MediaKey
	OwnerID uint
}

func newMangaMeta(id int, info MediaInfo) MangaMeta {
	genres := info.GenreNames
	if len(genres) > MangaMetaMaxGenres {
		genres = genres[:MangaMetaMaxGenres]
	}
	if genres == nil {
		genres = []string{} // stored as a JSON array, never null
	}
	return MangaMeta{
		MediaID:   id,
		Title:     info.Title,
		CoverURL:  info.PosterURL,
		Kind:      info.Kind,
		IsAdult:   info.IsAdult,
		StartYear: info.StartYear,
		Genres:    genres,
		UpdatedAt: time.Now(),
	}
}

// notFoundMangaMeta is the marker stored for an id AniList does not know. See MangaMeta.isNotFound.
func notFoundMangaMeta(id int) MangaMeta {
	return MangaMeta{MediaID: id, IsAdult: true, Genres: []string{}, UpdatedAt: time.Now()}
}

// storeMangaMeta upserts one row.
func storeMangaMeta(ctx context.Context, row MangaMeta) error {
	if mangaDB == nil {
		return nil
	}
	return mangaDB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "media_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "cover_url", "kind", "is_adult", "start_year", "genres", "updated_at"}),
	}).Create(&row).Error
}

// RecordMangaMeta stores the title data of a manga detail the service just fetched, so later reads use the
// stored flags. A failed write is logged, not returned: the read itself succeeded.
func RecordMangaMeta(ctx context.Context, id int, info MediaInfo) {
	if err := storeMangaMeta(ctx, newMangaMeta(id, info)); err != nil {
		log.Printf("manga_meta: record %d failed: %v", id, err)
	}
}

// loadStoredMangaMeta reads the stored rows for ids in one query. Ids without a row are absent from the map.
func loadStoredMangaMeta(ctx context.Context, ids []int) (map[int]MangaMeta, error) {
	rows := make(map[int]MangaMeta, len(ids))
	if mangaDB == nil || len(ids) == 0 {
		return rows, nil
	}
	var found []MangaMeta
	if err := mangaDB.WithContext(ctx).Where("media_id IN ?", ids).Find(&found).Error; err != nil {
		return nil, err
	}
	for _, r := range found {
		rows[r.MediaID] = r
	}
	return rows, nil
}

// resolveMangaMeta asks AniList for one title and stores the answer. An id AniList does not know is stored as
// the not-found marker and returned without error. An upstream failure returns an error and stores nothing.
func resolveMangaMeta(ctx context.Context, id int) (MangaMeta, error) {
	info, err := ResolveMedia(ctx, id, MediaTypeManga)
	var row MangaMeta
	switch {
	case errors.Is(err, ErrMediaNotFound):
		row = notFoundMangaMeta(id)
	case err != nil:
		return MangaMeta{}, err
	default:
		row = newMangaMeta(id, info)
	}
	if err := storeMangaMeta(ctx, row); err != nil {
		log.Printf("manga_meta: store %d failed: %v", id, err)
	}
	return row, nil
}

// refreshMangaRows resolves ids from AniList with bounded concurrency. Ids that cannot be answered are absent
// from the result, so the caller keeps any stored row for them.
func refreshMangaRows(ctx context.Context, ids []int) map[int]MangaMeta {
	fresh := make([]*MangaMeta, len(ids))
	var g errgroup.Group
	g.SetLimit(mangaMetaFallbackWorkers)
	for i, id := range ids {
		g.Go(func() error {
			if row, err := resolveMangaMeta(ctx, id); err == nil {
				fresh[i] = &row
			}
			return nil
		})
	}
	_ = g.Wait()
	out := make(map[int]MangaMeta, len(ids))
	for i, row := range fresh {
		if row != nil {
			out[ids[i]] = *row
		}
	}
	return out
}

// MangaRows returns the manga_meta row of every id. Ids with no row, or with a stale one, get one bounded AniList
// lookup each (at most MangaMetaFallbackLimit per call). A stale row whose refresh fails keeps its stored flags;
// an id with no row that cannot be resolved is absent, so the caller fails closed.
func MangaRows(ctx context.Context, ids []int) (map[int]MangaMeta, error) {
	ids = uniqueInts(ids)
	rows, err := loadStoredMangaMeta(ctx, ids)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var refresh []int
	for _, id := range ids {
		if row, ok := rows[id]; !ok || row.isStale(now) {
			refresh = append(refresh, id)
		}
	}
	if len(refresh) > MangaMetaFallbackLimit {
		refresh = refresh[:MangaMetaFallbackLimit]
	}
	for id, row := range refreshMangaRows(ctx, refresh) {
		rows[id] = row
	}
	return rows, nil
}

// mangaHidden reports whether a manga title must be left out of an ineligible viewer's data: adult, or unknown.
func mangaHidden(rows map[int]MangaMeta, id int) bool {
	row, ok := rows[id]
	return !ok || row.IsAdult
}

// MangaSafeSQL is a WHERE predicate that keeps every non-manga row and the manga rows whose manga_meta row
// is not adult. Unknown manga is excluded, which matches mangaHidden. The arguments are trusted column
// names from code, never user input.
func MangaSafeSQL(mediaTypeCol, mediaIDCol string) string {
	return fmt.Sprintf(
		"(COALESCE(%[1]s, '') <> 'manga' OR %[2]s IN (SELECT media_id FROM manga_meta WHERE is_adult = false))",
		mediaTypeCol, mediaIDCol,
	)
}

// ReceivableMedia reports, per title, whether a notification about it may reach userID. Only manga is
// restricted: an adult or unknown title needs an eligible user. userID 0 asks whether titles are safe for
// everyone, which is how follower fan-out decides (guests are never eligible). One batched manga_meta read.
func ReceivableMedia(ctx context.Context, userID uint, keys []MediaKey) (map[MediaKey]bool, error) {
	var ids []int
	for _, k := range keys {
		if k.Type == MediaTypeManga {
			ids = append(ids, k.ID)
		}
	}
	rows, err := MangaRows(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[MediaKey]bool, len(keys))
	eligible, checked := false, false
	for _, k := range keys {
		if k.Type != MediaTypeManga || !mangaHidden(rows, k.ID) {
			out[k] = true
			continue
		}
		if !checked {
			if eligible, err = viewerEligible(ctx, userID); err != nil {
				return nil, err
			}
			checked = true
		}
		out[k] = eligible
	}
	return out, nil
}

// MediaSummaries builds the summary of every key. Manga titles come from one manga_meta read (plus the
// bounded fallback); movie and TV summaries are fetched concurrently. Keys the provider cannot answer get an empty summary.
func MediaSummaries(ctx context.Context, keys []MediaKey) (map[MediaKey]MediaSummary, error) {
	var mangaIDs []int
	var others []MediaKey
	seen := make(map[MediaKey]struct{}, len(keys))
	for _, k := range keys {
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		if k.Type == MediaTypeManga {
			mangaIDs = append(mangaIDs, k.ID)
		} else {
			others = append(others, k)
		}
	}
	rows, err := MangaRows(ctx, mangaIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[MediaKey]MediaSummary, len(keys))
	for _, k := range keys {
		if k.Type != MediaTypeManga {
			continue
		}
		s := MediaSummary{ID: k.ID, MediaType: k.Type}
		if row, ok := rows[k.ID]; ok {
			s.Title = row.Title
			s.PosterURL = row.CoverURL
			s.Kind = row.Kind
			if !row.isNotFound() {
				s.StartYear = row.StartYear
				s.GenreNames = row.Genres
				s.IsAdult = row.IsAdult
			}
		}
		out[k] = s
	}
	for i, s := range fetchSummaries(others) {
		out[others[i]] = s
	}
	return out, nil
}

// fetchSummaries builds the movie and TV summaries of keys, at most tmdbSummaryConcurrency at a time.
func fetchSummaries(keys []MediaKey) []MediaSummary {
	out := make([]MediaSummary, len(keys))
	var g errgroup.Group
	g.SetLimit(tmdbSummaryConcurrency)
	for i, k := range keys {
		g.Go(func() error {
			out[i] = fetchSummary(k.ID, k.Type)
			return nil
		})
	}
	_ = g.Wait()
	return out
}

// uniqueInts keeps the first occurrence of each id, so one title is looked up at most once.
func uniqueInts(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}
