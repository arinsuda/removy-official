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

/** Language-neutral payload; the client builds the text from type + params. */
export interface NotificationParams {
  media_id?: number
  media_type?: string
  media_title?: string
  action_type?: string
  target_count?: number
  exp?: number
  name?: string
}

export interface AppNotification {
  id: number
  type: NotificationType
  category: NotificationCategory
  /** Empty on new rows; only legacy rows carry server-written text. */
  message: string
  params?: NotificationParams | null
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

const MEDIA_ROUTE: Record<string, string> = { movie: "movies", tv: "tv" }

/** In-app route for a notification, or null when the target has no page. */
export function getNotificationActionUrl(n: AppNotification): string | null {
  if (!n.target_ref || n.target_id == null) return null
  switch (n.target_ref) {
    case "user":
      return `/users/${n.target_id}`
    case "movie":
    case "tv":
      return `/${MEDIA_ROUTE[n.target_ref]}/${n.target_id}`
    case "review": {
      // Reviews have no page of their own; open the title they belong to.
      const { media_id, media_type } = n.params ?? {}
      const base = media_type ? MEDIA_ROUTE[media_type] : undefined
      return base && media_id ? `/${base}/${media_id}` : null
    }
    default:
      return null
  }
}

export const NOTIFICATION_CATEGORIES: NotificationCategory[] = ["social", "media", "achievement", "system"]
