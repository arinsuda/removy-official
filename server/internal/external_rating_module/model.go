package external_rating_module

import "time"

// ExternalRating caches IMDb and Rotten Tomatoes data (from OMDb) per TMDB title.
// NotFound marks titles OMDb could not resolve so they are retried less eagerly.
type ExternalRating struct {
	ID         uint      `gorm:"primaryKey" json:"-"`
	MediaType  string    `gorm:"size:10;not null;uniqueIndex:idx_external_ratings_media" json:"media_type"`
	TMDBID     int       `gorm:"not null;uniqueIndex:idx_external_ratings_media" json:"tmdb_id"`
	IMDbID     string    `gorm:"column:imdb_id;size:20" json:"imdb_id"`
	IMDbRating *float64  `gorm:"column:imdb_rating" json:"imdb_rating"`
	IMDbVotes  *int      `gorm:"column:imdb_votes" json:"imdb_votes"`
	RTScore    *int      `gorm:"column:rt_score" json:"rt_score"` // Tomatometer percent
	NotFound   bool      `gorm:"not null;default:false" json:"not_found"`
	FetchedAt  time.Time `gorm:"not null" json:"fetched_at"`
}

func (ExternalRating) TableName() string { return "external_ratings" }
