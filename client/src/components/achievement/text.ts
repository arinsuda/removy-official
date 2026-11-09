import { useI18n } from "vue-i18n"
import type { Achievement } from "@/types/achievement"

export const REQUIREMENT_KEY_PREFIX = "achievements.requirements."

/** Requirement copy from i18n by action type; falls back to the server description when a key is missing. */
export function useAchievementText() {
  const { t, te, n } = useI18n()
  function description(a: Achievement): string {
    const key = `${REQUIREMENT_KEY_PREFIX}${a.action_type}`
    return te(key) ? t(key, { target: n(a.target_count) }, a.target_count) : a.description
  }
  return { description }
}
