package shared

import (
	"context"

	"gorm.io/gorm"
)

// SummaryRatingSource mirrors the {average,count,available,scale} shape used by movie_module.
type SummaryRatingSource struct {
	Average   float64 `json:"average"`
	Count     int     `json:"count"`
	Available bool    `json:"available"`
	Scale     float64 `json:"scale"`
}

// SummaryRatings is attached to MediaSummary so cards can show the POPCORN score.
type SummaryRatings struct {
	POPCORN SummaryRatingSource `json:"popcorn"`
	TMDB    SummaryRatingSource `json:"tmdb"`
}

// AttachRatings fills Ratings on every summary with one grouped query per media type
// (public, non-deleted reviews only; same rules as the review rating stats).
// A title without reviews gets popcorn average 0 and available=false.
func AttachRatings(ctx context.Context, db *gorm.DB, items []*MediaSummary) error {
	if len(items) == 0 {
		return nil
	}
	idsByType := map[MediaType][]int{}
	seen := map[MediaKey]struct{}{}
	for _, it := range items {
		k := MediaKey{ID: it.ID, Type: it.MediaType}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		idsByType[it.MediaType] = append(idsByType[it.MediaType], it.ID)
	}

	type row struct {
		MediaID     int
		AvgRating   float64
		ReviewCount int
	}
	stats := map[MediaKey]row{}
	for t, ids := range idsByType {
		var rows []row
		err := db.WithContext(ctx).Table("reviews").
			Select("media_id, AVG(rating) AS avg_rating, COUNT(*) AS review_count").
			Where("media_type = ? AND media_id IN ? AND visibility = 'public' AND deleted_at IS NULL", string(t), ids).
			Group("media_id").Scan(&rows).Error
		if err != nil {
			return err
		}
		for _, r := range rows {
			stats[MediaKey{ID: r.MediaID, Type: t}] = r
		}
	}

	for _, it := range items {
		s := stats[MediaKey{ID: it.ID, Type: it.MediaType}]
		it.Ratings = &SummaryRatings{
			POPCORN: SummaryRatingSource{Average: s.AvgRating, Count: s.ReviewCount, Available: s.ReviewCount > 0, Scale: 5},
			TMDB:    SummaryRatingSource{Average: float64(it.VoteAverage), Available: it.VoteAverage > 0, Scale: 10},
		}
	}
	return nil
}
