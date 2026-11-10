import api from "../index"
import type {
  ReviewResponse,
  CreateReviewRequest,
  UpdateReviewRequest,
  CommentResponse,
  CreateCommentRequest,
  PaginatedResponse,
  FollowStatsResponse,
  FollowRelationshipStatus,
  FollowActionResponse,
  FollowUserSummary,
  FeedItemResponse,
} from "@/types"

export type ReviewVisibilityFilter = "all" | "public" | "private"

export interface GetUserReviewsParams {
  page?: number
  limit?: number
  visibility?: ReviewVisibilityFilter
  date_from?: string
  date_to?: string
  year?: number
  month?: number
}

export const reviewApi = {
  createReview: (data: CreateReviewRequest) =>
    api.post<{ review: ReviewResponse }>("/reviews", data),

  getUserReviews: (userId: number, params: GetUserReviewsParams = {}) =>
    api.get<{ reviews: ReviewResponse[] }>(`/reviews/user/${userId}`, {
      params,
    }),

  getMediaReviews: (
    mediaType: "movie" | "tv" | "movies",
    mediaId: number,
    params?: { page?: number; page_size?: number },
  ) => {
    // Convert movies -> movie, tv remains tv/series
    const mt = mediaType === "movies" ? "movie" : mediaType === "tv" ? "tv" : mediaType;
    return api.get<{ reviews: ReviewResponse[]; pagination: { page: number; page_size: number; total: number; has_more: boolean } }>(
      `/reviews/media/${mt}/${mediaId}`,
      { params },
    );
  },

  /** The signed-in user's own review of a title, or null. GET /reviews/media/:type/:id/me */
  getMyReview: (mediaType: "movie" | "tv", mediaId: number) =>
    api.get<{ review: ReviewResponse | null }>(`/reviews/media/${mediaType}/${mediaId}/me`),

  updateReview: (reviewId: number, data: UpdateReviewRequest) =>
    api.patch<{ review: ReviewResponse }>(
      `/reviews/${reviewId}`,
      data,
    ),

  deleteReview: (reviewId: number) =>
    api.delete(`/reviews/${reviewId}`),

  likeReview: (reviewId: number) =>
    api.post(`/reviews/${reviewId}/like`),

  unlikeReview: (reviewId: number) =>
    api.delete(`/reviews/${reviewId}/like`),

  markHelpful: (reviewId: number) =>
    api.post<{ helpful_count: number }>(`/reviews/${reviewId}/helpful`),

  unmarkHelpful: (reviewId: number) =>
    api.delete<{ helpful_count: number }>(`/reviews/${reviewId}/helpful`),

  getComments: (reviewId: number) =>
    api.get<{ comments: CommentResponse[] }>(
      `/reviews/${reviewId}/comments`,
    ),

  createComment: (reviewId: number, data: CreateCommentRequest) =>
    api.post<{ comment: CommentResponse }>(
      `/reviews/${reviewId}/comments`,
      data,
    ),

  updateComment: (commentId: number, body: string) =>
    api.patch<{ comment: CommentResponse }>(
      `/reviews/comments/${commentId}`,
      { body },
    ),

  deleteComment: (reviewId: number, commentId: number) =>
    api.delete(`/reviews/${reviewId}/comments/${commentId}`),
}



export const followApi = {
  follow: (userId: number) =>
    api.post<FollowActionResponse>(`/users/${userId}/follow`),

  unfollow: (userId: number) => api.delete(`/users/${userId}/follow`),

  getFollowStats: (userId: number) =>
    api.get<FollowStatsResponse>(`/users/${userId}/follow-stats`),

  getFollowStatus: (userId: number) =>
    api.get<FollowRelationshipStatus>(`/users/${userId}/follow-status`),

  getFollowers: (userId: number) =>
    api.get<{ followers: FollowUserSummary[] }>(`/users/${userId}/followers`),

  getFollowing: (userId: number) =>
    api.get<{ following: FollowUserSummary[] }>(`/users/${userId}/following`),

  getPendingRequests: (userId: number) =>
    api.get<{ requests: FollowUserSummary[] }>(
      `/users/${userId}/follow-requests`,
    ),

  acceptRequest: (userId: number, followerId: number) =>
    api.post(`/users/${userId}/follow-requests/${followerId}/accept`),

  rejectRequest: (userId: number, followerId: number) =>
    api.delete(`/users/${userId}/follow-requests/${followerId}`),

  getFeed: (page = 1, limit = 20) =>
    api.get<PaginatedResponse<FeedItemResponse>>("/feed", {
      params: { page, limit },
    }),
}
