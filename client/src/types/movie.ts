import type { MediaType, ListType } from "./common"

export interface Genre {
  id: number
  name: string
}

export interface RatingSource {
  average: number | null
  count: number
  available: boolean
  scale: number
}

/** POPCORN (0-5, from reviews) is primary; the others are secondary. imdb and rotten_tomatoes exist on detail only. */
export interface MediaRatings {
  popcorn: RatingSource
  tmdb: RatingSource
  imdb?: RatingSource
  rotten_tomatoes?: RatingSource
}

export interface Movie {
  id: number
  title: string
  original_title: string
  overview: string
  poster_path: string
  backdrop_path: string
  release_date: string
  vote_average: number
  vote_count: number
  popularity: number
  genre_ids: number[]
  adult: boolean
  ratings?: MediaRatings
}

export interface MovieDetail extends Movie {
  english_title?: string
  runtime: number
  status: string
  tagline: string
  budget: number
  revenue: number
  homepage: string
  genres: Genre[]
  videos?: Video[]
}

export interface TVSeries {
  id: number
  name: string
  original_name: string
  overview: string
  poster_path: string
  backdrop_path: string
  first_air_date: string
  vote_average: number
  vote_count: number
  popularity: number
  genre_ids: number[]
  adult: boolean
  ratings?: MediaRatings
}

export interface TVSeriesDetail extends TVSeries {
  english_title?: string
  number_of_seasons: number
  number_of_episodes: number
  status: string
  tagline: string
  homepage: string
  genres: Genre[]
}

export interface Cast {
  id: number
  name: string
  character: string
  profile_path: string
  order: number
}

export interface Crew {
  id: number
  name: string
  job: string
  department: string
  profile_path: string
}

export interface Credits {
  cast: Cast[]
  crew: Crew[]
}

export interface Video {
  id: string
  key: string
  name: string
  site: string
  type: string
  official: boolean
}

export interface MediaSummary {
  id: number
  title: string
  english_title?: string
  original_title?: string
  original_name?: string
  name?: string
  poster_url: string
  media_type: MediaType
  genres: Genre[] 
  vote_average: number
  ratings?: MediaRatings
  created_at: string
}

export interface MediaStats {
  media_id: number
  media_type: string
  like_count: number
  view_count: number
  review_count: number
  watchlist_count: number
  average_rating: number
  has_rating: boolean
  liked_at: string | null
  watchlisted_at: string | null
}

export interface MediaStatsResponse {
  stats: MediaStats
}

export interface AddItemRequest {
  media_id: number
  media_type: MediaType
  list_type: ListType
  watched_at?: string
  note?: string
  tags?: string[]
}

export interface LibraryItemResponse {
  id: number
  media: MediaSummary
  list_type: ListType
  watched_at: string | null
  tags: string[]
  note: string | null
  created_at: string
}

export interface MediaItemStatus {
  list_type: ListType
  item_id: number
}

export interface MediaStatusResponse {
  media_id: number
  media_type: MediaType
  in_lists: MediaItemStatus[]
}

export interface BMOLItemResponse {
  id: number
  media: MediaSummary
  media_type: MediaType
  rank: number
  created_at: string
}

export interface PersonDetail {
  id: number
  name: string
  biography: string
  birthday: string | null
  deathday: string | null
  place_of_birth: string | null
  profile_path: string | null
  known_for_department: string
  popularity: number
  gender: number
  imdb_id: string | null
  homepage: string | null
}

export interface ActorMovieCredit extends Movie {
  character?: string
  credit_id?: string
}

export interface PersonMovieCredits {
  cast: ActorMovieCredit[]
  crew: Movie[]
}

export interface ActorTVCredit extends TVSeries {
  character?: string
  credit_id?: string
  episode_count?: number
}

export interface PersonTVCredits {
  cast: ActorTVCredit[]
  crew: TVSeries[]
}

