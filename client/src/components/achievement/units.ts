import type { UserAchievement } from "@/types/achievement"
import { progressOf } from "./achievements"
import { CATEGORY_ORDER, categoryOf, type CategoryKey, type SortKey, type StatusFilter } from "./config"

export type UnitStatus = Exclude<StatusFilter, "all">

/** One card on the page: a single badge, or a family of tiers that share an action type. */
export interface AchievementUnit {
  key: string
  category: CategoryKey
  /** Tiers by ascending target. A single badge has exactly one. */
  tiers: UserAchievement[]
  /** First locked tier, or the last one when every tier is unlocked. */
  current: UserAchievement
  status: UnitStatus
  /** Progress of the current tier, 0-100. */
  pct: number
  /** Newest unlock time in ms, 0 when nothing is unlocked. */
  lastUnlockedAt: number
  /** Ids of the tiers that pass the active filters; set by `filterUnits`, absent on raw units. */
  matched?: ReadonlySet<number>
}

export interface CategorySummary {
  key: CategoryKey
  units: AchievementUnit[]
  unlocked: number
  total: number
}

export interface OverallSummary {
  unlocked: number
  total: number
  pct: number
  xp: number
}

const time = (a: UserAchievement) => (a.unlocked_at ? new Date(a.unlocked_at).getTime() : 0)

function makeUnit(key: string, group: UserAchievement[]): AchievementUnit {
  const tiers = [...group].sort((a, b) => a.achievement.target_count - b.achievement.target_count || a.achievement_id - b.achievement_id)
  const current = tiers.find((t) => !t.is_unlocked) ?? tiers[tiers.length - 1]!
  const unlockedCount = tiers.filter((t) => t.is_unlocked).length
  const started = unlockedCount > 0 || progressOf(current) > 0
  const status: UnitStatus = unlockedCount === tiers.length ? "unlocked" : started ? "progress" : "locked"
  return {
    key,
    category: categoryOf(tiers[0]!.achievement.action_type),
    tiers,
    current,
    status,
    pct: status === "unlocked" ? 100 : progressOf(current),
    lastUnlockedAt: Math.max(0, ...tiers.map(time)),
  }
}

/** Groups tiers of the same action type into families; the rest stay single badges. */
export function buildUnits(list: UserAchievement[]): AchievementUnit[] {
  const byType = new Map<string, UserAchievement[]>()
  for (const a of list) byType.set(a.achievement.action_type, [...(byType.get(a.achievement.action_type) ?? []), a])
  const units: AchievementUnit[] = []
  for (const [type, group] of byType) {
    if (group.length > 1) units.push(makeUnit(`family:${type}`, group))
    else units.push(makeUnit(`single:${group[0]!.achievement_id}`, group))
  }
  return units
}

export function summarize(list: UserAchievement[]): OverallSummary {
  const done = list.filter((a) => a.is_unlocked)
  return { unlocked: done.length, total: list.length, pct: list.length ? Math.round((done.length / list.length) * 100) : 0, xp: done.reduce((sum, a) => sum + a.achievement.exp_reward, 0) }
}

export function latestUnlock(list: UserAchievement[]): UserAchievement | null {
  return list.filter((a) => a.is_unlocked).sort((a, b) => time(b) - time(a))[0] ?? null
}

/** Locked achievement closest to done; ties go to the smaller target. */
export function nextUp(list: UserAchievement[]): UserAchievement | null {
  return list.filter((a) => !a.is_unlocked).sort((a, b) => progressOf(b) - progressOf(a) || a.achievement.target_count - b.achievement.target_count)[0] ?? null
}

export interface UnitQuery {
  status: StatusFilter
  category: CategoryKey | "all"
  search: string
  /** Text shown for a requirement, so search matches what the user reads. */
  describe?: (a: UserAchievement) => string
}

/** Status of a single tier; a locked tier counts as in progress once it has any progress. */
export function tierStatus(a: UserAchievement): UnitStatus {
  return a.is_unlocked ? "unlocked" : progressOf(a) > 0 ? "progress" : "locked"
}

function matchesSearch(a: UserAchievement, search: string, describe?: UnitQuery["describe"]): boolean {
  const needle = search.trim().toLowerCase()
  if (!needle) return true
  return a.achievement.name.toLowerCase().includes(needle) || (describe?.(a) ?? a.achievement.description).toLowerCase().includes(needle)
}

/** Keeps units with at least one tier matching status and search; `matched` lists those tiers. */
export function filterUnits(units: AchievementUnit[], q: UnitQuery): AchievementUnit[] {
  const out: AchievementUnit[] = []
  for (const u of units) {
    if (q.category !== "all" && u.category !== q.category) continue
    const hits = u.tiers.filter((t) => (q.status === "all" || tierStatus(t) === q.status) && matchesSearch(t, q.search, q.describe))
    if (hits.length > 0) out.push({ ...u, matched: new Set(hits.map((t) => t.achievement_id)) })
  }
  return out
}

/** Number of individual achievements (tiers) that pass the filters. */
export function matchedCount(units: AchievementUnit[]): number {
  return units.reduce((sum, u) => sum + (u.matched?.size ?? u.tiers.length), 0)
}

export function sortUnits(units: AchievementUnit[], sort: SortKey, locale: string): AchievementUnit[] {
  const byName = (a: AchievementUnit, b: AchievementUnit) => a.current.achievement.name.localeCompare(b.current.achievement.name, locale)
  const cmp: Record<SortKey, (a: AchievementUnit, b: AchievementUnit) => number> = {
    recent: (a, b) => b.lastUnlockedAt - a.lastUnlockedAt || b.pct - a.pct || byName(a, b),
    progress: (a, b) => Number(a.status === "unlocked") - Number(b.status === "unlocked") || b.pct - a.pct || byName(a, b),
    name: byName,
  }
  return [...units].sort(cmp[sort])
}

/** Sections in config order; empty categories are dropped. */
export function groupByCategory(units: AchievementUnit[]): CategorySummary[] {
  return CATEGORY_ORDER.map((key) => {
    const own = units.filter((u) => u.category === key)
    const tiers = own.flatMap((u) => u.tiers)
    return { key, units: own, unlocked: tiers.filter((t) => t.is_unlocked).length, total: tiers.length }
  }).filter((c) => c.units.length > 0)
}

/** Counts individual achievements per status; pass units already filtered by status "all". */
export function countByStatus(units: AchievementUnit[]): Record<StatusFilter, number> {
  const out: Record<StatusFilter, number> = { all: 0, unlocked: 0, progress: 0, locked: 0 }
  for (const u of units) {
    for (const t of u.tiers) {
      if (u.matched && !u.matched.has(t.achievement_id)) continue
      out.all++
      out[tierStatus(t)]++
    }
  }
  return out
}
