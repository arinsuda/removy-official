import type { LocationQuery, LocationQueryRaw } from "vue-router"
import type { MangaCountry, MangaFormat, MangaQuery, MangaSort, MangaStatus } from "@/types/manga"
import {
  MANGA_COUNTRIES,
  MANGA_DEFAULT_COUNTRY,
  MANGA_DEFAULT_FORMAT,
  MANGA_DEFAULT_SORT,
  MANGA_DEFAULT_STATUS,
  MANGA_FORMATS,
  MANGA_GENRE_MAX_LENGTH,
  MANGA_SORTS,
  MANGA_STATUSES,
} from "./mangaConfig"

const ADULT_ON = "1"

export const MANGA_QUERY_KEYS = ["country", "format", "sort", "genre", "status", "adult"] as const

export interface MangaBrowseState {
  country: MangaCountry
  format: MangaFormat
  sort: MangaSort
  /** Genre name, or null for every genre. */
  genre: string | null
  status: MangaStatus
  /** The "Show 18+ titles" toggle. Sent to the API only with canShowAdult (see useMangaAdult). */
  adult: boolean
}

export const DEFAULT_MANGA_STATE: MangaBrowseState = {
  country: MANGA_DEFAULT_COUNTRY,
  format: MANGA_DEFAULT_FORMAT,
  sort: MANGA_DEFAULT_SORT,
  genre: null,
  status: MANGA_DEFAULT_STATUS,
  adult: false,
}

const first = (v: LocationQuery[string] | undefined): string => (Array.isArray(v) ? (v[0] ?? "") : (v ?? ""))

/** The allowed value, or null when the raw text is not one of them. */
function pick<T extends string>(raw: string, allowed: readonly T[]): T | null {
  return (allowed as readonly string[]).includes(raw) ? (raw as T) : null
}

/** Genre from URL text: trimmed, non-empty and not absurdly long. */
function toGenre(raw: string): string | null {
  const genre = raw.trim()
  return genre !== "" && genre.length <= MANGA_GENRE_MAX_LENGTH ? genre : null
}

/** Reads ?country&format&sort&genre&status. Unknown values fall back to the default. */
export function parseMangaQuery(query: LocationQuery): MangaBrowseState {
  return {
    country: pick(first(query.country), MANGA_COUNTRIES) ?? DEFAULT_MANGA_STATE.country,
    format: pick(first(query.format), MANGA_FORMATS) ?? DEFAULT_MANGA_STATE.format,
    sort: pick(first(query.sort), MANGA_SORTS) ?? DEFAULT_MANGA_STATE.sort,
    genre: toGenre(first(query.genre)),
    status: pick(first(query.status), MANGA_STATUSES) ?? DEFAULT_MANGA_STATE.status,
    adult: first(query.adult) === ADULT_ON,
  }
}

/** Values equal to their default are omitted to keep URLs short. */
export function serializeMangaQuery(state: MangaBrowseState): Record<string, string> {
  const d = DEFAULT_MANGA_STATE
  const out: Record<string, string> = {}
  if (state.country !== d.country) out.country = state.country
  if (state.format !== d.format) out.format = state.format
  if (state.sort !== d.sort) out.sort = state.sort
  if (state.genre !== null) out.genre = state.genre
  if (state.status !== d.status) out.status = state.status
  if (state.adult) out.adult = ADULT_ON
  return out
}

/** Replaces the manga keys of an existing query and keeps every other key. */
export function mergeMangaQuery(current: LocationQuery, state: MangaBrowseState): LocationQueryRaw {
  const rest: LocationQueryRaw = { ...current }
  for (const key of MANGA_QUERY_KEYS) delete rest[key]
  return { ...rest, ...serializeMangaQuery(state) }
}

/** Params for the manga list endpoint (the page number is added by the caller). */
export function toMangaApiQuery(state: MangaBrowseState, includeAdult: boolean): MangaQuery {
  return {
    country: state.country,
    format: state.format,
    sort: state.sort,
    status: state.status,
    genre: state.genre ?? undefined,
    include_adult: includeAdult,
  }
}
