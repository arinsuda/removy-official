import type { HomeMedia } from "@/components/home/homeMedia"
import type { Movie, TVSeries } from "@/types"
import type { MediaSummary } from "@/types/movie"

/** What the library and likes endpoints return for a title. */
export type SummaryLike = Pick<MediaSummary, "id" | "title" | "media_type" | "poster_url" | "vote_average" | "ratings" | "release_date">

/** A saved title (list entry or like) in the shape the tab kit renders. */
export interface ProfileEntry {
  /** Library item id (used to remove it); likes use the media id. */
  id: number
  media: HomeMedia
  /** ISO timestamp the entry was added (or watched). */
  at: string
}

/**
 * Library summaries carry a TMDB poster path, not a full title record, so the
 * popup source is a minimal stand-in (no overview or backdrop).
 */
export function summaryToHomeMedia(m: SummaryLike): HomeMedia {
  const tv = m.media_type === "tv"
  const base = { id: m.id, poster_path: m.poster_url ?? "", backdrop_path: "", overview: "", vote_average: m.vote_average ?? 0, vote_count: 0, popularity: 0, genre_ids: [], adult: false, ratings: m.ratings }
  const source: Movie | TVSeries = tv
    ? { ...base, name: m.title, original_name: m.title, first_air_date: m.release_date ?? "" }
    : { ...base, title: m.title, original_title: m.title, release_date: m.release_date ?? "" }
  return {
    key: `${tv ? "tv" : "movie"}-${m.id}`,
    id: m.id,
    type: tv ? "tv" : "movie",
    title: m.title,
    poster_path: base.poster_path,
    backdrop_path: "",
    date: m.release_date ?? "",
    overview: "",
    vote_average: base.vote_average,
    ratings: m.ratings,
    source,
  }
}

const popcornAvg = (e: ProfileEntry) => e.media.ratings?.popcorn?.average ?? 0
const time = (e: ProfileEntry) => new Date(e.at).getTime() || 0

export const ENTRY_SORT_KEYS = ["newest", "oldest", "release", "title", "tmdb", "popcorn"] as const

/** Release date as a sortable day number; titles without a date sort last. */
const releaseTime = (e: ProfileEntry) => Date.parse(e.media.date) || Number.NEGATIVE_INFINITY

/** Shared sorts for the poster-grid tabs. */
export const ENTRY_SORTS: Record<(typeof ENTRY_SORT_KEYS)[number], (a: ProfileEntry, b: ProfileEntry) => number> = {
  newest: (a, b) => time(b) - time(a),
  oldest: (a, b) => time(a) - time(b),
  release: (a, b) => (releaseTime(a) === releaseTime(b) ? 0 : releaseTime(a) === Number.NEGATIVE_INFINITY ? 1 : releaseTime(b) === Number.NEGATIVE_INFINITY ? -1 : releaseTime(b) - releaseTime(a)),
  title: (a, b) => a.media.title.localeCompare(b.media.title),
  tmdb: (a, b) => b.media.vote_average - a.media.vote_average,
  popcorn: (a, b) => popcornAvg(b) - popcornAvg(a),
}

export const ENTRY_FILTERS: Record<string, (e: ProfileEntry) => boolean> = {
  movie: (e) => e.media.type === "movie",
  tv: (e) => e.media.type === "tv",
}

