import { Bookmark, Eye, Globe, Heart, Lock, Star, Users } from "lucide-vue-next"
import type { Component } from "vue"
import type { LibraryListKey, LibraryTab, Visibility } from "@/types"

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

/** Custom lists. Limits mirror server/internal/list_module. */
export const CUSTOM_LIST_PREFIX = "custom-"
export const MAX_CUSTOM_LISTS = 50
export const MAX_LIST_ITEMS = 500
export const LIST_NAME_MAX = 80
export const LIST_DESC_MAX = 500
export const DEFAULT_LIST_VISIBILITY: Visibility = "public"
export const LIST_VISIBILITIES: readonly Visibility[] = ["public", "followers", "private"]
export const LIST_VISIBILITY_ICONS: Record<Visibility, Component> = { public: Globe, followers: Users, private: Lock }
export const COVER_POSTER_SIZE = "w342"
export const COVER_PICK_SIZE = "w92"
/** Pointer travel (px) before a handle press becomes a drag. */
export const LIST_DRAG_THRESHOLD = 6
/** Server error codes with their own message at library.customLists.errors.<code>. */
export const LIST_ERROR_CODES = ["LIST_NOT_FOUND", "ITEM_NOT_FOUND", "ITEM_EXISTS", "LIST_LIMIT", "LIST_FULL", "ORDER_MISMATCH", "INVALID_INPUT", "INVALID_VISIBILITY", "INVALID_MEDIA_TYPE"] as const
