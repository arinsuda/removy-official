import type { UserAchievement } from "@/types/achievement"

export type AchievementFilter = "all" | "unlocked" | "locked"

export function countUnlocked(list: UserAchievement[]): number {
  return list.filter((a) => a.is_unlocked).length
}

/** Progress as 0-100, derived from counts when the server percentage is missing. */
export function progressOf(a: UserAchievement): number {
  const target = a.achievement.target_count
  const pct = a.progress_pct || (target > 0 ? (a.current_count / target) * 100 : 0)
  return Math.max(0, Math.min(100, Math.round(pct)))
}

/** Unlocked first (newest unlock on top), then locked by how close they are. */
export function sortAchievements(list: UserAchievement[]): UserAchievement[] {
  const when = (a: UserAchievement) => (a.unlocked_at ? new Date(a.unlocked_at).getTime() : 0)
  return [...list].sort((a, b) => Number(b.is_unlocked) - Number(a.is_unlocked) || (a.is_unlocked ? when(b) - when(a) : progressOf(b) - progressOf(a)))
}

export function filterAchievements(list: UserAchievement[], filter: AchievementFilter): UserAchievement[] {
  if (filter === "unlocked") return list.filter((a) => a.is_unlocked)
  if (filter === "locked") return list.filter((a) => !a.is_unlocked)
  return list
}
