import { Bookmark, CheckCircle2, Eye, Film, Heart, MessageSquare, PenTool, Repeat2, Trophy, UserPlus, type LucideIcon } from "lucide-vue-next"
import type { ActivitySettingsResponse } from "@/types/feed"

/** Activity types a user can show or hide in followers' feeds. Labels live at `settings.activity.options.<key>`. */
export const ACTIVITY_OPTIONS: { key: keyof ActivitySettingsResponse; icon: LucideIcon }[] = [
  { key: "review_created", icon: PenTool },
  { key: "review_commented", icon: MessageSquare },
  { key: "review_liked", icon: Heart },
  { key: "media_liked", icon: Film },
  { key: "watchlist_added", icon: Bookmark },
  { key: "watched_added", icon: CheckCircle2 },
  { key: "watch_log_created", icon: Eye },
  { key: "rewatched", icon: Repeat2 },
  { key: "achievement_unlocked", icon: Trophy },
  { key: "user_followed", icon: UserPlus },
]
