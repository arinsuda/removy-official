import { describe, expect, it } from "vitest"
import { fromMovie, type HomeMedia } from "@/components/home/homeMedia"
import { daysUntil, groupByRelease, parseLocalDate, pickHero } from "@/components/upcoming/upcomingGroups"
import type { Movie } from "@/types"

const movie = (id: number, date: string, extra: Partial<Movie> = {}): HomeMedia =>
  fromMovie({ id, title: `m${id}`, release_date: date, backdrop_path: "/b", popularity: 1, ...extra } as Movie)

// Sunday 8 Nov 2026: this week is just today, next week starts Monday the 9th.
const today = new Date(2026, 10, 8)
const ids = (g: { items: HomeMedia[] }) => g.items.map((m) => m.id)

describe("groupByRelease", () => {
  it("buckets out-now, this week, next week (Monday start) and months", () => {
    const groups = groupByRelease(
      [movie(1, "2026-11-07"), movie(2, "2026-11-08"), movie(3, "2026-11-09"), movie(4, "2026-11-15"), movie(5, "2026-11-16"), movie(6, "2026-12-25")],
      today,
      "en-US",
    )
    expect(groups.map((g) => [g.id, ids(g)])).toEqual([
      ["out", [1]],
      ["week", [2]],
      ["next", [3, 4]],
      ["m-2026-11", [5]],
      ["m-2026-12", [6]],
    ])
    expect(groups[3].monthLabel).toBe("November 2026")
  })

  it("starts weeks on Monday for a midweek today", () => {
    const wed = new Date(2026, 10, 11)
    const groups = groupByRelease([movie(1, "2026-11-15"), movie(2, "2026-11-16"), movie(3, "2026-11-22"), movie(4, "2026-11-23")], wed, "en-US")
    expect(groups.map((g) => [g.id, ids(g)])).toEqual([
      ["week", [1]],
      ["next", [2, 3]],
      ["m-2026-11", [4]],
    ])
  })

  it("sorts by date, omits empty groups and drops invalid dates", () => {
    const groups = groupByRelease([movie(1, "2027-02-01"), movie(2, ""), movie(3, "2026-12-01"), movie(4, "nope")], today, "en-US")
    expect(groups.map((g) => g.id)).toEqual(["m-2026-12", "m-2027-02"])
  })

  it("formats month names in the requested locale", () => {
    const [g] = groupByRelease([movie(1, "2026-12-25")], today, "th-TH")
    expect(g.monthLabel).toContain("ธันวาคม")
  })
})

describe("date helpers", () => {
  it("parses as a local calendar day", () => {
    const d = parseLocalDate("2026-11-08")!
    expect([d.getFullYear(), d.getMonth(), d.getDate()]).toEqual([2026, 10, 8])
    expect(parseLocalDate("")).toBeNull()
  })
  it("counts whole days", () => {
    expect(daysUntil(new Date(2026, 10, 8), today)).toBe(0)
    expect(daysUntil(new Date(2026, 10, 11), today)).toBe(3)
    expect(daysUntil(new Date(2026, 10, 7), today)).toBe(-1)
  })
})

describe("pickHero", () => {
  it("picks the most popular title among the soonest releases with a backdrop", () => {
    const hero = pickHero(
      [
        movie(1, "2026-11-07", { popularity: 999 }),
        movie(2, "2026-11-10", { popularity: 10 }),
        movie(3, "2026-11-12", { popularity: 50 }),
        movie(4, "2026-11-13", { popularity: 500, backdrop_path: "" }),
        movie(5, "2027-03-01", { popularity: 900 }),
      ],
      today,
    )
    expect(hero?.id).toBe(3)
  })
  it("returns null when nothing qualifies", () => {
    expect(pickHero([movie(1, "2026-11-01")], today)).toBeNull()
  })
})
