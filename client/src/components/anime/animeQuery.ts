import type { LocationQuery, LocationQueryRaw } from "vue-router"
import type { AnimeCountry, AnimeKeyword, AnimeMediaType, AnimeQuery, AnimeSeason, AnimeSort } from "@/types/anime"
import {
  ANIME_COUNTRIES,
  ANIME_DEFAULT_COUNTRY,
  ANIME_DEFAULT_SORT,
  ANIME_DEFAULT_TYPE,
  ANIME_EXCLUDED_GENRE,
  ANIME_KEYWORDS,
  ANIME_SEASONS,
  ANIME_SORTS,
  ANIME_TYPES,
  animeDecades,
  animeYears,
} from "./animeConfig"

export const ANIME_QUERY_KEYS = ["type", "sort", "genre", "keyword", "country", "decade", "year", "season"] as const

export interface AnimeBrowseState {
  type: AnimeMediaType
  sort: AnimeSort
  genre: number | null
  keyword: AnimeKeyword | null
  country: AnimeCountry
  decade: number | null
  year: number | null
  season: AnimeSeason | null
}

export const DEFAULT_ANIME_STATE: AnimeBrowseState = {
  type: ANIME_DEFAULT_TYPE,
  sort: ANIME_DEFAULT_SORT,
  genre: null,
  keyword: null,
  country: ANIME_DEFAULT_COUNTRY,
  decade: null,
  year: null,
  season: null,
}

const first = (v: LocationQuery[string] | undefined): string => (Array.isArray(v) ? (v[0] ?? "") : (v ?? ""))

/** The allowed value, or null when the raw text is not one of them. */
function pick<T extends string>(raw: string, allowed: readonly T[]): T | null {
  return (allowed as readonly string[]).includes(raw) ? (raw as T) : null
}

/** Whole number from URL text, or null. */
function toInt(raw: string): number | null {
  if (raw.trim() === "") return null
  const n = Number(raw)
  return Number.isInteger(n) ? n : null
}

/**
 * Enforces the API rules: season needs a year and is TV only, decade and year never combine
 * (year wins), the sort must exist for the type and genre 16 is never a filter.
 */
export function normalizeAnimeState(s: AnimeBrowseState): AnimeBrowseState {
  const hasYear = s.year !== null
  return {
    ...s,
    sort: ANIME_SORTS[s.type].includes(s.sort) ? s.sort : DEFAULT_ANIME_STATE.sort,
    genre: s.genre === ANIME_EXCLUDED_GENRE ? null : s.genre,
    decade: hasYear ? null : s.decade,
    season: s.type === "tv" && hasYear ? s.season : null,
  }
}

/**
 * Reads ?type&sort&genre&keyword&country&decade&year&season. Unknown or out-of-range values
 * fall back to the default. `currentYear` bounds the decade and year lists.
 */
export function parseAnimeQuery(query: LocationQuery, currentYear: number): AnimeBrowseState {
  const type = pick(first(query.type), ANIME_TYPES) ?? DEFAULT_ANIME_STATE.type
  const genre = toInt(first(query.genre))
  const decade = toInt(first(query.decade))
  const year = toInt(first(query.year))
  return normalizeAnimeState({
    type,
    sort: pick(first(query.sort), ANIME_SORTS[type]) ?? DEFAULT_ANIME_STATE.sort,
    genre: genre !== null && genre > 0 ? genre : null,
    keyword: pick(first(query.keyword), ANIME_KEYWORDS),
    country: pick(first(query.country), ANIME_COUNTRIES) ?? DEFAULT_ANIME_STATE.country,
    decade: decade !== null && animeDecades(currentYear).includes(decade) ? decade : null,
    year: year !== null && animeYears(currentYear).includes(year) ? year : null,
    season: pick(first(query.season), ANIME_SEASONS),
  })
}

/** Query object for the state; values equal to their default are omitted to keep URLs short. */
export function serializeAnimeQuery(state: AnimeBrowseState): Record<string, string> {
  const d = DEFAULT_ANIME_STATE
  const out: Record<string, string> = {}
  if (state.type !== d.type) out.type = state.type
  if (state.sort !== d.sort) out.sort = state.sort
  if (state.genre !== null) out.genre = String(state.genre)
  if (state.keyword !== null) out.keyword = state.keyword
  if (state.country !== d.country) out.country = state.country
  if (state.decade !== null) out.decade = String(state.decade)
  if (state.year !== null) out.year = String(state.year)
  if (state.season !== null) out.season = state.season
  return out
}

/** Replaces the anime keys of an existing query and keeps every other key. */
export function mergeAnimeQuery(current: LocationQuery, state: AnimeBrowseState): LocationQueryRaw {
  const rest: LocationQueryRaw = { ...current }
  for (const key of ANIME_QUERY_KEYS) delete rest[key]
  return { ...rest, ...serializeAnimeQuery(state) }
}

/**
 * One user change, applied to the state with the API rules kept.
 * Switching type clears the genre (the ids differ) and TV-only options.
 * Picking a decade or a year replaces the other one.
 */
export function applyAnimeChange(state: AnimeBrowseState, patch: Partial<AnimeBrowseState>): AnimeBrowseState {
  let next: AnimeBrowseState = { ...state, ...patch }
  if (patch.type !== undefined && patch.type !== state.type) next = { ...next, genre: null }
  if (patch.decade) next = { ...next, year: null, season: null }
  if (patch.year) next = { ...next, decade: null }
  return normalizeAnimeState(next)
}

/** Params for the anime endpoints (the page number is added by the caller). */
export function toAnimeApiQuery(state: AnimeBrowseState): AnimeQuery {
  return {
    sort: state.sort,
    country: state.country,
    genre: state.genre ?? undefined,
    keyword: state.keyword ?? undefined,
    decade: state.decade ?? undefined,
    year: state.year ?? undefined,
    season: state.season ?? undefined,
  }
}
