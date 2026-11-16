import type { Movie, TVSeries } from "@/types"
import type { MediaAward } from "@/types/event"
import type { MediaRatings } from "@/types/movie"
import { mediaRoute } from "@/utils/mediaRoute"
import { mediaImage } from "@/utils/image"
import type { MediaType } from "@/types/common"

/** Library, likes and feed entries can be manga too; only movie and tv ids exist on TMDB. */
export type HomeMediaType = MediaType

/** False for manga, whose ids are AniList ids: never send them to TMDB lookups. */
export const isTmdbType = (type: MediaType): boolean => type !== "manga"

/** Movies and TV series flattened to the fields the home page renders. */
export interface HomeMedia {
  /** Unique across types, e.g. "movie-27205". */
  key: string
  id: number
  type: HomeMediaType
  title: string
  poster_path: string
  backdrop_path: string
  /** Release or first-air date (YYYY-MM-DD) or "". */
  date: string
  overview: string
  vote_average: number
  ratings?: MediaRatings
  awards?: MediaAward[]
  source: Movie | TVSeries
}

export function fromMovie(m: Movie): HomeMedia {
  return {
    key: `movie-${m.id}`,
    id: m.id,
    type: "movie",
    title: m.title,
    poster_path: m.poster_path,
    backdrop_path: m.backdrop_path,
    date: m.release_date ?? "",
    overview: m.overview,
    vote_average: m.vote_average,
    ratings: m.ratings,
    awards: m.awards,
    source: m,
  }
}

export function fromSeries(s: TVSeries): HomeMedia {
  return {
    key: `tv-${s.id}`,
    id: s.id,
    type: "tv",
    title: s.name,
    poster_path: s.poster_path,
    backdrop_path: s.backdrop_path,
    date: s.first_air_date ?? "",
    overview: s.overview,
    vote_average: s.vote_average,
    ratings: s.ratings,
    awards: s.awards,
    source: s,
  }
}

export function detailRoute(m: { id: number; type: MediaType }) {
  return mediaRoute(m.type, m.id)
}

/** Absolute URLs (manga) yield the same string for both widths, so the srcset degrades to one candidate. */
export const posterSrc = (path: string) => mediaImage(path, "w342") ?? ""
export const posterSrcset = (path: string) => {
  const small = mediaImage(path, "w185") ?? ""
  const large = mediaImage(path, "w342") ?? ""
  return small === large ? large : `${small} 185w, ${large} 342w`
}
export const backdropSrc = (path: string) => mediaImage(path, "w1280") ?? ""
