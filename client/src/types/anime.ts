/** Query params of GET /api/anime/tv and /api/anime/movies. The response is a PaginatedResult, like /tv/popular. */
export type AnimeMediaType = "movie" | "tv"
/** `airing` applies to TV only. */
export type AnimeSort = "popular" | "top_rated" | "airing" | "newest"
export type AnimeKeyword =
  | "isekai"
  | "mecha"
  | "shoujo"
  | "shounen"
  | "seinen"
  | "josei"
  | "slice_of_life"
  | "sports"
  | "idol"
  | "school"
export type AnimeCountry = "JP" | "CN" | "KR" | "all"
/** `season` applies to TV only and needs `year`. */
export type AnimeSeason = "winter" | "spring" | "summer" | "fall"

/** At most one of decade | year | season+year. The page number travels separately. */
export interface AnimeQuery {
  sort?: AnimeSort
  /** TMDB genre id; the server rejects 16 (Animation), which the client never sends. */
  genre?: number
  keyword?: AnimeKeyword
  country?: AnimeCountry
  /** Decade start year, e.g. 1990. */
  decade?: number
  year?: number
  season?: AnimeSeason
}
