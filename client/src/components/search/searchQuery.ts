import type { LocationQuery, LocationQueryRaw } from "vue-router"
import { BROWSE_SORTS, type BrowseSort } from "@/components/browse/browseSort"

export const SEARCH_TABS = ["all", "movie", "tv", "people"] as const
export type SearchTab = (typeof SEARCH_TABS)[number]

export interface SearchState {
  q: string
  tab: SearchTab
  sort: BrowseSort
  year: number | null
  genre: number | null
}

export const DEFAULT_SEARCH_STATE: SearchState = { q: "", tab: "all", sort: "default", year: null, genre: null }
export const SEARCH_QUERY_KEYS = ["q", "tab", "sort", "year", "genre"] as const

const first = (v: LocationQuery[string] | undefined): string => (Array.isArray(v) ? (v[0] ?? "") : (v ?? ""))

function oneOf<T extends string>(raw: string, allowed: readonly T[], fallback: T): T {
  return (allowed as readonly string[]).includes(raw) ? (raw as T) : fallback
}

const positiveInt = (raw: string): number | null => {
  const n = Number(raw)
  return Number.isInteger(n) && n > 0 ? n : null
}

/** Reads ?q&tab&sort&year&genre; anything unknown falls back to the default. */
export function parseSearchQuery(query: LocationQuery): SearchState {
  const year = positiveInt(first(query.year))
  return {
    q: first(query.q).trim(),
    tab: oneOf(first(query.tab), SEARCH_TABS, DEFAULT_SEARCH_STATE.tab),
    sort: oneOf(first(query.sort), BROWSE_SORTS, DEFAULT_SEARCH_STATE.sort),
    year: year && year >= 1870 && year <= 2200 ? year : null,
    genre: positiveInt(first(query.genre)),
  }
}

/** Values equal to their default are omitted to keep URLs short. */
export function serializeSearchQuery(state: SearchState): Record<string, string> {
  const d = DEFAULT_SEARCH_STATE
  const out: Record<string, string> = {}
  if (state.q.trim()) out.q = state.q.trim()
  if (state.tab !== d.tab) out.tab = state.tab
  if (state.sort !== d.sort) out.sort = state.sort
  if (state.year) out.year = String(state.year)
  if (state.genre) out.genre = String(state.genre)
  return out
}

/** Replaces the search keys of an existing query and keeps every other key. */
export function mergeSearchQuery(current: LocationQuery, state: SearchState): LocationQueryRaw {
  const rest: LocationQueryRaw = { ...current }
  for (const key of SEARCH_QUERY_KEYS) delete rest[key]
  return { ...rest, ...serializeSearchQuery(state) }
}
