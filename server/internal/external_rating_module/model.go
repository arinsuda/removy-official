package external_rating_module

import "time"

// ExternalRating caches IMDb and Rotten Tomatoes data (from OMDb) per TMDB title.
// NotFound marks titles OMDb could not resolve so they are retried less eagerly.
type ExternalRating struct {
	ID         uint      `gorm:"primaryKey" json:"-"`
	MediaType  string    `gorm:"size:10;not null;uniqueIndex:idx_external_ratings_media" json:"media_type"`
	TMDBID     int       `gorm:"not null;uniqueIndex:idx_external_ratings_media" json:"tmdb_id"`
	IMDbID     string    `gorm:"size:20" json:"imdb_id"`
	IMDbRating *float64  `json:"imdb_rating"`
	IMDbVotes  *int      `json:"imdb_votes"`
	RTScore    *int      `json:"rt_score"` // Tomatometer percent
	NotFound   bool      `gorm:"not null;default:false" json:"not_found"`
	FetchedAt  time.Time `gorm:"not null" json:"fetched_at"`
}

func (ExternalRating) TableName() string { return "external_ratings" }
