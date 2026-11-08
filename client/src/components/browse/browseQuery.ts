import type { LocationQuery, LocationQueryRaw } from "vue-router"
import type { PopcornTier } from "@/components/movie/popcornTier"
import { BROWSE_TABS, type BrowseTab } from "./browseConfig"
import { BROWSE_SORTS, type BrowseSort } from "./browseSort"

export type BrowseTierFilter = "all" | PopcornTier

export const BROWSE_TIERS: BrowseTierFilter[] = ["all", "hot", "fresh", "stale", "none"]
export const BROWSE_QUERY_KEYS = ["tab", "genre", "sort", "tier", "q"] as const

export interface BrowseState {
  tab: BrowseTab
  genre: number | null
  sort: BrowseSort
  tier: BrowseTierFilter
  q: string
}

export const DEFAULT_BROWSE_STATE: BrowseState = { tab: "popular", genre: null, sort: "default", tier: "all", q: "" }

const first = (v: LocationQuery[string] | undefined): string => (Array.isArray(v) ? (v[0] ?? "") : (v ?? ""))

function oneOf<T extends string>(raw: string, allowed: readonly T[], fallback: T): T {
  return (allowed as readonly string[]).includes(raw) ? (raw as T) : fallback
}

/** Reads ?tab&genre&sort&tier&q; anything unknown falls back to the default. */
export function parseBrowseQuery(query: LocationQuery): BrowseState {
  const genre = Number(first(query.genre))
  return {
    tab: oneOf(first(query.tab), BROWSE_TABS, DEFAULT_BROWSE_STATE.tab),
    genre: Number.isInteger(genre) && genre > 0 ? genre : null,
    sort: oneOf(first(query.sort), BROWSE_SORTS, DEFAULT_BROWSE_STATE.sort),
    tier: oneOf(first(query.tier), BROWSE_TIERS, DEFAULT_BROWSE_STATE.tier),
    q: first(query.q).trim(),
  }
}

/** Query object for the state; values equal to their default are omitted to keep URLs short. */
export function serializeBrowseQuery(state: BrowseState): Record<string, string> {
  const d = DEFAULT_BROWSE_STATE
  const out: Record<string, string> = {}
  if (state.tab !== d.tab) out.tab = state.tab
  if (state.genre) out.genre = String(state.genre)
  if (state.sort !== d.sort) out.sort = state.sort
  if (state.tier !== d.tier) out.tier = state.tier
  if (state.q.trim()) out.q = state.q.trim()
  return out
}

/** Replaces the browse keys of an existing query and keeps every other key. */
export function mergeBrowseQuery(current: LocationQuery, state: BrowseState): LocationQueryRaw {
  const rest: LocationQueryRaw = { ...current }
  for (const key of BROWSE_QUERY_KEYS) delete rest[key]
  return { ...rest, ...serializeBrowseQuery(state) }
}
