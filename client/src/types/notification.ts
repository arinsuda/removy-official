export type NotificationType =
  | "followed_you"
  | "following_reviewed"
  | "following_liked_review"
  | "following_added_watchlist"
  | "following_added_watched"
  | "following_marked_helpful"
  | "following_commented"
  | "following_liked_media"
  | "review_liked"
  | "review_commented"
  | "review_marked_helpful"
  | "achievement_unlocked"
  | "movie_now_playing"
  | "email_verified"
  | "password_changed"
  | "welcome"
  | "follow_requested"
  | "follow_accepted"

export type NotificationCategory = "system" | "social" | "media" | "achievement"

export interface NotificationActor {
  id: number
  username: string
  display_name?: string | null
  avatar_url?: string | null
}

export interface AppNotification {
  id: number
  type: NotificationType
  category: NotificationCategory
  message: string
  is_read: boolean
  read_at?: string | null
  actor?: NotificationActor | null
  target_id?: number | null
  target_ref?: string | null
  created_at: string
}

export interface NotificationPaginationMeta {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface NotificationListResponse {
  notifications: AppNotification[]
  unread_count: number
  pagination: NotificationPaginationMeta
}

export type SocketNotificationNewPayload = AppNotification
export interface SocketNotificationReadPayload {
  ids: number[]
}
export interface SocketNotificationDeletedPayload {
  ids: number[] | null
}

/** i18n key of the notification title; unknown types fall back to the generic one. */
export function notificationTitleKey(type: NotificationType): string {
  return `notifications.titles.${type}`
}

/**
 * In-app route for a notification, or null when the target has no page.
 * Reviews have no standalone page, so review notifications are not links.
 */
export function getNotificationActionUrl(n: AppNotification): string | null {
  if (!n.target_ref || n.target_id == null) return null
  switch (n.target_ref) {
    case "user":
      return `/users/${n.target_id}`
    case "movie":
      return `/movies/${n.target_id}`
    default:
      return null
  }
}

export const NOTIFICATION_CATEGORIES: NotificationCategory[] = ["social", "media", "achievement", "system"]
