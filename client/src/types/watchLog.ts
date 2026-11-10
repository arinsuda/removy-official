import type { Visibility } from "./common";
import type { MediaSummary } from "./movie";

export interface WatchLogResponse {
  id: number;
  media: MediaSummary;
  watched_on: string; // YYYY-MM-DD
  visibility: Visibility;
  created_at: string;
  updated_at: string;
}

export interface CreateWatchLogRequest {
  watched_on: string; // YYYY-MM-DD
  visibility: Visibility;
}

export interface UpdateWatchLogRequest {
  watched_on?: string;
  visibility?: Visibility;
}

export interface WatchSummaryResponse {
  watch_count: number;
  first_watched_on: string | null;
  last_watched_on: string | null;
  has_watched: boolean;
}

export interface WatchLogListResponse {
  items: WatchLogResponse[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}
