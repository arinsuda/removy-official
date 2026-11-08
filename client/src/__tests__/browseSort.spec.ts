import { describe, expect, it } from "vitest"
import { filterTier, sortMedia } from "@/components/browse/browseSort"
import { fromMovie, type HomeMedia } from "@/components/home/homeMedia"
import type { Movie } from "@/types"

function m(id: number, o: { avg?: number; count?: number; tmdb?: number; date?: string } = {}): HomeMedia {
  return fromMovie({
    id,
    title: `m${id}`,
    poster_path: "/p.jpg",
    release_date: o.date ?? "",
    vote_average: o.tmdb ?? 0,
    ratings: { popcorn: { average: o.avg ?? 0, count: o.count ?? 0, available: !!o.count, scale: 5 } },
  } as unknown as Movie)
}
const ids = (xs: HomeMedia[]) => xs.map((x) => x.id)

describe("sortMedia", () => {
  it("keeps list order for default", () => {
    const xs = [m(1), m(2)]
    expect(sortMedia(xs, "default")).toBe(xs)
  })

  it("popcorn: score desc, count tiebreak, unrated last in list order", () => {
    const xs = [m(1), m(2, { avg: 4, count: 3 }), m(3, { avg: 4.5, count: 1 }), m(4, { avg: 4, count: 10 }), m(5)]
    expect(ids(sortMedia(xs, "popcorn"))).toEqual([3, 4, 2, 1, 5])
  })

  it("tmdb sorts by vote average", () => {
    expect(ids(sortMedia([m(1, { tmdb: 6 }), m(2, { tmdb: 8 })], "tmdb"))).toEqual([2, 1])
  })

  it("date sorts newest first with missing dates last", () => {
    const xs = [m(1), m(2, { date: "2020-01-01" }), m(3, { date: "2024-05-05" })]
    expect(ids(sortMedia(xs, "date"))).toEqual([3, 2, 1])
  })

  it("does not mutate the input", () => {
    const xs = [m(1, { tmdb: 1 }), m(2, { tmdb: 9 })]
    sortMedia(xs, "tmdb")
    expect(ids(xs)).toEqual([1, 2])
  })
})

describe("filterTier", () => {
  const xs = [m(1), m(2, { avg: 4.5, count: 2 }), m(3, { avg: 3, count: 2 }), m(4, { avg: 1, count: 2 })]
  it("all returns everything", () => expect(filterTier(xs, "all")).toBe(xs))
  it("filters by tier", () => {
    expect(ids(filterTier(xs, "hot"))).toEqual([2])
    expect(ids(filterTier(xs, "fresh"))).toEqual([3])
    expect(ids(filterTier(xs, "stale"))).toEqual([4])
    expect(ids(filterTier(xs, "none"))).toEqual([1])
  })
})
