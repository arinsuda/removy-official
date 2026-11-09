import type { Component } from "vue"
import type { ActivityType } from "@/types/feed"
import {
  PenTool,
  MessageSquare,
  Heart,
  Bookmark,
  CheckCircle2,
  Trophy,
  UserPlus,
  HelpCircle,
  Eye,
  Repeat,
} from "lucide-vue-next"

export type ActivityTone = "success" | "danger" | "info" | "brand" | "warning" | "muted"

interface ActivityMeta {
  /** i18n key of the short label, e.g. "feed.types.review_created.label". */
  labelKey: string
  icon: Component
  tone: ActivityTone
}

const ACTIVITY_META: Record<ActivityType, Omit<ActivityMeta, "labelKey">> = {
  review_created: { icon: PenTool, tone: "success" },
  review_commented: { icon: MessageSquare, tone: "success" },
  review_liked: { icon: Heart, tone: "danger" },
  media_liked: { icon: Heart, tone: "danger" },
  watchlist_added: { icon: Bookmark, tone: "info" },
  watched_added: { icon: CheckCircle2, tone: "brand" },
  achievement_unlocked: { icon: Trophy, tone: "warning" },
  user_followed: { icon: UserPlus, tone: "info" },
  watch_log_created: { icon: Eye, tone: "brand" },
  rewatched: { icon: Repeat, tone: "brand" },
}

export function getActivityMeta(type: ActivityType): ActivityMeta {
  const meta = ACTIVITY_META[type] ?? { icon: HelpCircle, tone: "muted" as const }
  return { ...meta, labelKey: ACTIVITY_META[type] ? `feed.types.${type}.label` : "feed.types.unknown.label" }
}

// Maps the backend media_type to the client route prefix.
export function mediaRoutePrefix(mediaType: string): "movies" | "tv" {
  return mediaType === "tv" || mediaType === "series" ? "tv" : "movies"
}
