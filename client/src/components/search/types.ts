import type { MediaRatings } from "@/types/movie"

export interface SearchSuggestion {
  id: number
  type: "movie" | "tv"
  title: string
  poster_path: string | null
  /** Release (movie) or first air (series) year, empty when unknown. */
  year: string
  /** Legacy TMDB 0-10 score; used when `ratings.tmdb` is missing. */
  rating: number | null
  ratings?: MediaRatings
}

export type SearchStatus = "idle" | "loading" | "success" | "error"
