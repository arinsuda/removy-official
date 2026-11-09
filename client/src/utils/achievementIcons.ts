import {
  UserCheck,
  PenLine,
  Eye,
  Heart,
  Flame,
  Star,
  type LucideIcon,
} from "lucide-vue-next"
import { CATEGORIES, categoryOf } from "@/components/achievement/config"

const ICON_MAP: Record<string, LucideIcon> = {
  profile_complete: UserCheck,
  review_count: PenLine,
  watchlist_count: Eye,
  like_count: Heart,
  login_streak: Flame,
  rating_count: Star,
}

export function getAchievementIcon(actionType: string): LucideIcon {
  return ICON_MAP[actionType] ?? CATEGORIES[categoryOf(actionType)].icon
}

const STATIC_LABELS: Record<string, string> = {
  achievement_unlocked_count: "จำนวน Achievement ที่ปลดล็อก",
  classic_movie_watched_count: "ดูหนังคลาสสิก",
  contrarian_review_count: "รีวิวสวนกระแส",
  early_adopter: "ผู้บุกเบิก",
  follower_count: "จำนวนผู้ติดตาม",
  following_count: "จำนวนที่กำลังติดตาม",
  genre_all_unlocked: "ปลดล็อกทุกหมวดหมู่หนัง",
}

const ACTOR_NAMES: Record<string, string> = {
  meryl_streep: "Meryl Streep",
  rdj: "Robert Downey Jr.",
}

const DIRECTOR_NAMES: Record<string, string> = {
  miyazaki: "Hayao Miyazaki",
  nolan: "Christopher Nolan",
  tarantino: "Quentin Tarantino",
}

const COUNTRY_NAMES: Record<string, string> = {
  france: "ฝรั่งเศส",
  india: "อินเดีย",
  japan: "ญี่ปุ่น",
  korea: "เกาหลี",
  thailand: "ไทย",
}

const GENRE_NAMES: Record<string, string> = {
  action: "แอ็กชัน",
  animation: "แอนิเมชัน",
  comedy: "คอมเมดี้",
  drama: "ดราม่า",
  horror: "สยองขวัญ",
  romance: "โรแมนติก",
  scifi: "ไซไฟ",
  thriller: "ทริลเลอร์",
}

/** fallback: แปลง snake_case ให้เป็นคำอ่านง่ายๆ */
function humanize(str: string): string {
  return str
    .split("_")
    .map(w => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ")
}

/** แปลง action_type raw string -> label ที่อ่านเข้าใจ */
export function getActionTypeLabel(type: string): string {
  if (STATIC_LABELS[type]) return STATIC_LABELS[type]

  if (type.startsWith("actor_watch_")) {
    const key = type.replace("actor_watch_", "")
    return `ดูหนังของ ${ACTOR_NAMES[key] ?? humanize(key)}`
  }
  if (type.startsWith("director_watch_")) {
    const key = type.replace("director_watch_", "")
    return `ดูหนังของผู้กำกับ ${DIRECTOR_NAMES[key] ?? humanize(key)}`
  }
  if (type.startsWith("country_watch_")) {
    const key = type.replace("country_watch_", "")
    return `ดูหนังจากประเทศ${COUNTRY_NAMES[key] ?? humanize(key)}`
  }
  if (type.startsWith("genre_watch_")) {
    const key = type.replace("genre_watch_", "")
    return `ดูหนังแนว${GENRE_NAMES[key] ?? humanize(key)}`
  }

  return humanize(type)
}
