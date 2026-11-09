import { BookOpen, Clapperboard, Compass, Heart, Library, MessageSquareText, Sparkles, type LucideIcon } from "lucide-vue-next"
import type { RouteLocationRaw } from "vue-router"

export type CategoryKey = "watching" | "genres" | "cinema" | "reviews" | "social" | "library" | "journey"
export type StatusFilter = "all" | "unlocked" | "progress" | "locked"
export type SortKey = "recent" | "progress" | "name"

/** Server maximum page size; the catalog (about 105 rows) needs a couple of pages. */
export const PAGE_LIMIT = 100
export const MAX_PAGES = 5
/** A section with this many cards or fewer stretches them to fill the row. */
export const FEW_UNITS = 2
/** Grid cards animate in with a stagger; later cards share the last delay step. */
export const STAGGER_STEPS = 12

export const STATUS_FILTERS: readonly StatusFilter[] = ["all", "unlocked", "progress", "locked"]
export const SORT_KEYS: readonly SortKey[] = ["recent", "progress", "name"]
export const DEFAULT_SORT: SortKey = "recent"

/** Tier labels for families with several levels of the same goal. */
export const TIER_NUMERALS = ["I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"] as const

interface CategoryConfig {
  icon: LucideIcon
  /** Exact action types and prefixes (ending with "_") that belong to the category. */
  types: readonly string[]
  prefixes: readonly string[]
}

/** Display order of the sections; the first match wins. */
export const CATEGORIES: Record<CategoryKey, CategoryConfig> = {
  watching: {
    icon: Clapperboard,
    types: ["watched_count", "watched_in_one_day", "series_completed_count", "classic_movie_watched_count", "golden_age_movie_watched_count", "long_movie_watched_count", "short_movie_watched_count", "remake_watched_count", "new_release_early_watched_count", "watched_oscar_best_picture_count"],
    prefixes: [],
  },
  genres: { icon: Sparkles, types: ["genre_all_unlocked"], prefixes: ["genre_watch_"] },
  cinema: { icon: Compass, types: ["unique_country_count"], prefixes: ["actor_watch_", "director_watch_", "country_watch_", "studio_watch_"] },
  reviews: {
    icon: MessageSquareText,
    types: ["review_count", "review_daily_streak", "reviews_in_one_day", "contrarian_review_count", "obscure_movie_review_count", "rating_count", "rating_five_star_count", "rating_one_star_count", "low_rating_count", "sequel_rated_lower_count"],
    prefixes: [],
  },
  social: { icon: Heart, types: ["follower_count", "following_count", "like_count", "review_like_given_count", "review_like_received_count"], prefixes: [] },
  library: { icon: Library, types: ["library_total_count", "watchlist_count", "series_watchlist_no_watch_count"], prefixes: [] },
  journey: { icon: BookOpen, types: [], prefixes: [] },
}

export const CATEGORY_ORDER = Object.keys(CATEGORIES) as CategoryKey[]
/** Anything unmapped lands here so new seeded achievements still show up. */
export const FALLBACK_CATEGORY: CategoryKey = "journey"

export function categoryOf(actionType: string): CategoryKey {
  for (const key of CATEGORY_ORDER) {
    const c = CATEGORIES[key]
    if (c.types.includes(actionType) || c.prefixes.some((p) => actionType.startsWith(p))) return key
  }
  return FALLBACK_CATEGORY
}

export interface AchievementCta {
  labelKey: string
  to: (userId: number) => RouteLocationRaw
}

const BROWSE: AchievementCta = { labelKey: "achievements.cta.browse", to: () => ({ name: "movies" }) }

/** Owner-only tips shown in the detail modal. Action types without an entry fall back to their category. */
const CTA_BY_TYPE: Record<string, AchievementCta | null> = {
  profile_complete: { labelKey: "achievements.cta.profile", to: () => ({ name: "settings" }) },
  review_count: { labelKey: "achievements.cta.review", to: () => ({ name: "movies" }) },
  login_streak: null,
  reach_level: null,
  early_adopter: null,
  achievement_unlocked_count: null,
  midnight_activity_count: null,
}

const CTA_BY_CATEGORY: Record<CategoryKey, AchievementCta | null> = {
  watching: BROWSE,
  genres: BROWSE,
  cinema: BROWSE,
  reviews: { labelKey: "achievements.cta.rate", to: () => ({ name: "movies" }) },
  social: { labelKey: "achievements.cta.discover", to: () => ({ name: "feed" }) },
  library: { labelKey: "achievements.cta.library", to: (userId) => ({ name: "user-library", params: { userId } }) },
  journey: null,
}

export function ctaFor(actionType: string): AchievementCta | null {
  if (actionType in CTA_BY_TYPE) return CTA_BY_TYPE[actionType] ?? null
  return CTA_BY_CATEGORY[categoryOf(actionType)]
}
