import { describe, expect, it } from "vitest"
import { ENTRY_SORTS, summaryToHomeMedia, type ProfileEntry } from "@/components/profile/tabs/profileMedia"
import { REVIEW_FILTERS, summarizeReviews } from "@/components/profile/tabs/reviewList"
import { dedupeWatched, groupByMonth, inYear, yearsOf, type WatchedEntry } from "@/components/profile/tabs/watched"
import { countUnlocked, filterAchievements, progressOf, sortAchievements } from "@/components/achievement/achievements"
import type { ReviewResponse } from "@/types"
import type { UserAchievement } from "@/types/achievement"

const entry = (id: number, title: string, at: string, type: "movie" | "tv" = "movie"): WatchedEntry => ({
  id,
  at,
  media: summaryToHomeMedia({ id, title, media_type: type, poster_url: "/p.jpg", vote_average: 7 }),
})

describe("watched grouping", () => {
  const list = [entry(1, "A", "2025-03-20T12:00:00"), entry(2, "B", "2025-03-05T12:00:00"), entry(3, "C", "2024-12-15T12:00:00")]

  it("groups consecutive entries by month in the given order", () => {
    const groups = groupByMonth(list, "en")
    expect(groups.map((g) => [g.key, g.items.length])).toEqual([["2025-03", 2], ["2024-12", 1]])
    expect(groups[0]?.label).toBe("March 2025")
  })

  it("lists years newest first and filters by year", () => {
    expect(yearsOf(list)).toEqual([2025, 2024])
    expect(inYear(list, 2024).map((e) => e.id)).toEqual([3])
  })

  it("keeps the most recent watch of a duplicated title", () => {
    const out = dedupeWatched([{ ...entry(1, "A", "2024-01-10T12:00:00") }, { ...entry(1, "A", "2025-01-10T12:00:00"), id: 9 }])
    expect(out).toHaveLength(1)
    expect(out[0]?.id).toBe(9)
  })
})

describe("entry sorts", () => {
  it("orders by date and title", () => {
    const rows: ProfileEntry[] = [entry(1, "Zed", "2025-01-01"), entry(2, "Alpha", "2025-02-01")]
    expect([...rows].sort(ENTRY_SORTS.newest).map((r) => r.id)).toEqual([2, 1])
    expect([...rows].sort(ENTRY_SORTS.oldest).map((r) => r.id)).toEqual([1, 2])
    expect([...rows].sort(ENTRY_SORTS.title).map((r) => r.id)).toEqual([2, 1])
  })
})

const review = (id: number, rating: number, body: string, genres: string[], created = "2025-01-01T00:00:00Z"): ReviewResponse =>
  ({ id, rating, body, created_at: created, updated_at: created, media: { id, genres: genres.map((name, i) => ({ id: i, name })) } }) as unknown as ReviewResponse

describe("review list helpers", () => {
  const rs = [review(1, 5, "great", ["Drama", "Sci-Fi"]), review(2, 2, "", ["Drama"]), review(3, 4, "ok", ["Comedy"], "2025-02-01T00:00:00Z")]

  it("filters high, critical and with-text", () => {
    expect(rs.filter(REVIEW_FILTERS.high!).map((r) => r.id)).toEqual([1, 3])
    expect(rs.filter(REVIEW_FILTERS.low!).map((r) => r.id)).toEqual([2])
    expect(rs.filter(REVIEW_FILTERS.withText!).map((r) => r.id)).toEqual([1, 3])
  })

  it("summarizes count, average and top genre", () => {
    expect(summarizeReviews(rs)).toEqual({ count: 3, average: 3.7, topGenre: "Drama" })
    expect(summarizeReviews([])).toEqual({ count: 0, average: 0, topGenre: null })
  })
})

const ach = (id: number, unlocked: boolean, current: number, target: number, at: string | null, pct = 0): UserAchievement => ({
  achievement_id: id,
  achievement: { id, name: `A${id}`, description: "", exp_reward: 10, action_type: "x", target_count: target },
  current_count: current,
  is_unlocked: unlocked,
  unlocked_at: at,
  progress_pct: pct,
})

describe("achievement helpers", () => {
  const list = [ach(1, false, 1, 10, null), ach(2, true, 5, 5, "2025-01-01"), ach(3, false, 9, 10, null), ach(4, true, 1, 1, "2025-03-01")]

  it("counts and filters unlocked", () => {
    expect(countUnlocked(list)).toBe(2)
    expect(filterAchievements(list, "unlocked")).toHaveLength(2)
    expect(filterAchievements(list, "locked")).toHaveLength(2)
    expect(filterAchievements(list, "all")).toHaveLength(4)
  })

  it("sorts unlocked first (newest), then closest to done", () => {
    expect(sortAchievements(list).map((a) => a.achievement_id)).toEqual([4, 2, 3, 1])
  })

  it("derives and clamps progress", () => {
    expect(progressOf(ach(1, false, 3, 10, null))).toBe(30)
    expect(progressOf(ach(1, false, 3, 10, null, 250))).toBe(100)
    expect(progressOf(ach(1, false, 0, 0, null))).toBe(0)
  })
})
