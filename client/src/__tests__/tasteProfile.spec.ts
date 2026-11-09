import { describe, expect, it } from "vitest"
import { tasteSummary } from "@/components/profile/tasteProfile"
import { ratingDistribution } from "@/components/detail/ratingDistribution"

const review = (rating: number, genres: [number, string][] = [], visibility = "public") =>
  ({ rating, visibility, media: { genres: genres.map(([id, name]) => ({ id, name })) } }) as never

describe("tasteSummary", () => {
  it("is 'new' with no average for an empty list", () => {
    const s = tasteSummary([])
    expect(s.average).toBeNull()
    expect(s.persona).toBe("new")
    expect(s.topGenres).toEqual([])
    expect(s.distribution.total).toBe(0)
  })

  it("stays 'new' below three reviews", () => {
    expect(tasteSummary([review(5), review(5)]).persona).toBe("new")
  })

  it.each([
    [[5, 4.5, 4], "generous"],
    [[3, 3.5, 4.5], "balanced"],
    [[1, 2, 2.5], "tough"],
  ])("labels %j as %s", (ratings, persona) => {
    expect(tasteSummary((ratings as number[]).map((r) => review(r))).persona).toBe(persona)
  })

  it("returns the top 3 genres ordered by count then name", () => {
    const g = (...ids: [number, string][]) => review(4, ids)
    const s = tasteSummary([
      g([1, "Drama"], [2, "Comedy"]),
      g([1, "Drama"], [3, "Action"]),
      g([2, "Comedy"], [4, "Horror"]),
      g([5, "Anime"]),
    ])
    expect(s.topGenres.map((x) => x.name)).toEqual(["Comedy", "Drama", "Action"])
  })

  it("counts private reviews in the distribution (includeAll)", () => {
    const s = tasteSummary([review(4, [], "private"), review(4)])
    expect(s.distribution.total).toBe(2)
    expect(ratingDistribution([review(4, [], "private")] as never).total).toBe(0)
  })
})
