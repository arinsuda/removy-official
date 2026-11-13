import type { Component } from "vue"
import type { RouteLocationNormalizedLoaded } from "vue-router"
import { CalendarClock, Film, House, Info, Megaphone, Tv } from "lucide-vue-next"

export interface NavItem {
  key: string
  to: string
  labelKey: string
  icon: Component
  isActive: (route: RouteLocationNormalizedLoaded) => boolean
}

const nameStarts = (route: RouteLocationNormalizedLoaded, prefix: string) =>
  route.name?.toString().startsWith(prefix) ?? false

export const NAV_ITEMS: NavItem[] = [
  { key: "home", to: "/", labelKey: "navigation.home", icon: House, isActive: (r) => r.name === "home" },
  { key: "movies", to: "/movies", labelKey: "navigation.movies", icon: Film, isActive: (r) => r.path.startsWith("/movies") },
  { key: "tv", to: "/tv", labelKey: "navigation.tvSeries", icon: Tv, isActive: (r) => r.path.startsWith("/tv") },
  { key: "upcoming", to: "/upcoming", labelKey: "navigation.upcoming", icon: CalendarClock, isActive: (r) => nameStarts(r, "upcoming") },
  { key: "events", to: "/events", labelKey: "navigation.events", icon: Megaphone, isActive: (r) => nameStarts(r, "event") },
  { key: "about", to: "/about-us", labelKey: "navigation.aboutUs", icon: Info, isActive: (r) => nameStarts(r, "aboutUs") },
]

/** Bell badge text: hidden at 0 (null), capped at 99+. */
export function formatBadge(count: number): string | null {
  if (!count || count < 1) return null
  return count > 99 ? "99+" : String(count)
}
