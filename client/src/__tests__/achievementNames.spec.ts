import { describe, expect, it } from "vitest"
import { defineComponent } from "vue"
import { mount } from "@vue/test-utils"
import { nameKey, useAchievementText } from "@/components/achievement/text"
import { filterUnits, sortUnits, type AchievementUnit } from "@/components/achievement/units"
import type { Achievement, UserAchievement } from "@/types/achievement"
import seed from "../../../server/database/seeder/achievement.json"
import en from "@/i18n/locales/en/achievements.json"
import th from "@/i18n/locales/th/achievements.json"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()

function withText<T>(fn: (text: ReturnType<typeof useAchievementText>) => T): T {
  let out!: T
  mount(defineComponent({ setup: () => { out = fn(useAchievementText()); return () => null } }), { global: { plugins: [i18n] } })
  return out
}

const ach = (over: Partial<Achievement>): Achievement =>
  ({ id: 1, name: "Server Name", description: "", action_type: "watched_count", target_count: 10, exp_reward: 0, ...over }) as Achievement

describe("achievement names", () => {
  it("has an en and th name for every seeded achievement", () => {
    for (const a of seed as { action_type: string; target_count: number }[]) {
      const key = nameKey(a).replace("achievements.names.", "")
      expect(en.names, key).toHaveProperty(key)
      expect(th.names, key).toHaveProperty(key)
    }
  })

  it("translates by action type and target, and falls back to the server name", () => {
    i18n.global.locale.value = "th"
    const { name } = withText((x) => x)
    expect(name(ach({}))).toBe(th.names.watched_count_10)
    expect(name(ach({ action_type: "unknown_type", target_count: 3, name: "Fallback" }))).toBe("Fallback")
  })

  it("searches and sorts by the translated name", () => {
    i18n.global.locale.value = "th"
    const ua = (a: Achievement) => ({ achievement_id: a.id, achievement: a, is_unlocked: false, current_progress: 0 }) as unknown as UserAchievement
    const unit = (u: UserAchievement) => ({ key: String(u.achievement_id), category: "watching", tiers: [u], current: u, status: "locked", pct: 0, lastUnlockedAt: 0 }) as unknown as AchievementUnit
    const a = ua(ach({ id: 1, action_type: "watched_count", target_count: 1 }))
    const b = ua(ach({ id: 2, action_type: "review_count", target_count: 1 }))
    const { name } = withText((x) => x)
    const byName = (x: UserAchievement) => name(x.achievement)
    const hits = filterUnits([unit(a), unit(b)], { status: "all", category: "all", search: th.names.review_count_1, name: byName })
    expect(hits.map((h) => h.key)).toEqual(["2"])
    const sorted = sortUnits([unit(a), unit(b)], "name", "th", byName).map((u) => u.key)
    const expected = [{ key: "1", n: th.names.watched_count_1 }, { key: "2", n: th.names.review_count_1 }]
      .sort((x, y) => x.n.localeCompare(y.n, "th"))
      .map((x) => x.key)
    expect(sorted).toEqual(expected)
  })
})
