import { Bookmark, Eye, Heart, Star } from "lucide-vue-next"
import type { Component } from "vue"
import type { LibraryListKey, LibraryTab } from "@/types"

export const LIBRARY_TABS: readonly LibraryTab[] = ["dashboard", "lists", "bmol"]
export const DEFAULT_LIBRARY_TAB: LibraryTab = "dashboard"
export const LIBRARY_LIST_KEYS: readonly LibraryListKey[] = ["watchlist", "watched", "likes", "reviews"]

/** Route query keys: ?tab=<LibraryTab>&list=<LibraryListKey>. */
export const TAB_QUERY_KEY = "tab"
export const LIST_QUERY_KEY = "list"

export const LIST_ICONS: Record<LibraryListKey, Component> = {
  watchlist: Bookmark,
  watched: Eye,
  likes: Heart,
  reviews: Star,
}

/** Posters shown in a list card cover. */
export const COLLAGE_SIZE = 4
/** Titles in the dashboard "up next" and "recently watched" rails. */
export const RAIL_SIZE = 6
export const TOP_GENRES = 5

export const RATING_MAX = 5
export const RATING_STEP = 0.5

/** Add-titles sheet. */
export const ADD_LIST_TYPE = "watchlist" as const
export const ADD_SEARCH_DEBOUNCE_MS = 300
export const ADD_SEARCH_LIMIT = 8
export const ADD_POSTER_SIZE = "w92"
export const COLLAGE_POSTER_SIZE = "w185"
