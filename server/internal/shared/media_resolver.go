package shared

import (
	"context"
	"errors"
	"strings"
	"time"

	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

var (
	// ErrAgeRestricted means the title is adult and the viewer may not see it.
	ErrAgeRestricted = errors.New("media: age restricted")
	// ErrMediaNotFound means the provider has no such title.
	ErrMediaNotFound = errors.New("media: not found")
	// ErrMediaUnavailable means the provider could not answer (network, budget, bad payload).
	ErrMediaUnavailable = errors.New("media: unavailable")
	// ErrUnknownMediaType means no provider serves the media type. It is never routed to TV.
	ErrUnknownMediaType = errors.New("media: unknown media type")
)

const (
	msgAgeRestricted    = "adult title is not available to this account"
	msgMediaNotFound    = "media not found"
	msgMediaUnavailable = "media provider unavailable"
)

// MediaInfo is the title data that summaries, notifications and moderation views need.
type MediaInfo struct {
	Title        string
	EnglishTitle string
	ReleaseDate  string             // TMDB release or air date; manga: start year only
	PosterURL    string             // TMDB: raw poster path ("/x.jpg"); manga: absolute URL
	Genres       []tmdbmodule.Genre // TMDB only
	VoteAverage  float32            // TMDB only
	Kind         string             // manga only: manga, manhwa, manhua, novel
	IsAdult      bool               // manga only
}

// MangaLookup resolves one AniList title by its id. manga_module implements it.
type MangaLookup interface {
	LookupMedia(ctx context.Context, id int) (MediaInfo, error)
}

// AdultGate reports whether a viewer may see adult titles. Guests (userID 0) are never eligible.
type AdultGate interface {
	Eligible(ctx context.Context, userID uint) (bool, error)
}

// MediaResolver routes a title lookup by media type: movie and TV go to TMDB, manga goes to AniList.
type MediaResolver struct {
	manga MangaLookup
}

// Resolve returns the title for (id, t). Upstream failures come back as the TMDB error or ErrMediaUnavailable.
func (r MediaResolver) Resolve(ctx context.Context, id int, t MediaType) (MediaInfo, error) {
	switch t {
	case MediaTypeMovie:
		d, err := tmdbmodule.GetMovieByID(id)
		if err != nil {
			return MediaInfo{}, err
		}
		if d == nil {
			return MediaInfo{}, ErrMediaNotFound
		}
		return MediaInfo{
			Title: d.Title, EnglishTitle: d.EnglishTitle, ReleaseDate: d.ReleaseDate,
			PosterURL: d.PosterPath, Genres: d.Genres, VoteAverage: float32(d.VoteAverage),
		}, nil
	case MediaTypeTV:
		d, err := tmdbmodule.GetSeriesByID(id)
		if err != nil {
			return MediaInfo{}, err
		}
		if d == nil {
			return MediaInfo{}, ErrMediaNotFound
		}
		return MediaInfo{
			Title: d.Name, EnglishTitle: d.EnglishTitle, ReleaseDate: d.FirstAirDate,
			PosterURL: d.PosterPath, Genres: d.Genres, VoteAverage: float32(d.VoteAverage),
		}, nil
	case MediaTypeManga:
		if r.manga == nil {
			return MediaInfo{}, ErrMediaUnavailable
		}
		return r.manga.LookupMedia(ctx, id)
	}
	return MediaInfo{}, ErrUnknownMediaType
}

// mediaResolver and adultGate are wired once at boot by SetMangaSupport, before any request is served.
var (
	mediaResolver MediaResolver
	adultGate     AdultGate
)

// SetMangaSupport wires the AniList lookup, the adult gate and the manga_meta store. Until it runs, manga
// lookups report ErrMediaUnavailable, adult checks treat every viewer as ineligible and nothing is stored.
// A nil db disables manga_meta reads and writes (tests).
func SetMangaSupport(lookup MangaLookup, gate AdultGate, db *gorm.DB) {
	mediaResolver = MediaResolver{manga: lookup}
	adultGate = gate
	mangaDB = db
}

// ResolveMedia looks a title up through the wired resolver.
func ResolveMedia(ctx context.Context, id int, t MediaType) (MediaInfo, error) {
	return mediaResolver.Resolve(ctx, id, t)
}

// MediaPosterURL makes a poster absolute: manga URLs pass through, TMDB paths get the TMDB image base.
func MediaPosterURL(p string) string {
	if p == "" || strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p
	}
	return tmdbmodule.ImageURL(p)
}

func viewerEligible(ctx context.Context, viewerID uint) (bool, error) {
	if adultGate == nil {
		return false, nil
	}
	return adultGate.Eligible(ctx, viewerID)
}

// CheckMediaAccess is the write-side and public-read gate: a viewer may act on a title only if it is not
// adult or the viewer is eligible. Non-manga types always pass. Manga reads manga_meta first and asks AniList
// only on a miss. An unresolvable manga id is refused (fail closed).
func CheckMediaAccess(ctx context.Context, viewerID uint, id int, t MediaType) error {
	if t != MediaTypeManga {
		return nil
	}
	adult, err := mangaIsAdult(ctx, id)
	if err != nil {
		return err
	}
	if !adult {
		return nil
	}
	eligible, err := viewerEligible(ctx, viewerID)
	if err != nil {
		return err
	}
	if !eligible {
		return ErrAgeRestricted
	}
	return nil
}

// mangaIsAdult answers from the stored row while it is fresh. On a miss or a stale row it resolves the title
// and stores it, so the next request needs no AniList call. A stale row keeps its stored answer when AniList
// cannot answer now; a missing row does not, so the error goes back to the caller (404 or 503 stay distinct).
func mangaIsAdult(ctx context.Context, id int) (bool, error) {
	stored, err := loadStoredMangaMeta(ctx, []int{id})
	if err != nil {
		return false, err
	}
	row, has := stored[id]
	if has && !row.isStale(time.Now()) {
		return storedAdult(row)
	}
	fresh, err := resolveMangaMeta(ctx, id)
	if err != nil {
		if has {
			return storedAdult(row)
		}
		return false, err
	}
	return storedAdult(fresh)
}

// storedAdult reads the adult flag of a row. The not-found marker answers like an unknown title (404).
func storedAdult(row MangaMeta) (bool, error) {
	if row.isNotFound() {
		return false, ErrMediaNotFound
	}
	return row.IsAdult, nil
}

// WriteMediaAccessError answers a failed media lookup or access check with its error code.
// handled is false for any other error, which the caller must map itself.
func WriteMediaAccessError(c fiber.Ctx, err error) (handled bool, resp error) {
	switch {
	case errors.Is(err, ErrAgeRestricted):
		return true, WriteAPIError(c, fiber.StatusForbidden, ErrorCodeAgeRestricted, msgAgeRestricted)
	case errors.Is(err, ErrMediaNotFound):
		return true, WriteAPIError(c, fiber.StatusNotFound, ErrorCodeNotFound, msgMediaNotFound)
	case errors.Is(err, ErrMediaUnavailable):
		return true, WriteAPIError(c, fiber.StatusServiceUnavailable, ErrorCodeUpstreamUnavailable, msgMediaUnavailable)
	}
	return false, nil
}

// MediaVisibility decides which adult titles one viewer may see inside other users' data.
// The zero value hides adult titles, so a forgotten constructor fails closed.
type MediaVisibility struct {
	viewerID  uint
	showAdult bool
}

// NewMediaVisibility asks the age gate once for the viewer. A gate error is returned, never read as eligible.
func NewMediaVisibility(ctx context.Context, viewerID uint) (MediaVisibility, error) {
	eligible, err := viewerEligible(ctx, viewerID)
	if err != nil {
		return MediaVisibility{}, err
	}
	return MediaVisibility{viewerID: viewerID, showAdult: eligible}, nil
}

// ShowsAdult reports whether the viewer is eligible for adult titles everywhere (no filtering needed).
func (v MediaVisibility) ShowsAdult() bool {
	return v.showAdult
}

// Restricted reports whether adult manga must be filtered out of ownerID's data for this viewer. The owner
// and eligible viewers are never restricted. Callers use it to add MangaSafeSQL to count queries.
func (v MediaVisibility) Restricted(ownerID uint) bool {
	return !v.showAdult && !v.isOwner(ownerID)
}

func (v MediaVisibility) isOwner(ownerID uint) bool {
	return v.viewerID != 0 && ownerID == v.viewerID
}

// KeepVisible returns the items the viewer may see, in order. ref names each item's title and owner.
// Items the viewer owns always stay. Other manga stays only when its stored row is known and not adult:
// one batched manga_meta read, plus the bounded AniList fallback for rows that are missing.
// The input is not modified.
func KeepVisible[T any](ctx context.Context, v MediaVisibility, items []T, ref func(T) VisibleRef) ([]T, error) {
	out := make([]T, 0, len(items))
	if v.showAdult {
		return append(out, items...), nil
	}
	refs := make([]VisibleRef, len(items))
	var ids []int
	for i, it := range items {
		refs[i] = ref(it)
		if refs[i].Media.Type == MediaTypeManga && !v.isOwner(refs[i].OwnerID) {
			ids = append(ids, refs[i].Media.ID)
		}
	}
	rows, err := MangaRows(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i, it := range items {
		r := refs[i]
		if r.Media.Type == MediaTypeManga && !v.isOwner(r.OwnerID) && mangaHidden(rows, r.Media.ID) {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}
