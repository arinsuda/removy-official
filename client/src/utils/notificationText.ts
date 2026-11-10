import { nameKey } from "@/components/achievement/text"
import { notificationTitleKey, type AppNotification } from "@/types/notification"

/** The slice of vue-i18n the resolver needs; works with the composer and `i18n.global`. */
export interface NotificationI18n {
  t: (key: string, named?: Record<string, unknown>) => string
  te: (key: string) => boolean
  n: (value: number) => string
}

const TYPE_KEY_PREFIX = "notifications.types."
const ACTOR_TYPES = new Set([
  "followed_you",
  "follow_requested",
  "follow_accepted",
  "following_reviewed",
  "review_liked",
  "following_liked_review",
  "review_marked_helpful",
  "following_marked_helpful",
  "review_commented",
  "following_commented",
  "following_added_watched",
  "following_added_watchlist",
  "following_liked_media",
])
const MEDIA_TYPES = new Set([
  "following_reviewed",
  "review_liked",
  "following_liked_review",
  "review_marked_helpful",
  "following_marked_helpful",
  "review_commented",
  "following_commented",
  "following_added_watched",
  "following_added_watchlist",
  "following_liked_media",
  "movie_now_playing",
])

/** Display name of the actor, resolved live from the notification's actor. */
export function notificationActorName(n: AppNotification): string {
  return n.actor?.display_name || n.actor?.username || ""
}

/**
 * Notification body in the current UI language.
 * 1) the type's i18n key, when every param it needs is present;
 * 2) the legacy server message of old rows;
 * 3) the title key, then the generic title.
 */
export function formatNotification(n: AppNotification, i18n: NotificationI18n): string {
  const { t, te, n: num } = i18n
  const key = `${TYPE_KEY_PREFIX}${n.type}`
  const params = n.params ?? {}

  if (te(key)) {
    const named: Record<string, unknown> = {}
    let complete = true
    if (ACTOR_TYPES.has(n.type)) {
      named.name = notificationActorName(n) || t("notifications.someone")
    }
    if (MEDIA_TYPES.has(n.type)) {
      // New rows always carry params (the title may be empty); legacy rows have none.
      if (params.media_title) named.title = params.media_title
      else if (n.params) named.title = t("notifications.untitled")
      else complete = false
    }
    if (n.type === "achievement_unlocked") {
      const ach = { action_type: params.action_type ?? "", target_count: params.target_count ?? 0 }
      const localized = te(nameKey(ach)) ? t(nameKey(ach)) : ""
      named.name = localized || params.name || ""
      named.exp = num(params.exp ?? 0)
      complete = !!named.name
    }
    if (complete) return t(key, named)
  }

  if (n.message) return n.message
  const titleKey = notificationTitleKey(n.type)
  return te(titleKey) ? t(titleKey) : t("notifications.title")
}
