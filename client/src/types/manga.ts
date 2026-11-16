import type { RatingSource } from "./movie"

/** Origin country; "all" removes the filter. JP is manga, KR manhwa, CN manhua. */
export type MangaCountry = "JP" | "KR" | "CN" | "all"
export type MangaFormat = "manga" | "one_shot" | "novel" | "all"
export type MangaSort = "popular" | "top_rated" | "trending" | "newest"
export type MangaStatus = "releasing" | "finished" | "all"
export type MangaKind = "manga" | "manhwa" | "manhua" | "novel"

/** Query params of GET /api/manga. The page number travels separately. */
export interface MangaQuery {
  country?: MangaCountry
  format?: MangaFormat
  sort?: MangaSort
  /** Genre name as the API lists it (GET /api/manga/genres). */
  genre?: string
  status?: MangaStatus
  /** Only honored for signed-in viewers who are 18+; the server ignores it otherwise. */
  include_adult?: boolean
}

export interface MangaTitles {
  romaji: string
  english: string
  native: string
  thai: string
}

export interface MangaRatings {
  popcorn?: RatingSource
  anilist?: AniListRating
}

/** AniList average score; `average` is meaningful only when `available` is true. */
export interface AniListRating {
  average: number
  available: boolean
  /** Upper bound of the score, 100 for AniList. */
  scale: number
}

/** AniList average score, or null when AniList has none. */
export function anilistScore(ratings: MangaRatings): number | null {
  const r = ratings.anilist
  return r?.available ? r.average : null
}

/** One title in a list, search result or similar rail. */
export interface MangaItem {
  id: number
  media_type: "manga"
  titles: MangaTitles
  display_title: string
  /** Absolute AniList URL. Never run it through the TMDB image helpers. */
  cover_url: string
  banner_url: string
  format: Exclude<MangaFormat, "all">
  country: Exclude<MangaCountry, "all">
  kind: MangaKind
  /** Adult titles show an 18+ badge and their covers stay blurred until the viewer confirms. */
  is_adult: boolean
  status: string
  start_year: number | null
  chapters: number | null
  volumes: number | null
  genres: string[]
  ratings: MangaRatings
}

export interface MangaReadLink {
  site: string
  url: string
  language: string
}

export interface MangaRelation {
  id: number
  title: string
}

/** GET /api/manga/:id returns this under the `manga` key. */
export interface MangaDetail extends MangaItem {
  description: string
  tags: { name: string }[]
  read_links: MangaReadLink[]
  relations: MangaRelation[]
}
