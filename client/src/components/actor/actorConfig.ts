/** Credits shown per "show more" step in the poster grid. */
export const GRID_PAGE_SIZE = 24

/** Titles in the "known for" rail. */
export const KNOWN_FOR_COUNT = 10

/** Biographies longer than this start clamped, with a read-more toggle. */
export const BIO_CLAMP_CHARS = 320

export const ACTOR_PROFILE_BASE = "https://image.tmdb.org/t/p/w500"
export const IMDB_NAME_URL = "https://www.imdb.com/name/"

export type CreditRole = "acting" | "crew"
export type CreditSort = "popular" | "newest" | "oldest"
export type CreditFilter = "all" | "movie" | "tv"
export type CreditView = "grid" | "timeline"

export const CREDIT_ROLES: CreditRole[] = ["acting", "crew"]
export const CREDIT_SORTS: CreditSort[] = ["popular", "newest", "oldest"]
export const CREDIT_FILTERS: CreditFilter[] = ["all", "movie", "tv"]
export const CREDIT_VIEWS: CreditView[] = ["grid", "timeline"]

/** TMDB departments that have a translated label under actor.departments.* */
export const KNOWN_DEPARTMENTS = [
  "Acting",
  "Directing",
  "Writing",
  "Production",
  "Sound",
  "Camera",
  "Editing",
  "Art",
  "Crew",
  "Costume & Make-Up",
  "Visual Effects",
  "Lighting",
  "Creator",
]
