package review_module

import (
	"context"
	"math"

	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	users "github.com/arinsuda/movie-hub/internal/user_module"
	stats "github.com/arinsuda/movie-hub/internal/user_stats_module"
	"gorm.io/gorm"
)

type Service struct {
	repo       *repository
	db         *gorm.DB
	minio      *storage.MinIOClient
	expPort    stats.ExpAdder
	achieveSvc achievementsmodule.Service
	notifSvc   *notification_module.Service
	feedSvc    feed_module.Service
	policy     privacy_policy.UserAccessPolicy
}

func NewService(
	db *gorm.DB,
	mc *storage.MinIOClient,
	exp stats.ExpAdder,
	achieve achievementsmodule.Service,
	notif *notification_module.Service,
	feed feed_module.Service,
	policy privacy_policy.UserAccessPolicy,
) *Service {
	return &Service{
		repo:       newRepository(db),
		db:         db,
		minio:      mc,
		expPort:    exp,
		achieveSvc: achieve,
		notifSvc:   notif,
		feedSvc:    feed,
		policy:     policy,
	}
}

func resolveVisibility(visibility string, isPublic *bool) (string, error) {
	if visibility != "" {
		if visibility != "public" && visibility != "followers" && visibility != "private" {
			return "", ErrInvalidVisibility
		}
		if isPublic != nil {
			if (*isPublic && visibility != "public") || (!*isPublic && visibility == "public") {
				return "", ErrConflictingVisibility
			}
		}
		return visibility, nil
	}
	if isPublic != nil {
		if *isPublic {
			return "public", nil
		}
		return "private", nil
	}
	return "public", nil
}

func (s *Service) CreateReview(ctx context.Context, userID uint, req CreateReviewRequest) (*ReviewResponse, error) {
	if err := validateReviewRequest(req); err != nil {
		return nil, err
	}

	vis, err := resolveVisibility(req.Visibility, req.IsPublic)
	if err != nil {
		return nil, err
	}

	review := &Review{
		UserID:     userID,
		MediaID:    req.MediaID,
		MediaType:  req.MediaType,
		Rating:     req.Rating,
		Body:       req.Body,
		Visibility: vis,
		IsPublic:   vis == "public",
	}

	if err := s.repo.CreateReview(review); err != nil {
		return nil, err
	}

	_ = s.expPort.AddExperience(userID, stats.ExpPerReview)

	var reviewCount int64
	s.db.WithContext(ctx).Model(&Review{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&reviewCount)
	unlocked := shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "review_count", int(reviewCount))
	s.createAchievementFeedActivities(ctx, userID, unlocked)

	if req.Rating <= 1.0 {
		var oneStarCount int64
		s.db.WithContext(ctx).Model(&Review{}).
			Where("user_id = ? AND rating <= 1.0 AND deleted_at IS NULL", userID).
			Count(&oneStarCount)
		unlocked = shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "rating_one_star_count", int(oneStarCount))
		s.createAchievementFeedActivities(ctx, userID, unlocked)
	}
	if req.Rating == 5.0 {
		var fiveStarCount int64
		s.db.WithContext(ctx).Model(&Review{}).
			Where("user_id = ? AND rating = 5.0 AND deleted_at IS NULL", userID).
			Count(&fiveStarCount)
		unlocked = shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "rating_five_star_count", int(fiveStarCount))
		s.createAchievementFeedActivities(ctx, userID, unlocked)
	}
	if req.Rating < 3.0 {
		var lowRatingCount int64
		s.db.WithContext(ctx).Model(&Review{}).
			Where("user_id = ? AND rating < 3.0 AND deleted_at IS NULL", userID).
			Count(&lowRatingCount)
		unlocked = shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "low_rating_count", int(lowRatingCount))
		s.createAchievementFeedActivities(ctx, userID, unlocked)
	}

	if vis == "public" && s.notifSvc != nil {
		if actor, err := s.getUserSummary(userID); err == nil {
			title, _ := fetchMediaSummary(req.MediaID, req.MediaType)
			_ = s.notifSvc.PushFollowingReviewed(
				ctx,
				userID,
				actor.Username,
				review.ID,
				title,
			)
		}
	}

	if (vis == "public" || vis == "followers") && s.feedSvc != nil {
		_ = s.feedSvc.CreateActivity(ctx, userID, feed_module.ActivityReviewCreated, feed_module.ActivityPayload{
			MediaID:   &req.MediaID,
			MediaType: &req.MediaType,
			ReviewID:  &review.ID,
			Message:   review.Body,
		})
	}

	inserted, err := s.repo.FindReviewByID(review.ID)
	if err != nil {
		return nil, err
	}
	return toReviewResponse(inserted, false, false, s.minio), nil
}

func (s *Service) GetUserReviews(ctx context.Context, userID, requesterID uint, filter ReviewFilter) ([]ReviewResponse, Pagination, UserReviewsSummary, error) {
	filter.Page = clampPage(filter.Page)
	if filter.PageSize < 1 {
		filter.PageSize = DefaultMediaReviewsPageSize
	}
	if filter.PageSize > MaxMediaReviewsPageSize {
		filter.PageSize = MaxMediaReviewsPageSize
	}
	pg := Pagination{Page: filter.Page, PageSize: filter.PageSize}

	canView, err := s.policy.CanViewProfileSection(ctx, requesterID, userID, privacy_policy.SectionReviews)
	if err != nil {
		return nil, pg, UserReviewsSummary{}, err
	}
	if !canView {
		return nil, pg, UserReviewsSummary{}, ErrForbidden
	}

	if userID != requesterID {
		filter.Visibility = "public"
	}

	reviews, total, err := s.repo.FindReviewsByUser(userID, filter, (filter.Page-1)*filter.PageSize, filter.PageSize)
	if err != nil {
		return nil, pg, UserReviewsSummary{}, err
	}
	summaryRow, err := s.repo.SummarizeReviewsByUser(userID, filter)
	if err != nil {
		return nil, pg, UserReviewsSummary{}, err
	}
	pg.Total = total
	pg.HasMore = int64(filter.Page*filter.PageSize) < total

	ids := make([]uint, len(reviews))
	for i, r := range reviews {
		ids[i] = r.ID
	}

	likedMap, _ := s.repo.FindLikedIDs(ids, requesterID)
	helpfulMap, _ := s.repo.FindHelpfulIDs(ids, requesterID)

	responses := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		responses[i] = *toReviewResponse(&r, likedMap[r.ID], helpfulMap[r.ID], s.minio)
	}
	return responses, pg, buildUserReviewsSummary(summaryRow), nil
}

// buildUserReviewsSummary turns the grouped aggregate row into the API summary (average to one decimal).
func buildUserReviewsSummary(row userReviewSummaryRow) UserReviewsSummary {
	avg := float32(0)
	if row.Count > 0 {
		avg = float32(math.Round(row.RatingSum/float64(row.Count)*10) / 10)
	}
	return UserReviewsSummary{
		Count:         row.Count,
		AverageRating: avg,
		FilterCounts:  ReviewFilterCounts{All: row.Count, High: row.High, Low: row.Low, WithText: row.WithText},
	}
}

// GetMediaReviews returns one page of a title's reviews. page and pageSize are normalised
// (page >= 1, pageSize defaults to DefaultMediaReviewsPageSize and is capped at MaxMediaReviewsPageSize).
func (s *Service) GetMediaReviews(ctx context.Context, mediaID int, mediaType string, requesterID uint, page, pageSize int) ([]ReviewResponse, Pagination, error) {
	page = clampPage(page)
	if pageSize < 1 {
		pageSize = DefaultMediaReviewsPageSize
	}
	if pageSize > MaxMediaReviewsPageSize {
		pageSize = MaxMediaReviewsPageSize
	}
	pg := Pagination{Page: page, PageSize: pageSize}

	reviews, total, err := s.repo.FindReviewsByMedia(mediaID, mediaType, requesterID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, pg, err
	}
	pg.Total = total
	pg.HasMore = int64(page*pageSize) < total

	ids := make([]uint, len(reviews))
	for i, r := range reviews {
		ids[i] = r.ID
	}
	likedMap, _ := s.repo.FindLikedIDs(ids, requesterID)
	helpfulMap, _ := s.repo.FindHelpfulIDs(ids, requesterID)

	responses := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		responses[i] = *toReviewResponse(&r, likedMap[r.ID], helpfulMap[r.ID], s.minio)
	}
	return responses, pg, nil
}

// GetMyMediaReview returns the caller's own review of a title regardless of visibility, or nil.
func (s *Service) GetMyMediaReview(userID uint, mediaID int, mediaType string) (*ReviewResponse, error) {
	review, err := s.repo.FindUserReviewByMedia(userID, mediaID, mediaType)
	if err != nil || review == nil {
		return nil, err
	}
	liked, _ := s.repo.IsLiked(review.ID, userID)
	helpful, _ := s.repo.IsHelpful(review.ID, userID)
	return toReviewResponse(review, liked, helpful, s.minio), nil
}

func (s *Service) UpdateReview(ctx context.Context, reviewID, requesterID uint, req UpdateReviewRequest) (*ReviewResponse, error) {
	if err := validateUpdateReviewRequest(req); err != nil {
		return nil, err
	}

	review, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return nil, err
	}
	if review.UserID != requesterID {
		return nil, ErrForbidden
	}

	updates, err := buildUpdateMap(req)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		liked, _ := s.repo.IsLiked(reviewID, requesterID)
		helpful, _ := s.repo.IsHelpful(reviewID, requesterID)
		return toReviewResponse(review, liked, helpful, s.minio), nil
	}

	if err := s.repo.UpdateReview(reviewID, updates); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return nil, err
	}

	if s.feedSvc != nil && req.Body != nil {
		_ = s.feedSvc.CreateActivity(ctx, requesterID, feed_module.ActivityReviewCreated, feed_module.ActivityPayload{
			MediaID:   &updated.MediaID,
			MediaType: &updated.MediaType,
			ReviewID:  &updated.ID,
			Message:   updated.Body,
		})
	}

	liked, _ := s.repo.IsLiked(reviewID, requesterID)
	helpful, _ := s.repo.IsHelpful(reviewID, requesterID)
	return toReviewResponse(updated, liked, helpful, s.minio), nil
}

func (s *Service) DeleteReview(ctx context.Context, reviewID, requesterID uint) error {
	review, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return err
	}
	if review.UserID != requesterID {
		return ErrForbidden
	}

	if err := s.repo.DeleteReview(reviewID); err != nil {
		return err
	}

	_ = s.expPort.AddExperience(requesterID, -stats.ExpPerReview)

	if s.feedSvc != nil {
		_ = s.feedSvc.DeleteReviewActivity(ctx, requesterID, reviewID)
	}

	return nil
}

var ratingSteps = [10]float32{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0}

func buildRatingResponse(mediaID int, mediaType string, rows []RatingBucket) *RatingResponse {
	dist := make([]RatingBucket, len(ratingSteps))
	for i, step := range ratingSteps {
		dist[i] = RatingBucket{Rating: step, Count: 0}
	}

	totalCount := 0
	weightedSum := 0.0
	for _, r := range rows {
		if r.Count <= 0 {
			continue
		}
		idx := int(math.Round(float64(r.Rating)*2)) - 1
		if idx >= 0 && idx < len(dist) {
			dist[idx].Count += r.Count
			totalCount += r.Count
			weightedSum += float64(dist[idx].Rating) * float64(r.Count)
		}
	}

	hasRating := totalCount > 0
	avg := float32(0)
	if hasRating {
		avg = float32(math.Round((weightedSum/float64(totalCount))*10) / 10)
	}

	return &RatingResponse{
		MediaID:       mediaID,
		MediaType:     mediaType,
		AverageRating: avg,
		ReviewCount:   totalCount,
		HasRating:     hasRating,
		Distribution:  dist,
	}
}

func (s *Service) GetMediaRating(mediaID int, mediaType string) (*RatingResponse, error) {
	if mediaType != "movie" && mediaType != "tv" {
		return nil, ErrInvalidMediaType
	}
	if mediaID <= 0 {
		return nil, ErrInvalidMediaID
	}

	rows, err := s.repo.GetMediaRatingBuckets(context.Background(), shared.MediaIdentity{
		ID:   mediaID,
		Type: shared.MediaType(mediaType),
	})
	if err != nil {
		return nil, err
	}

	return buildRatingResponse(mediaID, mediaType, rows), nil
}

func (s *Service) LikeReview(ctx context.Context, reviewID, requesterID uint) error {
	review, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return err
	}
	if review.Visibility == "private" && review.UserID != requesterID {
		return ErrForbidden
	}

	if err := s.repo.CreateLike(reviewID, requesterID); err != nil {
		return err
	}

	_ = s.expPort.AddExperience(review.UserID, stats.ExpPerLike)

	var likeGivenCount int64
	s.db.WithContext(ctx).Model(&ReviewLike{}).Where("user_id = ?", requesterID).Count(&likeGivenCount)
	unlocked := shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, requesterID, "review_like_given_count", int(likeGivenCount))
	s.createAchievementFeedActivities(ctx, requesterID, unlocked)

	var likeReceivedCount int64
	s.db.WithContext(ctx).Model(&ReviewLike{}).
		Joins("JOIN reviews ON reviews.id = review_likes.review_id").
		Where("reviews.user_id = ? AND reviews.deleted_at IS NULL", review.UserID).
		Count(&likeReceivedCount)
	unlocked = shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, review.UserID, "review_like_received_count", int(likeReceivedCount))
	s.createAchievementFeedActivities(ctx, review.UserID, unlocked)

	if s.notifSvc != nil {
		if actor, err := s.getUserSummary(requesterID); err == nil {
			title, _ := fetchMediaSummary(review.MediaID, review.MediaType)
			_ = s.notifSvc.PushReviewLiked(ctx, review.UserID, requesterID, reviewID, actor.Username, title)
			_ = s.notifSvc.PushFollowingLikedReview(ctx, requesterID, actor.Username, reviewID, title)
		}
	}

	if (review.Visibility == "public" || review.IsPublic) && s.feedSvc != nil {
		_ = s.feedSvc.CreateActivity(ctx, requesterID, feed_module.ActivityReviewLiked, feed_module.ActivityPayload{
			MediaID:   &review.MediaID,
			MediaType: &review.MediaType,
			ReviewID:  &reviewID,
		})
	}

	return nil
}

func (s *Service) UnlikeReview(ctx context.Context, reviewID, requesterID uint) error {
	review, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteLike(reviewID, requesterID); err != nil {
		return err
	}

	_ = s.expPort.AddExperience(review.UserID, -stats.ExpPerLike)

	if s.feedSvc != nil {
		_ = s.feedSvc.DeleteReviewActivity(ctx, requesterID, reviewID)
	}

	return nil
}

func (s *Service) MarkHelpful(ctx context.Context, reviewID, requesterID uint) error {
	review, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return err
	}
	if review.Visibility == "private" && review.UserID != requesterID {
		return ErrForbidden
	}
	if review.UserID == requesterID {
		return ErrForbidden
	}

	if err := s.repo.CreateHelpful(reviewID, requesterID); err != nil {
		return err
	}

	var helpfulReceivedCount int64
	s.db.WithContext(ctx).Model(&ReviewHelpful{}).
		Joins("JOIN reviews ON reviews.id = review_helpfuls.review_id").
		Where("reviews.user_id = ? AND reviews.deleted_at IS NULL", review.UserID).
		Count(&helpfulReceivedCount)
	unlocked := shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, review.UserID, "review_helpful_received_count", int(helpfulReceivedCount))
	s.createAchievementFeedActivities(ctx, review.UserID, unlocked)

	if s.notifSvc != nil {
		if actor, err := s.getUserSummary(requesterID); err == nil {
			title, _ := fetchMediaSummary(review.MediaID, review.MediaType)
			_ = s.notifSvc.PushReviewMarkedHelpful(ctx, review.UserID, requesterID, reviewID, actor.Username, title)
			_ = s.notifSvc.PushFollowingMarkedHelpful(ctx, requesterID, actor.Username, reviewID, title)
		}
	}

	return nil
}

func (s *Service) UnmarkHelpful(ctx context.Context, reviewID, requesterID uint) error {
	if _, err := s.repo.FindReviewByID(reviewID); err != nil {
		return err
	}
	return s.repo.DeleteHelpful(reviewID, requesterID)
}

func (s *Service) CreateComment(ctx context.Context, reviewID, requesterID uint, req CreateCommentRequest) (*CommentResponse, error) {
	review, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return nil, err
	}
	if review.Visibility == "private" && review.UserID != requesterID {
		return nil, ErrForbidden
	}

	comment := &ReviewComment{ReviewID: reviewID, UserID: requesterID, Body: req.Body}
	if err := s.repo.CreateComment(comment); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		if actor, err := s.getUserSummary(requesterID); err == nil {
			title, _ := fetchMediaSummary(review.MediaID, review.MediaType)
			_ = s.notifSvc.PushReviewCommented(ctx, review.UserID, requesterID, reviewID, actor.Username, title)
			_ = s.notifSvc.PushFollowingCommented(ctx, requesterID, actor.Username, reviewID, title)
		}
	}

	if (review.Visibility == "public" || review.IsPublic) && s.feedSvc != nil {
		_ = s.feedSvc.CreateActivity(ctx, requesterID, feed_module.ActivityReviewCommented, feed_module.ActivityPayload{
			MediaID:   &review.MediaID,
			MediaType: &review.MediaType,
			ReviewID:  &reviewID,
			CommentID: &comment.ID,
		})
	}

	return toCommentResponse(comment), nil
}

func (s *Service) GetComments(ctx context.Context, reviewID, requesterID uint) ([]CommentResponse, error) {
	review, err := s.repo.FindReviewByID(reviewID)
	if err != nil {
		return nil, err
	}
	if review.Visibility == "private" && review.UserID != requesterID {
		return nil, ErrForbidden
	}

	comments, err := s.repo.FindCommentsByReview(reviewID)
	if err != nil {
		return nil, err
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = *toCommentResponse(&c)
	}
	return responses, nil
}

func (s *Service) UpdateComment(ctx context.Context, commentID, requesterID uint, req UpdateCommentRequest) (*CommentResponse, error) {
	comment, err := s.repo.FindCommentByID(commentID)
	if err != nil {
		return nil, err
	}
	if comment.UserID != requesterID {
		return nil, ErrForbidden
	}
	if err := s.repo.UpdateComment(commentID, req.Body); err != nil {
		return nil, err
	}
	comment.Body = req.Body
	return toCommentResponse(comment), nil
}

func (s *Service) DeleteComment(ctx context.Context, commentID, reviewID, requesterID uint) error {
	comment, err := s.repo.FindCommentByID(commentID)
	if err != nil {
		return err
	}
	if comment.UserID != requesterID {
		return ErrForbidden
	}
	if err := s.repo.DeleteComment(commentID, reviewID); err != nil {
		return err
	}

	if s.feedSvc != nil {
		_ = s.feedSvc.DeleteCommentActivity(ctx, requesterID, commentID)
	}

	return nil
}

func (s *Service) createAchievementFeedActivities(ctx context.Context, userID uint, unlocked []shared.NeutralUnlocked) {
	if s.feedSvc == nil {
		return
	}
	for _, u := range unlocked {
		_ = s.feedSvc.CreateActivity(ctx, userID, feed_module.ActivityAchievementUnlocked, feed_module.ActivityPayload{
			AchievementID: &u.ID,
			Message:       u.Name,
		})
	}
}

func (s *Service) getUserSummary(userID uint) (*users.User, error) {
	var u users.User
	err := s.db.First(&u, userID).Error
	return &u, err
}

func validateReviewRequest(req CreateReviewRequest) error {
	if !shared.IsValidRating(req.Rating) {
		return ErrInvalidRating
	}
	if req.MediaType != "movie" && req.MediaType != "tv" {
		return ErrInvalidMediaType
	}
	if req.MediaID <= 0 {
		return ErrInvalidMediaID
	}
	return nil
}

func validateUpdateReviewRequest(req UpdateReviewRequest) error {
	if req.Rating != nil {
		if !shared.IsValidRating(*req.Rating) {
			return ErrInvalidRating
		}
	}
	return nil
}

func buildUpdateMap(req UpdateReviewRequest) (map[string]any, error) {
	updates := map[string]any{}
	if req.Rating != nil {
		updates["rating"] = *req.Rating
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.Visibility != nil || req.IsPublic != nil {
		visStr := ""
		if req.Visibility != nil {
			visStr = *req.Visibility
		}
		vis, err := resolveVisibility(visStr, req.IsPublic)
		if err != nil {
			return nil, err
		}
		updates["visibility"] = vis
		updates["is_public"] = (vis == "public")
	}
	return updates, nil
}

func toReviewResponse(r *Review, isLiked, isHelpful bool, minio *storage.MinIOClient) *ReviewResponse {
	if r == nil {
		return nil
	}

	title, posterURL := fetchMediaSummary(r.MediaID, r.MediaType)

	return &ReviewResponse{
		ID:   r.ID,
		User: toUserSummaryResponse(&r.User, minio),
		Media: tmdbmodule.Media{
			ID:        r.MediaID,
			MediaType: r.MediaType,
			Title:     title,
			PosterURL: posterURL,
		},
		Rating:         r.Rating,
		Body:           r.Body,
		Visibility:     r.Visibility,
		LikeCount:      r.LikeCount,
		CommentCount:   r.CommentCount,
		IsLiked:        isLiked,
		HelpfulCount:   r.HelpfulCount,
		IsHelpfulVoted: isHelpful,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func toUserSummaryResponse(u *users.User, minio *storage.MinIOClient) users.UserSummaryResponse {
	if u == nil || u.ID == 0 {
		return users.UserSummaryResponse{Username: "Unknown User"}
	}
	res := users.UserSummaryResponse{ID: u.ID, Username: u.Username}
	if u.DisplayName != nil && *u.DisplayName != "" {
		res.DisplayName = u.DisplayName
	}
	if u.AvatarURL != nil && *u.AvatarURL != "" {
		if minio != nil {
			if presigned, err := minio.PresignURL(context.Background(), *u.AvatarURL); err == nil {
				res.AvatarURL = &presigned
			} else {
				res.AvatarURL = u.AvatarURL
			}
		} else {
			res.AvatarURL = u.AvatarURL
		}
	}
	return res
}

func fetchMediaSummary(mediaID int, mediaType string) (title, posterURL string) {
	if mediaType == "movie" {
		movie, err := tmdbmodule.GetMovieByID(mediaID)
		if err != nil {
			return "", ""
		}
		return movie.Title, tmdbmodule.ImageURL(movie.PosterPath)
	}

	series, err := tmdbmodule.GetSeriesByID(mediaID)
	if err != nil {
		return "", ""
	}
	return series.Name, tmdbmodule.ImageURL(series.PosterPath)
}

func toCommentResponse(c *ReviewComment) *CommentResponse {
	return &CommentResponse{
		ID:        c.ID,
		ReviewID:  c.ReviewID,
		UserID:    c.UserID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// clampPage keeps page within [1, MaxReviewsPage] so (page-1)*pageSize cannot overflow.
func clampPage(page int) int {
	if page < 1 {
		return 1
	}
	if page > MaxReviewsPage {
		return MaxReviewsPage
	}
	return page
}
