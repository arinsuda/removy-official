import type { MediaRatings } from "@/types/movie";
import type { UserSummary } from "./user";
import type { MediaSummary } from "./movie";
import type { MediaType, ListType, Visibility } from "./common";

export interface ReviewResponse {
  id: number;
  user: UserSummary;
  media: MediaSummary;
  rating: number;
  body: string;
  visibility: Visibility;
  is_public?: boolean;
  like_count: number;
  comment_count: number;
  is_liked: boolean;
  helpful_count: number;
  is_helpful_voted: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateReviewRequest {
  media_id: number;
  media_type: MediaType;
  rating: number;
  body: string;
  visibility: Visibility;
}

export interface UpdateReviewRequest {
  rating?: number;
  body?: string;
  visibility?: Visibility;
}

export interface CommentResponse {
  id: number;
  review_id: number;
  user: UserSummary;
  body: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCommentRequest {
  body: string;
}

export interface LikeItem {
  id: number;
  created_at: string;
  media: {
    id: number;
    title: string;
    media_type: "movie" | "tv";
    poster_url: string;
    vote_average: number;
    ratings?: MediaRatings;
    genres: { id: number; name: string }[];
  };
}

export interface LikesResponse {
  likes: LikeItem[];
}
