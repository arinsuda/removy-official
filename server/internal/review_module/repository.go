package review_module

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

var (
	ErrReviewNotFound        = errors.New("review not found")
	ErrCommentNotFound       = errors.New("comment not found")
	ErrForbidden             = errors.New("forbidden")
	ErrAlreadyLiked          = errors.New("already liked")
	ErrNotLiked              = errors.New("not liked")
	ErrAlreadyMarkedHelpful  = errors.New("already marked helpful")
	ErrNotMarkedHelpful      = errors.New("not marked helpful")
	ErrInvalidWatchedAt      = errors.New("invalid watched_at")
	ErrInvalidRating         = errors.New("rating must be between 0.5 and 5 in increments of 0.5")
	ErrInvalidMediaType      = errors.New("media_type must be 'movie' or 'tv'")
	ErrInvalidMediaID        = errors.New("invalid media_id")
	ErrInvalidVisibility     = errors.New("invalid visibility")
	ErrConflictingVisibility = errors.New("conflicting visibility and is_public values")
)

// ReviewFilter ใช้กับ GET /users/:userId/reviews
// Visibility: "all" (default) | "public" | "private"
// DateFrom / DateTo: กรองตาม created_at (วันที่ user เขียนรีวิว), inclusive ทั้งสองด้าน
type ReviewFilter struct {
	Visibility   string
	DateFrom     *time.Time
	DateTo       *time.Time
	Sort         string
	RatingFilter string
	Search       string
	Page         int
	PageSize     int
}

func escapeLikePattern(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(s) + "%"
}

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

// ── Review ────────────────────────────────────────────────────────

func (r *repository) CreateReview(review *Review) error {
	// บันทึกตัวรีวิวเพียวๆ ลงไปก่อน
	return r.db.Create(review).Error
}

// userReviewScope applies the owner/privacy-relevant scope shared by the list and its summary:
// author, visibility and date range. Rating filter, search and paging are NOT part of it.
func (r *repository) userReviewScope(userID uint, filter ReviewFilter) *gorm.DB {
	q := r.db.Model(&Review{}).Where("user_id = ?", userID)

	switch filter.Visibility {
	case "public":
		q = q.Where("visibility = ?", "public")
	case "private":
		q = q.Where("visibility = ?", "private")
	case "followers":
		q = q.Where("visibility = ?", "followers")
	}

	if filter.DateFrom != nil {
		q = q.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		q = q.Where("created_at <= ?", *filter.DateTo)
	}
	return q
}

// userReviewSummaryRow is one grouped-aggregate result over the scope of userReviewScope.
type userReviewSummaryRow struct {
	Count     int
	RatingSum float64
	High      int
	Low       int
	WithText  int
}

// SummarizeReviewsByUser counts the user's reviews in one query, ignoring rating filter, search and paging.
func (r *repository) SummarizeReviewsByUser(userID uint, filter ReviewFilter) (userReviewSummaryRow, error) {
	var row userReviewSummaryRow
	err := r.userReviewScope(userID, filter).
		Select(`COUNT(*) AS count,
			COALESCE(SUM(rating), 0) AS rating_sum,
			COUNT(*) FILTER (WHERE rating >= ?) AS high,
			COUNT(*) FILTER (WHERE rating <= ?) AS low,
			COUNT(*) FILTER (WHERE BTRIM(body) <> '') AS with_text`, HighRatingMin, LowRatingMax).
		Scan(&row).Error
	return row, err
}

func (r *repository) FindReviewsByUser(userID uint, filter ReviewFilter, offset, limit int) ([]Review, int64, error) {
	q := r.userReviewScope(userID, filter)

	switch filter.RatingFilter {
	case ReviewRatingFilterHigh:
		q = q.Where("rating >= ?", HighRatingMin)
	case ReviewRatingFilterLow:
		q = q.Where("rating <= ?", LowRatingMax)
	case ReviewRatingFilterWithText:
		q = q.Where("BTRIM(body) <> ''")
	}

	if filter.Search != "" {
		q = q.Where("body ILIKE ? ESCAPE '\\'", escapeLikePattern(filter.Search))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC, id DESC"
	switch filter.Sort {
	case ReviewSortOldest:
		orderClause = "created_at ASC, id ASC"
	case ReviewSortHighest:
		orderClause = "rating DESC, created_at DESC, id DESC"
	case ReviewSortLowest:
		orderClause = "rating ASC, created_at DESC, id DESC"
	}

	var reviews []Review
	err := q.Preload("User").Order(orderClause).Offset(offset).Limit(limit).Find(&reviews).Error
	return reviews, total, err
}

// FindReviewsByMedia returns one page of the public reviews of a title whose authors the viewer
// may see (privacy_policy.VisibleAuthorScope, one query) plus the total. Guests (viewerID 0)
// additionally need the legacy is_public flag.
func (r *repository) FindReviewsByMedia(mediaID int, mediaType string, viewerID uint, offset, limit int) ([]Review, int64, error) {
	q := r.db.Model(&Review{}).
		Where("reviews.media_id = ? AND reviews.media_type = ? AND reviews.visibility = 'public'", mediaID, mediaType).
		Scopes(privacy_policy.VisibleAuthorScope(viewerID, "reviews.user_id"))
	if viewerID == 0 {
		q = q.Where("reviews.is_public = true")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var reviews []Review
	err := q.Preload("User").Order("reviews.created_at DESC, reviews.id DESC").
		Offset(offset).Limit(limit).Find(&reviews).Error
	return reviews, total, err
}

// FindUserReviewByMedia returns the user's own live review of a title, any visibility; nil when none.
func (r *repository) FindUserReviewByMedia(userID uint, mediaID int, mediaType string) (*Review, error) {
	var review Review
	err := r.db.Preload("User").
		Where("user_id = ? AND media_id = ? AND media_type = ?", userID, mediaID, mediaType).
		First(&review).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &review, err
}

func (r *repository) FindReviewByID(reviewID uint) (*Review, error) {
	var review Review
	// เติม Preload("User") เข้าไปตรงนี้ด้วย เพื่อให้ตอน Update/Delete ทำงานได้ถูกต้อง
	err := r.db.Preload("User").First(&review, reviewID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReviewNotFound
	}
	return &review, err
}

func (r *repository) UpdateReview(reviewID uint, updates map[string]any) error {
	result := r.db.Model(&Review{}).Where("id = ?", reviewID).Updates(updates)
	if result.RowsAffected == 0 {
		return ErrReviewNotFound
	}
	return result.Error
}

func (r *repository) DeleteReview(reviewID uint) error {
	result := r.db.Delete(&Review{}, reviewID)
	if result.RowsAffected == 0 {
		return ErrReviewNotFound
	}
	return result.Error
}

// GetMediaRating คืน average rating และจำนวน review ของ media นั้น
// นับเฉพาะ public reviews และ deleted_at IS NULL (soft-delete safe)
func (r *repository) GetMediaRating(ctx context.Context, id shared.MediaIdentity) (shared.RatingStats, error) {
	type row struct {
		AvgRating   *float64
		ReviewCount int
	}
	var res row
	err := r.db.WithContext(ctx).Model(&Review{}).
		Select("AVG(rating) AS avg_rating, COUNT(*) AS review_count").
		Where("media_id = ? AND media_type = ? AND visibility = 'public' AND deleted_at IS NULL",
			id.ID, string(id.Type)).
		Scan(&res).Error
	return shared.RatingStats{Average: res.AvgRating, Count: res.ReviewCount}, err
}

// GetMediaRatingBuckets returns grouped rating counts for the exact same population as GetMediaRating
// (visibility = 'public' AND deleted_at IS NULL).
func (r *repository) GetMediaRatingBuckets(ctx context.Context, id shared.MediaIdentity) ([]RatingBucket, error) {
	var rows []RatingBucket
	err := r.db.WithContext(ctx).Model(&Review{}).
		Select("rating, COUNT(*) AS count").
		Where("media_id = ? AND media_type = ? AND visibility = 'public' AND deleted_at IS NULL",
			id.ID, string(id.Type)).
		Group("rating").
		Order("rating ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *repository) GetBatchMediaRatings(ctx context.Context, ids []shared.MediaIdentity) (map[shared.MediaKey]shared.RatingStats, error) {
	result := make(map[shared.MediaKey]shared.RatingStats)
	if len(ids) == 0 {
		return result, nil
	}

	q := r.db.WithContext(ctx).Model(&Review{}).
		Select("media_id, media_type, AVG(rating) AS avg_rating, COUNT(*) AS review_count").
		Where("visibility = 'public' AND deleted_at IS NULL").
		Group("media_id, media_type")

	type row struct {
		MediaID     int
		MediaType   string
		AvgRating   *float64
		ReviewCount int
	}

	var conds []string
	var args []any
	for _, id := range ids {
		conds = append(conds, "(media_id = ? AND media_type = ?)")
		args = append(args, id.ID, string(id.Type))
	}

	orCond := strings.Join(conds, " OR ")
	q = q.Where(orCond, args...)

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, id := range ids {
		key := shared.MediaKey{ID: id.ID, Type: id.Type}
		result[key] = shared.RatingStats{Average: nil, Count: 0}
	}

	for _, rw := range rows {
		key := shared.MediaKey{ID: rw.MediaID, Type: shared.MediaType(rw.MediaType)}
		result[key] = shared.RatingStats{Average: rw.AvgRating, Count: rw.ReviewCount}
	}

	return result, nil
}

// ── Like ──────────────────────────────────────────────────────────

func (r *repository) CreateLike(reviewID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		like := ReviewLike{ReviewID: reviewID, UserID: userID}
		if err := tx.Create(&like).Error; err != nil {
			if isDuplicateError(err) {
				return ErrAlreadyLiked
			}
			return err
		}
		return tx.Model(&Review{}).Where("id = ?", reviewID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

func (r *repository) DeleteLike(reviewID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("review_id = ? AND user_id = ?", reviewID, userID).
			Delete(&ReviewLike{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotLiked
		}
		return tx.Model(&Review{}).Where("id = ?", reviewID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
	})
}

func (r *repository) IsLiked(reviewID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ReviewLike{}).
		Where("review_id = ? AND user_id = ?", reviewID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) FindLikedIDs(reviewIDs []uint, userID uint) (map[uint]bool, error) {
	if len(reviewIDs) == 0 {
		return map[uint]bool{}, nil
	}
	var likes []ReviewLike
	err := r.db.Where("review_id IN ? AND user_id = ?", reviewIDs, userID).Find(&likes).Error
	result := make(map[uint]bool)
	for _, l := range likes {
		result[l.ReviewID] = true
	}
	return result, err
}

// ── Helpful ───────────────────────────────────────────────────────
// Pattern เหมือนกับ Like ทุกประการ แยกตารางเพราะคนละความหมาย

func (r *repository) CreateHelpful(reviewID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		vote := ReviewHelpful{ReviewID: reviewID, UserID: userID}
		if err := tx.Create(&vote).Error; err != nil {
			if isDuplicateError(err) {
				return ErrAlreadyMarkedHelpful
			}
			return err
		}
		return tx.Model(&Review{}).Where("id = ?", reviewID).
			UpdateColumn("helpful_count", gorm.Expr("helpful_count + 1")).Error
	})
}

func (r *repository) DeleteHelpful(reviewID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("review_id = ? AND user_id = ?", reviewID, userID).
			Delete(&ReviewHelpful{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotMarkedHelpful
		}
		return tx.Model(&Review{}).Where("id = ?", reviewID).
			UpdateColumn("helpful_count", gorm.Expr("GREATEST(helpful_count - 1, 0)")).Error
	})
}

func (r *repository) IsHelpful(reviewID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ReviewHelpful{}).
		Where("review_id = ? AND user_id = ?", reviewID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) FindHelpfulIDs(reviewIDs []uint, userID uint) (map[uint]bool, error) {
	if len(reviewIDs) == 0 {
		return map[uint]bool{}, nil
	}
	var votes []ReviewHelpful
	err := r.db.Where("review_id IN ? AND user_id = ?", reviewIDs, userID).Find(&votes).Error
	result := make(map[uint]bool)
	for _, v := range votes {
		result[v.ReviewID] = true
	}
	return result, err
}

// ── Comment ───────────────────────────────────────────────────────

func (r *repository) CreateComment(comment *ReviewComment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&Review{}).Where("id = ?", comment.ReviewID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	})
}

func (r *repository) FindCommentsByReview(reviewID uint) ([]ReviewComment, error) {
	var comments []ReviewComment
	err := r.db.Where("review_id = ?", reviewID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *repository) FindCommentByID(commentID uint) (*ReviewComment, error) {
	var comment ReviewComment
	err := r.db.First(&comment, commentID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCommentNotFound
	}
	return &comment, err
}

func (r *repository) UpdateComment(commentID uint, body string) error {
	result := r.db.Model(&ReviewComment{}).Where("id = ?", commentID).Update("body", body)
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return result.Error
}

func (r *repository) DeleteComment(commentID uint, reviewID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ? AND review_id = ?", commentID, reviewID).Delete(&ReviewComment{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrCommentNotFound
		}
		return tx.Model(&Review{}).Where("id = ?", reviewID).
			UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error
	})
}

// ── Helpers ───────────────────────────────────────────────────────

func isDuplicateError(err error) bool {
	return shared.IsPgUniqueViolation(err)
}

func NewRatingStatsReader(db *gorm.DB) shared.RatingStatsReader {
	return &repository{db: db}
}
