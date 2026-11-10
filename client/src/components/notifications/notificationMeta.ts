import type { Component } from "vue"
import {
  Bell,
  Clapperboard,
  Heart,
  MessageCircle,
  PartyPopper,
  ThumbsUp,
  Trophy,
  UserCheck,
  UserPlus,
} from "lucide-vue-next"
import type { NotificationType } from "@/types/notification"

export const DROPDOWN_LIMIT = 5

/** How long a just-arrived notification keeps its highlight, in ms. */
export const FRESH_HIGHLIGHT_MS = 2400

const ICONS: Partial<Record<NotificationType, Component>> = {
  followed_you: UserPlus,
  follow_requested: UserPlus,
  follow_accepted: UserCheck,
  review_liked: Heart,
  following_liked_review: Heart,
  review_commented: MessageCircle,
  following_commented: MessageCircle,
  review_marked_helpful: ThumbsUp,
  following_marked_helpful: ThumbsUp,
  achievement_unlocked: Trophy,
  movie_now_playing: Clapperboard,
  welcome: PartyPopper,
}

export function notificationIcon(type: NotificationType): Component {
  return ICONS[type] ?? Bell
}

