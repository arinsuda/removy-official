import { MAX_PAGES } from "@/components/browse/browseConfig"
import type { AnimeCountry, AnimeKeyword, AnimeMediaType, AnimeSeason, AnimeSort } from "@/types/anime"

/** TMDB Animation genre. Anime is already Animation, so it is never offered as a genre filter. */
export const ANIME_EXCLUDED_GENRE = 16
export const ANIME_TYPES: AnimeMediaType[] = ["tv", "movie"]
export const ANIME_DEFAULT_TYPE: AnimeMediaType = "tv"
export const ANIME_DEFAULT_SORT: AnimeSort = "popular"
export const ANIME_DEFAULT_COUNTRY: AnimeCountry = "JP"
export const ANIME_SORTS: Record<AnimeMediaType, AnimeSort[]> = {
  tv: ["popular", "top_rated", "airing", "newest"],
  movie: ["popular", "top_rated", "newest"],
}
export const ANIME_KEYWORDS: AnimeKeyword[] = [
  "isekai",
  "mecha",
  "shoujo",
  "shounen",
  "seinen",
  "josei",
  "slice_of_life",
  "sports",
  "idol",
  "school",
]
export const ANIME_COUNTRIES: AnimeCountry[] = ["JP", "CN", "KR", "all"]
export const ANIME_SEASONS: AnimeSeason[] = ["winter", "spring", "summer", "fall"]
/** Earliest decade and year the release filter offers. */
export const ANIME_MIN_DECADE = 1960
const YEARS_PER_DECADE = 10
/** Browse pages are capped the same way as the other browse lists. */
export const ANIME_PAGE_CAP = MAX_PAGES
/** Titles shown in the home anime rail. */
export const ANIME_RAIL_SIZE = 12
export const ANIME_LIST_STALE_MS = 60_000
export const ANIME_RAIL_STALE_MS = 5 * 60_000

/** Decade start years from the current decade down to ANIME_MIN_DECADE, newest first. */
export function animeDecades(currentYear: number): number[] {
  const decades: number[] = []
  for (let d = Math.floor(currentYear / YEARS_PER_DECADE) * YEARS_PER_DECADE; d >= ANIME_MIN_DECADE; d -= YEARS_PER_DECADE) {
    decades.push(d)
  }
  return decades
}

/** Years from the current year down to ANIME_MIN_DECADE, newest first. */
export function animeYears(currentYear: number): number[] {
  const years: number[] = []
  for (let y = currentYear; y >= ANIME_MIN_DECADE; y--) years.push(y)
  return years
}
