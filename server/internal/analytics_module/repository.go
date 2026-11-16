package analytics_module

import (
	"errors"
	"time"

	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

var (
	ErrInvalidMediaType = errors.New("media_type must be 'movie', 'tv' or 'manga'")
	ErrInvalidMediaID   = errors.New("invalid media_id")
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

type libraryStatsRow struct {
	WatchlistCount int
	FavoriteCount  int
	WatchedCount   int
}

type reviewStatsRow struct {
	ReviewCount   int
	AverageRating float32
}

type trendingRow struct {
	MediaID        int
	WatchlistCount int
	FavoriteCount  int
	WatchedCount   int
	RecentActivity int
}

func (r *repository) GetLibraryStats(mediaID int, mediaType string) (*libraryStatsRow, error) {
	type row struct {
		ListType string
		Count    int
	}
	var rows []row

	err := r.db.Table("library_items").
		Select("list_type, COUNT(*) AS count").
		Where("media_id = ? AND media_type = ? AND deleted_at IS NULL", mediaID, mediaType).
		Group("list_type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := &libraryStatsRow{}
	for _, row := range rows {
		switch row.ListType {
		case "watchlist":
			result.WatchlistCount = row.Count
		case "favorite":
			result.FavoriteCount = row.Count
		case "watched":
			result.WatchedCount = row.Count
		}
	}
	return result, nil
}

func (r *repository) GetReviewStats(mediaID int, mediaType string) (*reviewStatsRow, error) {
	var result reviewStatsRow
	err := r.db.Table("reviews").
		Select("COUNT(*) AS review_count, COALESCE(AVG(rating), 0) AS average_rating").
		Where("media_id = ? AND media_type = ? AND is_public = true AND deleted_at IS NULL",
			mediaID, mediaType).
		Scan(&result).Error
	return &result, err
}

// GetTrending returns the top rows of one media type. hideAdult drops adult manga (and unknown manga) before the limit.
func (r *repository) GetTrending(mediaType string, limit int, hideAdult bool) ([]trendingRow, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	recentSQL := r.db.Table("library_items").
		Select("media_id, COUNT(*) AS recent_activity").
		Where("media_type = ? AND deleted_at IS NULL AND created_at >= ?", mediaType, sevenDaysAgo).
		Group("media_id")

	q := r.db.Table("library_items AS li")
	if hideAdult {
		q = q.Where(shared.MangaSafeSQL("li.media_type", "li.media_id"))
	}
	var rows []trendingRow
	err := q.
		Select(`
			li.media_id,
			SUM(CASE WHEN li.list_type = 'watchlist' THEN 1 ELSE 0 END) AS watchlist_count,
			SUM(CASE WHEN li.list_type = 'favorite'  THEN 1 ELSE 0 END) AS favorite_count,
			SUM(CASE WHEN li.list_type = 'watched'   THEN 1 ELSE 0 END) AS watched_count,
			COALESCE(r.recent_activity, 0)                               AS recent_activity
		`).
		Joins("LEFT JOIN (?) AS r ON r.media_id = li.media_id", recentSQL).
		Where("li.media_type = ? AND li.deleted_at IS NULL", mediaType).
		Group("li.media_id, r.recent_activity").
		Order(`
			(SUM(CASE WHEN li.list_type = 'watchlist' THEN 1 ELSE 0 END) * 1.0
			+ SUM(CASE WHEN li.list_type = 'favorite'  THEN 1 ELSE 0 END) * 2.0
			+ SUM(CASE WHEN li.list_type = 'watched'   THEN 1 ELSE 0 END) * 3.0
			+ COALESCE(r.recent_activity, 0) * 1.5) DESC
		`).
		Limit(limit).
		Scan(&rows).Error

	return rows, err
}
