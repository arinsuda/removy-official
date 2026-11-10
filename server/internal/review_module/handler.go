package review_module

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/internal/shared"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ── Review ────────────────────────────────────────────────────────

func (h *Handler) CreateReview(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req CreateReviewRequest
	if err := c.Bind().JSON(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	review, err := h.svc.CreateReview(c.Context(), claims.UserID, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"review": review})
}

// GetUserReviews handles GET /users/:userId/reviews
//
// Query params รองรับ:
//   - visibility: "all" (default) | "public" | "private"
//   - date_from, date_to: "YYYY-MM-DD" — กรองตามวันที่เขียนรีวิว (created_at)
//   - year, month: กรองแบบรายเดือน/รายปี (ลำดับความสำคัญสูงกว่า date_from/date_to)
//     ถ้าส่งแค่ year → กรองทั้งปี, ถ้าส่ง year+month → กรองเฉพาะเดือนนั้น
func (h *Handler) GetUserReviews(c fiber.Ctx) error {
	userID, err := parseUserID(c)
	if err != nil {
		return badRequest(c, "invalid user id")
	}
	claims := mw.GetClaims(c)

	filter, err := parseReviewFilter(c)
	if err != nil {
		return badRequest(c, err.Error())
	}

	reviews, pagination, summary, err := h.svc.GetUserReviews(c.Context(), userID, claims.UserID, filter)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"reviews": reviews, "pagination": pagination, "summary": summary})
}

func (h *Handler) GetMediaReviews(c fiber.Ctx) error {
	mediaID, err := strconv.Atoi(c.Params("mediaId"))
	if err != nil || mediaID <= 0 {
		return badRequest(c, "invalid media id")
	}

	mt := routeToMediaType(c.Params("mediaType"))
	if mt == "" {
		return badRequest(c, "invalid media type")
	}

	// Public route: guests have no claims and read as viewer 0.
	var viewerID uint
	if claims := mw.GetClaims(c); claims != nil {
		viewerID = claims.UserID
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	reviews, pagination, err := h.svc.GetMediaReviews(c.Context(), mediaID, mt, viewerID, page, pageSize)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"reviews": reviews, "pagination": pagination})
}

// GetMyMediaReview handles GET /reviews/media/:mediaType/:mediaId/me: the caller's own review
// of the title in any visibility, or {"review": null}.
func (h *Handler) GetMyMediaReview(c fiber.Ctx) error {
	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	mediaID, err := strconv.Atoi(c.Params("mediaId"))
	if err != nil || mediaID <= 0 {
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidInput, "invalid media id")
	}
	mt := routeToMediaType(c.Params("mediaType"))
	if mt == "" {
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidMediaType, "invalid media type")
	}
	review, err := h.svc.GetMyMediaReview(claims.UserID, mediaID, mt)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"review": review})
}

func (h *Handler) UpdateReview(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}

	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req UpdateReviewRequest
	if err := c.Bind().JSON(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	review, err := h.svc.UpdateReview(c.Context(), reviewID, claims.UserID, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"review": review})
}

func (h *Handler) DeleteReview(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}

	claims := mw.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := h.svc.DeleteReview(c.Context(), reviewID, claims.UserID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── In-app Rating ─────────────────────────────────────────────────

// GetMediaRating handles:
//
//	GET /movies/:mediaId/rating
//	GET /tv/:mediaId/rating
//
// ไม่ต้อง auth เพราะเป็นข้อมูล public aggregate
func (h *Handler) GetMediaRating(c fiber.Ctx) error {
	mediaID, err := strconv.Atoi(c.Params("mediaId"))
	if err != nil || mediaID <= 0 {
		return badRequest(c, "invalid media id")
	}

	mt := routeToMediaType(c.Params("mediaType"))
	if mt == "" {
		return badRequest(c, "invalid media type")
	}

	rating, err := h.svc.GetMediaRating(mediaID, mt)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"rating": rating})
}

// ── Like ──────────────────────────────────────────────────────────

func (h *Handler) LikeReview(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.LikeReview(c.Context(), reviewID, claims.UserID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) UnlikeReview(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.UnlikeReview(c.Context(), reviewID, claims.UserID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Comment ───────────────────────────────────────────────────────

func (h *Handler) CreateComment(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}
	claims := mw.GetClaims(c)

	var req CreateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	comment, err := h.svc.CreateComment(c.Context(), reviewID, claims.UserID, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"comment": comment})
}

func (h *Handler) GetComments(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}
	claims := mw.GetClaims(c)

	comments, err := h.svc.GetComments(c.Context(), reviewID, claims.UserID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"comments": comments})
}

func (h *Handler) UpdateComment(c fiber.Ctx) error {
	commentID, err := parseCommentID(c)
	if err != nil {
		return badRequest(c, "invalid comment id")
	}
	claims := mw.GetClaims(c)

	var req UpdateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	comment, err := h.svc.UpdateComment(c.Context(), commentID, claims.UserID, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(fiber.Map{"comment": comment})
}

func (h *Handler) DeleteComment(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}
	commentID, err := parseCommentID(c)
	if err != nil {
		return badRequest(c, "invalid comment id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.DeleteComment(c.Context(), commentID, reviewID, claims.UserID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Helpful ───────────────────────────────────────────────────────

func (h *Handler) MarkHelpful(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.MarkHelpful(c.Context(), reviewID, claims.UserID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) UnmarkHelpful(c fiber.Ctx) error {
	reviewID, err := parseReviewID(c)
	if err != nil {
		return badRequest(c, "invalid review id")
	}
	claims := mw.GetClaims(c)

	if err := h.svc.UnmarkHelpful(c.Context(), reviewID, claims.UserID); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ── Helpers ───────────────────────────────────────────────────────

// parseReviewFilter อ่าน query params ของ GET /users/:userId/reviews
// แล้วแปลงเป็น ReviewFilter สำหรับใช้ query ใน repository
func parseReviewFilter(c fiber.Ctx) (ReviewFilter, error) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	filter := ReviewFilter{
		Visibility:   "all",
		Sort:         ReviewSortNewest,
		RatingFilter: ReviewRatingFilterAll,
		Page:         page,
		PageSize:     pageSize,
	}

	if v := c.Query("visibility"); v != "" {
		switch v {
		case "all", "public", "private":
			filter.Visibility = v
		default:
			return filter, errors.New("visibility must be 'all', 'public', or 'private'")
		}
	}

	if s := c.Query("sort"); s != "" {
		switch s {
		case ReviewSortNewest, ReviewSortOldest, ReviewSortHighest, ReviewSortLowest:
			filter.Sort = s
		default:
			return filter, errors.New("invalid sort")
		}
	}

	if rf := c.Query("filter"); rf != "" {
		switch rf {
		case ReviewRatingFilterAll, ReviewRatingFilterHigh, ReviewRatingFilterLow, ReviewRatingFilterWithText:
			filter.RatingFilter = rf
		default:
			return filter, errors.New("invalid filter")
		}
	}

	search := strings.TrimSpace(c.Query("search"))
	if runes := []rune(search); len(runes) > MaxReviewSearchLength {
		search = string(runes[:MaxReviewSearchLength])
	}
	filter.Search = search

	// year (+month ถ้ามี) มีความสำคัญกว่า date_from/date_to แบบ manual
	if yearStr := c.Query("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1 {
			return filter, errors.New("invalid year")
		}

		month := 0
		if monthStr := c.Query("month"); monthStr != "" {
			month, err = strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				return filter, errors.New("invalid month, must be 1-12")
			}
		}

		if month > 0 {
			from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
			to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
			filter.DateFrom = &from
			filter.DateTo = &to
		} else {
			from := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
			to := from.AddDate(1, 0, 0).Add(-time.Nanosecond)
			filter.DateFrom = &from
			filter.DateTo = &to
		}
		return filter, nil
	}

	if v := c.Query("date_from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, errors.New("date_from must be YYYY-MM-DD")
		}
		filter.DateFrom = &t
	}
	if v := c.Query("date_to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, errors.New("date_to must be YYYY-MM-DD")
		}
		// รวมทั้งวันนั้นด้วย → set เวลาเป็นท้ายวัน
		end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
		filter.DateTo = &end
	}

	return filter, nil
}

// routeToMediaType แปลง URL segment → media_type value ใน DB
//
//	"movies" → "movie"
//	"series" → "tv"
func routeToMediaType(route string) string {
	switch route {
	case "movies", "movie":
		return "movie"
	case "series", "tv":
		return "tv"
	default:
		return ""
	}
}

func parseUserID(c fiber.Ctx) (uint, error) {
	id, err := strconv.Atoi(c.Params("userId"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid user id")
	}
	return uint(id), nil
}

func parseReviewID(c fiber.Ctx) (uint, error) {
	id, err := strconv.Atoi(c.Params("reviewId"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid review id")
	}
	return uint(id), nil
}

func parseCommentID(c fiber.Ctx) (uint, error) {
	id, err := strconv.Atoi(c.Params("commentId"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid comment id")
	}
	return uint(id), nil
}

func handleError(c fiber.Ctx, err error) error {
	switch {
	case shared.IsPgUniqueViolationOn(err, "uq_active_user_media_review"):
		return shared.WriteAPIError(c, fiber.StatusConflict, shared.ErrorCodeDuplicateReview, "you have already reviewed this media")
	case errors.Is(err, ErrReviewNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "review not found"})
	case errors.Is(err, ErrCommentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "comment not found"})
	case errors.Is(err, ErrForbidden):
		return forbidden(c)
	case errors.Is(err, ErrAlreadyLiked):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already liked"})
	case errors.Is(err, ErrNotLiked):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "not liked"})
	case errors.Is(err, ErrAlreadyMarkedHelpful):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already marked helpful"})
	case errors.Is(err, ErrNotMarkedHelpful):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "not marked helpful"})
	case errors.Is(err, ErrInvalidWatchedAt):
		return badRequest(c, "invalid watched_at format, use YYYY-MM-DD")
	case errors.Is(err, ErrInvalidRating):
		return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidRating, "rating must be between 0.5 and 5.0 with step 0.5")
	case errors.Is(err, ErrInvalidMediaType):
		return badRequest(c, "media_type must be 'movie' or 'tv'")
	case errors.Is(err, ErrInvalidMediaID):
		return badRequest(c, "invalid media_id")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}

func badRequest(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func forbidden(c fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
}
