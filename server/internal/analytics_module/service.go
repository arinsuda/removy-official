package analytics_module

import (
	"context"
	"math"

	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

const (
	trendingLimit        = 20
	weightWatchlist      = 1.0
	weightFavorite       = 2.0
	weightWatched        = 3.0
	weightRecentActivity = 1.5
)

type Service struct {
	repo *repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{repo: newRepository(db)}
}

// GetMediaAnalytics returns the counts of one title. An adult manga the viewer may not see is refused like any other read.
func (s *Service) GetMediaAnalytics(ctx context.Context, viewerID uint, mediaID int, mediaType string) (*MediaAnalyticsResponse, error) {
	if err := validateMedia(mediaID, mediaType); err != nil {
		return nil, err
	}
	if err := shared.CheckMediaAccess(ctx, viewerID, mediaID, shared.MediaType(mediaType)); err != nil {
		return nil, err
	}

	libStats, err := s.repo.GetLibraryStats(mediaID, mediaType)
	if err != nil {
		return nil, err
	}

	revStats, err := s.repo.GetReviewStats(mediaID, mediaType)
	if err != nil {
		return nil, err
	}

	score := trendingScore(libStats.WatchlistCount, libStats.FavoriteCount, libStats.WatchedCount, 0)

	isTrending, err := s.isInTrending(ctx, viewerID, mediaID, mediaType)
	if err != nil {

		isTrending = false
	}

	avg := float32(0)
	if revStats.ReviewCount > 0 {
		avg = float32(math.Round(float64(revStats.AverageRating)*10) / 10)
	}

	return &MediaAnalyticsResponse{
		MediaID:        mediaID,
		MediaType:      mediaType,
		WatchlistCount: libStats.WatchlistCount,
		FavoriteCount:  libStats.FavoriteCount,
		WatchedCount:   libStats.WatchedCount,
		ReviewCount:    revStats.ReviewCount,
		AverageRating:  avg,
		TrendingScore:  score,
		IsTrending:     isTrending,
	}, nil
}

// GetTrending ranks one media type for the viewer. Adult manga the viewer may not see is left out before the limit.
func (s *Service) GetTrending(ctx context.Context, viewerID uint, mediaType string) ([]TrendingItem, error) {
	if !shared.IsValidMediaType(shared.MediaType(mediaType)) {
		return nil, ErrInvalidMediaType
	}

	rows, err := s.trendingRows(ctx, viewerID, mediaType)
	if err != nil {
		return nil, err
	}

	items := make([]TrendingItem, len(rows))
	for i, r := range rows {
		items[i] = TrendingItem{
			MediaID:        r.MediaID,
			MediaType:      mediaType,
			TrendingScore:  trendingScore(r.WatchlistCount, r.FavoriteCount, r.WatchedCount, r.RecentActivity),
			WatchlistCount: r.WatchlistCount,
			FavoriteCount:  r.FavoriteCount,
			WatchedCount:   r.WatchedCount,
		}
	}
	return items, nil
}

func trendingScore(watchlist, favorite, watched, recent int) float64 {
	return float64(watchlist)*weightWatchlist +
		float64(favorite)*weightFavorite +
		float64(watched)*weightWatched +
		float64(recent)*weightRecentActivity
}

// trendingRows reads the top of the ranking, hiding adult manga from viewers who may not see it.
func (s *Service) trendingRows(ctx context.Context, viewerID uint, mediaType string) ([]trendingRow, error) {
	vis, err := shared.NewMediaVisibility(ctx, viewerID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetTrending(mediaType, trendingLimit, !vis.ShowsAdult())
}

func (s *Service) isInTrending(ctx context.Context, viewerID uint, mediaID int, mediaType string) (bool, error) {
	rows, err := s.trendingRows(ctx, viewerID, mediaType)
	if err != nil {
		return false, err
	}
	for _, r := range rows {
		if r.MediaID == mediaID {
			return true, nil
		}
	}
	return false, nil
}

func validateMedia(mediaID int, mediaType string) error {
	if mediaID <= 0 {
		return ErrInvalidMediaID
	}
	if !shared.IsValidMediaType(shared.MediaType(mediaType)) {
		return ErrInvalidMediaType
	}
	return nil
}
