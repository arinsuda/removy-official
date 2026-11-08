import type { RouteLocationRaw } from "vue-router"
import type { Movie, TVSeries } from "@/types"
import type { MediaRatings } from "@/types/movie"

export type HomeMediaType = "movie" | "tv"

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
    source: s,
  }
}

export function detailRoute(m: Pick<HomeMedia, "id" | "type">): RouteLocationRaw {
  return { name: m.type === "tv" ? "tv-detail" : "movie-detail", params: { id: m.id } }
}

const TMDB = "https://image.tmdb.org/t/p"

export const posterSrc = (path: string) => `${TMDB}/w342${path}`
export const posterSrcset = (path: string) => `${TMDB}/w185${path} 185w, ${TMDB}/w342${path} 342w`
export const backdropSrc = (path: string) => `${TMDB}/w1280${path}`
