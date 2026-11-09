import { describe, expect, it } from "vitest"
import { actingCredits, ageOf, creditStats, filterCredits, groupByYear, knownFor, sortCredits } from "@/components/actor/actorCredits"

const m = (id: number, date: string, popularity: number, character?: string) =>
  ({ id, title: `M${id}`, original_title: "", overview: "", poster_path: `/${id}.jpg`, backdrop_path: "", release_date: date, vote_average: 0, vote_count: 0, popularity, genre_ids: [], adult: false, character }) as never
const s = (id: number, date: string, popularity: number) =>
  ({ id, name: `S${id}`, original_name: "", overview: "", poster_path: `/${id}.jpg`, backdrop_path: "", first_air_date: date, vote_average: 0, vote_count: 0, popularity, genre_ids: [], adult: false }) as never

describe("actor credits", () => {
  it("dedupes titles and merges characters", () => {
    const credits = actingCredits([m(1, "2000-01-01", 1, "A"), m(1, "2000-01-01", 1, "B"), m(1, "2000-01-01", 1, "A")], [])
    expect(credits).toHaveLength(1)
    expect(credits[0]!.character).toBe("A / B")
  })

  it("keeps a movie and a series with the same id apart", () => {
    expect(actingCredits([m(5, "2000-01-01", 1)], [s(5, "2001-01-01", 1)])).toHaveLength(2)
  })

  it("sorts by date with unknown years last, in both directions", () => {
    const credits = actingCredits([m(1, "2010-01-01", 1), m(2, "", 99), m(3, "2020-01-01", 1)], [])
    expect(sortCredits(credits, "newest").map((c) => c.media.id)).toEqual([3, 1, 2])
    expect(sortCredits(credits, "oldest").map((c) => c.media.id)).toEqual([1, 3, 2])
    expect(sortCredits(credits, "popular").map((c) => c.media.id)).toEqual([2, 1, 3])
  })

  it("filters by type", () => {
    const credits = actingCredits([m(1, "2000-01-01", 1)], [s(2, "2001-01-01", 1)])
    expect(filterCredits(credits, "tv").map((c) => c.media.id)).toEqual([2])
    expect(filterCredits(credits, "all")).toHaveLength(2)
  })

  it("groups by year and computes stats", () => {
    const credits = sortCredits(actingCredits([m(1, "2000-05-01", 1), m(2, "2000-06-01", 2), m(3, "", 1), m(4, "2004-01-01", 1)], [s(5, "2004-01-01", 1)]), "newest")
    expect(groupByYear(credits, false).map((g) => g.year)).toEqual([2004, 2000, null])
    expect(groupByYear(credits, true).map((g) => g.year)).toEqual([2000, 2004, null])
    expect(creditStats(credits)).toEqual({ total: 5, movies: 4, tv: 1, firstYear: 2000, lastYear: 2004 })
    expect(creditStats([])).toMatchObject({ firstYear: null, lastYear: null })
  })

  it("picks known-for titles by popularity and skips posterless ones", () => {
    const noPoster = { ...(m(9, "2000-01-01", 999) as object), poster_path: "" } as never
    const picks = knownFor(actingCredits([m(1, "2000-01-01", 5), noPoster, m(2, "2000-01-01", 50)], []), 2)
    expect(picks.map((c) => c.media.id)).toEqual([2, 1])
  })

  it("computes age, birthday boundaries and death", () => {
    const now = new Date(2026, 5, 1)
    expect(ageOf("1996-06-01", null, now)).toEqual({ age: 30, deceased: false })
    expect(ageOf("1996-06-02", null, now)).toEqual({ age: 29, deceased: false })
    expect(ageOf("1976-11-29", "2020-08-28", now)).toEqual({ age: 43, deceased: true })
    expect(ageOf(null, null, now)).toBeNull()
    expect(ageOf("garbage", null, now)).toBeNull()
  })
})
