import type { RouteLocationRaw } from "vue-router"
import { movieApi } from "@/api/api"
import { fromMovie, fromSeries, type HomeMedia } from "@/components/home/homeMedia"
import type {
  Cast,
  Credits,
  Genre,
  MediaRatings,
  MovieDetail,
  TVNetwork,
  TVSeason,
  TVSeriesDetail,
  Video,
} from "@/types/movie"

export type DetailType = "movie" | "tv"

/** A row of the facts list. Labels are translated by the renderer from `key`. */
export interface DetailFact {
  key: "originalTitle" | "status" | "releaseDate" | "firstAirDate" | "lastAirDate" | "budget" | "revenue" | "networks" | "seasons" | "episodes"
  value: string
  kind: "text" | "status" | "date"
}

export interface CrewHighlight {
  id: number
  name: string
  role: "director" | "creator"
  profilePath: string | null
}

/** Movie and TV payloads flattened into the one shape DetailPage renders. */
export interface DetailData {
  id: number
  type: DetailType
  title: string
  englishTitle: string | null
  tagline: string
  overview: string
  /** First four digits of the release or first-air date, or null. */
  year: string | null
  runtimeMin: number | null
  seasonCount: number | null
  episodeCount: number | null
  status: string | null
  genres: Genre[]
  ratings: MediaRatings | null
  voteAverage: number
  posterPath: string | null
  backdropPath: string | null
  homepage: string
  facts: DetailFact[]
  cast: Cast[]
  crewHighlights: CrewHighlight[]
  videos: Video[]
  tvSeasons: TVSeason[]
  networks: TVNetwork[]
}

export interface DetailConfig {
  type: DetailType
  fetch: (id: number) => Promise<DetailData>
  fetchSimilar: (id: number) => Promise<HomeMedia[]>
  /** Where "not found" and "back" lead. */
  fallbackRoute: RouteLocationRaw
}

const CAST_LIMIT = 12

/** i18n key suffix for a TMDB status, e.g. "Post Production" -> "PostProduction". */
export function statusKey(status: string): string {
  return status.replace(/[^A-Za-z0-9]/g, "")
}

function yearOf(date: string | null | undefined): string | null {
  const y = date?.split("-")[0]
  return y && /^\d{4}$/.test(y) ? y : null
}

function money(amount: number | undefined): string | null {
  return amount ? `$${amount.toLocaleString("en-US")}` : null
}

function pushFact(facts: DetailFact[], key: DetailFact["key"], kind: DetailFact["kind"], value: string | null | undefined) {
  if (value) facts.push({ key, kind, value })
}

function directors(credits: Credits | undefined): CrewHighlight[] {
  const seen = new Set<number>()
  return (credits?.crew ?? [])
    .filter(c => c.job === "Director" && !seen.has(c.id) && seen.add(c.id))
    .slice(0, 3)
    .map(c => ({ id: c.id, name: c.name, role: "director" as const, profilePath: c.profile_path || null }))
}

export interface MovieDetailResponse {
  movie: MovieDetail
  credits?: Credits
  videos?: Video[]
}

export interface TvDetailResponse {
  series: TVSeriesDetail
  credits?: Credits
  videos?: Video[]
}

export function adaptMovieDetail({ movie, credits, videos }: MovieDetailResponse): DetailData {
  const facts: DetailFact[] = []
  if (movie.original_title && movie.original_title !== movie.title) {
    pushFact(facts, "originalTitle", "text", movie.original_title)
  }
  pushFact(facts, "status", "status", movie.status)
  pushFact(facts, "releaseDate", "date", movie.release_date)
  pushFact(facts, "budget", "text", money(movie.budget))
  pushFact(facts, "revenue", "text", money(movie.revenue))

  return {
    id: movie.id,
    type: "movie",
    title: movie.title,
    englishTitle: movie.english_title || null,
    tagline: movie.tagline ?? "",
    overview: movie.overview ?? "",
    year: yearOf(movie.release_date),
    runtimeMin: movie.runtime || null,
    seasonCount: null,
    episodeCount: null,
    status: movie.status || null,
    genres: movie.genres ?? [],
    ratings: movie.ratings ?? null,
    voteAverage: movie.vote_average ?? 0,
    posterPath: movie.poster_path || null,
    backdropPath: movie.backdrop_path || null,
    homepage: movie.homepage ?? "",
    facts,
    cast: (credits?.cast ?? []).slice(0, CAST_LIMIT),
    crewHighlights: directors(credits),
    videos: videos ?? movie.videos ?? [],
    tvSeasons: [],
    networks: [],
  }
}

export function adaptTvDetail({ series, credits, videos }: TvDetailResponse): DetailData {
  const networks = series.networks ?? []
  const facts: DetailFact[] = []
  if (series.original_name && series.original_name !== series.name) {
    pushFact(facts, "originalTitle", "text", series.original_name)
  }
  pushFact(facts, "status", "status", series.status)
  pushFact(facts, "firstAirDate", "date", series.first_air_date)
  pushFact(facts, "lastAirDate", "date", series.last_air_date)
  pushFact(facts, "networks", "text", networks.map(n => n.name).join(", "))
  pushFact(facts, "seasons", "text", series.number_of_seasons ? String(series.number_of_seasons) : null)
  pushFact(facts, "episodes", "text", series.number_of_episodes ? String(series.number_of_episodes) : null)

  const creators: CrewHighlight[] = (series.created_by ?? []).slice(0, 3).map(c => ({
    id: c.id,
    name: c.name,
    role: "creator" as const,
    profilePath: c.profile_path || null,
  }))

  return {
    id: series.id,
    type: "tv",
    title: series.name,
    englishTitle: series.english_title || null,
    tagline: series.tagline ?? "",
    overview: series.overview ?? "",
    year: yearOf(series.first_air_date),
    runtimeMin: series.episode_run_time?.find(m => m > 0) ?? null,
    seasonCount: series.number_of_seasons || null,
    episodeCount: series.number_of_episodes || null,
    status: series.status || null,
    genres: series.genres ?? [],
    ratings: series.ratings ?? null,
    voteAverage: series.vote_average ?? 0,
    posterPath: series.poster_path || null,
    backdropPath: series.backdrop_path || null,
    homepage: series.homepage ?? "",
    facts,
    cast: (credits?.cast ?? []).slice(0, CAST_LIMIT),
    crewHighlights: creators.length ? creators : directors(credits),
    videos: videos ?? [],
    // Season 0 is TMDB's "Specials"; keep it last.
    tvSeasons: [...(series.seasons ?? [])].sort((a, b) => (a.season_number || 99) - (b.season_number || 99)),
    networks,
  }
}

export const movieDetail: DetailConfig = {
  type: "movie",
  fetch: async id => adaptMovieDetail((await movieApi.getById(id)).data),
  fetchSimilar: async id => ((await movieApi.getSimilar(id)).data.results ?? []).map(fromMovie),
  fallbackRoute: { name: "movies" },
}

export const tvDetail: DetailConfig = {
  type: "tv",
  fetch: async id => adaptTvDetail((await movieApi.getSeriesById(id)).data),
  fetchSimilar: async id => ((await movieApi.getSimilarSeries(id)).data.results ?? []).map(fromSeries),
  fallbackRoute: { name: "tv" },
}
