import { describe, expect, it } from "vitest"
import { formatReleaseDate, releaseStatus } from "@/utils/releaseStatus"
import { groupByReleaseYear, UNKNOWN_YEAR } from "@/components/bmol/bmolTimeline"
import { ENTRY_SORTS, summaryToHomeMedia, type ProfileEntry } from "@/components/profile/tabs/profileMedia"
import type { BMOLItemResponse } from "@/types"

const today = new Date(2026, 5, 15, 23, 59)

describe("releaseStatus", () => {
  it("compares calendar days in local time", () => {
    expect(releaseStatus("2026-06-15", today)).toBe("released")
    expect(releaseStatus("2026-06-16", today)).toBe("upcoming")
    expect(releaseStatus("2020-01-01", today)).toBe("released")
  })
  it("treats missing or malformed dates as unknown", () => {
    for (const d of [undefined, null, "", "soon", "2026"]) expect(releaseStatus(d, today)).toBe("unknown")
  })
  it("formats without shifting the day", () => {
    expect(formatReleaseDate("2026-01-01", "en")).toContain("1")
    expect(formatReleaseDate("", "en")).toBe("")
  })
})

const bm = (id: number, rank: number, release_date?: string): BMOLItemResponse =>
  ({ id, rank, note: "", media_type: "movie", created_at: "", media: { id, title: `t${id}`, release_date } }) as unknown as BMOLItemResponse

describe("groupByReleaseYear", () => {
  it("groups newest year first, unknown last, rank order inside a year", () => {
    const g = groupByReleaseYear([bm(1, 3, "2019-05-01"), bm(2, 1, "2024-01-01"), bm(3, 2, undefined), bm(4, 1, "2019-02-01"), bm(5, 4, "")])
    expect(g.map((x) => x.year)).toEqual([2024, 2019, UNKNOWN_YEAR])
    expect(g[1]!.items.map((i) => i.id)).toEqual([4, 1])
    expect(g[2]!.items.map((i) => i.id)).toEqual([3, 5])
  })
  it("can run oldest first", () => {
    expect(groupByReleaseYear([bm(1, 1, "2024-01-01"), bm(2, 1, "2019-01-01"), bm(3, 1)], "asc").map((x) => x.year)).toEqual([2019, 2024, UNKNOWN_YEAR])
  })
})

describe("release sort", () => {
  const entry = (id: number, release?: string): ProfileEntry => ({
    id, at: "2026-01-01", media: summaryToHomeMedia({ id, title: `t${id}`, media_type: "movie", poster_url: "", vote_average: 0, release_date: release }),
  })
  it("orders newest release first with unknown dates last", () => {
    const sorted = [entry(1, "2020-01-01"), entry(2), entry(3, "2027-03-01")].sort(ENTRY_SORTS.release)
    expect(sorted.map((e) => e.id)).toEqual([3, 1, 2])
  })
})
