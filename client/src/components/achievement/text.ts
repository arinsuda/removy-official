import { useI18n } from "vue-i18n"
import type { Achievement } from "@/types/achievement"

export const REQUIREMENT_KEY_PREFIX = "achievements.requirements."
export const NAME_KEY_PREFIX = "achievements.names."

/** Stable i18n key for an achievement: names are keyed by action type and target, which never change. */
export function nameKey(a: Pick<Achievement, "action_type" | "target_count">): string {
  return `${NAME_KEY_PREFIX}${a.action_type}_${a.target_count}`
}

/** Name and requirement copy from i18n; each falls back to the server text when its key is missing. */
export function useAchievementText() {
  const { t, te, n } = useI18n()
  function description(a: Achievement): string {
    const key = `${REQUIREMENT_KEY_PREFIX}${a.action_type}`
    return te(key) ? t(key, { target: n(a.target_count) }, a.target_count) : a.description
  }
  function name(a: Achievement): string {
    const key = nameKey(a)
    return te(key) ? t(key) : a.name
  }
  return { name, description }
}
