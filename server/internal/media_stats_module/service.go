package media_stats_module

import (
	"math"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/movie_module"
	"gorm.io/gorm"
)

type Service struct {
	repo *repository
	now  func() time.Time
}

func NewService(db *gorm.DB) *Service {
	return &Service{repo: newRepository(db), now: time.Now}
}

// GetStats รวม stats จากหลาย source
// requesterID = 0 → anonymous (IsLiked = false เสมอ)
func (s *Service) GetStats(mediaID int, mediaType movie_module.MediaType, requesterID uint) (*StatsResponse, error) {
	viewCount, err := s.repo.GetViewCount(mediaID, mediaType)
	if err != nil {
		return nil, err
	}

	likeCount, err := s.repo.CountLikes(mediaID, mediaType)
	if err != nil {
		return nil, err
	}

	reviewCount, err := s.repo.CountReviews(mediaID, mediaType)
	if err != nil {
		return nil, err
	}

	watchlistCount, err := s.repo.CountWatchlist(mediaID, mediaType)
	if err != nil {
		return nil, err
	}

	// ── ใหม่ ──────────────────────────────────────────────────────
	ratingRow, err := s.repo.GetAverageRating(mediaID, mediaType)
	if err != nil {
		return nil, err
	}
	// round ทศนิยม 1 ตำแหน่ง เช่น 3.75 → 3.8
	avgRating := float32(math.Round(float64(ratingRow.AvgRating)*10) / 10)
	// ─────────────────────────────────────────────────────────────

	var likedAt *time.Time
	if requesterID > 0 {
		likedAt, err = s.repo.GetLikedAt(requesterID, mediaID, mediaType)
		if err != nil {
			return nil, err
		}
	}

	var watchlistedAt *time.Time
	if requesterID > 0 {
		watchlistedAt, err = s.repo.GetWatchlistedAt(requesterID, mediaID, mediaType)
		if err != nil {
			return nil, err
		}
	}

	return &StatsResponse{
		MediaID:        mediaID,
		MediaType:      mediaType,
		LikeCount:      likeCount,
		ViewCount:      viewCount,
		ReviewCount:    reviewCount,
		WatchlistCount: watchlistCount,
		AverageRating:  avgRating,
		HasRating:      ratingRow.ReviewCount > 0,
		LikedAt:        likedAt,
		WatchlistedAt:  watchlistedAt,
	}, nil
}

// RecordView บันทึกการเข้าชม: the total always goes up, while the per-day row that feeds voting
// is written once per user, title and Bangkok day.
func (s *Service) RecordView(userID uint, mediaID int, mediaType movie_module.MediaType) error {
	if err := s.repo.UpsertStat(mediaID, mediaType); err != nil {
		return err
	}
	if err := s.repo.InsertViewDay(userID, mediaID, mediaType, config.LocalDay(s.now())); err != nil {
		return err
	}
	return s.repo.IncrementView(mediaID, mediaType)
}

// LikeMedia เพิ่ม like
func (s *Service) LikeMedia(userID uint, mediaID int, mediaType movie_module.MediaType) error {
	return s.repo.CreateLike(userID, mediaID, mediaType)
}

// UnlikeMedia ลบ like
func (s *Service) UnlikeMedia(userID uint, mediaID int, mediaType movie_module.MediaType) error {
	return s.repo.DeleteLike(userID, mediaID, mediaType)
}
