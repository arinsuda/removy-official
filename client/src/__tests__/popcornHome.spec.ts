import { describe, expect, it } from "vitest"
import { fromMovie, fromSeries, type HomeMedia } from "@/components/home/homeMedia"
import { countByTier, dedupeMedia, filterByTier, interleave, mediaTier, pickHeroItems, tierWithBackfill } from "@/components/home/popcornHome"
import type { Movie, TVSeries } from "@/types"

const popcorn = (average: number, count: number) => ({
  popcorn: { average, count, available: count > 0, scale: 5 },
})

function movie(id: number, average = 0, count = 0, backdrop = "/b.jpg"): HomeMedia {
  return fromMovie({
    id,
    title: `m${id}`,
    poster_path: "/p.jpg",
    backdrop_path: backdrop,
    ratings: popcorn(average, count),
  } as unknown as Movie)
}

function series(id: number, average = 0, count = 0): HomeMedia {
  return fromSeries({
    id,
    name: `s${id}`,
    poster_path: "/p.jpg",
    backdrop_path: "/b.jpg",
    ratings: popcorn(average, count),
  } as unknown as TVSeries)
}

describe("tier filtering", () => {
  const pool = [movie(1, 4.5, 3), movie(2, 3, 2), movie(3, 1.5, 4), movie(4), series(5, 4.0, 9), series(6, 4.8, 1)]

  it("classifies by popcorn average and review count", () => {
    expect(pool.map(mediaTier)).toEqual(["hot", "fresh", "stale", "none", "hot", "hot"])
  })

  it("treats a title without ratings as none", () => {
    expect(mediaTier(fromMovie({ id: 9, title: "x" } as unknown as Movie))).toBe("none")
  })

  it("filters a tier and sorts hot by score then review count", () => {
    expect(filterByTier(pool, "hot").map((m) => m.key)).toEqual(["tv-6", "movie-1", "tv-5"])
    expect(filterByTier(pool, "none").map((m) => m.key)).toEqual(["movie-4"])
  })

  it("respects the limit and counts tiers", () => {
    expect(filterByTier(pool, "hot", 2)).toHaveLength(2)
    expect(countByTier(pool)).toEqual({ none: 1, stale: 1, fresh: 1, hot: 3 })
  })

  it("keeps movie and tv with the same id apart, and dedupes repeats", () => {
    expect(dedupeMedia([movie(1), series(1), movie(1)]).map((m) => m.key)).toEqual(["movie-1", "tv-1"])
  })

  it("interleaves lists", () => {
    expect(interleave([movie(1), movie(2)], [series(3)]).map((m) => m.key)).toEqual(["movie-1", "tv-3", "movie-2"])
  })
})

describe("pickHeroItems", () => {
  const fallback = [movie(10), movie(11), movie(12), movie(13), movie(14), movie(15)]

  it("uses hot titles when at least three qualify", () => {
    const pool = [movie(1, 4.2, 2), series(2, 4.6, 5), movie(3, 4.0, 1), movie(4, 3, 3)]
    const pick = pickHeroItems(pool, fallback)
    expect(pick.source).toBe("hot")
    expect(pick.items.map((m) => m.key)).toEqual(["tv-2", "movie-1", "movie-3"])
  })

  it("caps at five and dedupes the pool", () => {
    const pool = Array.from({ length: 8 }, (_, i) => movie(i + 1, 4.5, 2))
    const pick = pickHeroItems([...pool, ...pool], fallback)
    expect(pick.items).toHaveLength(5)
  })

  it("falls back to TMDB popular below three hot titles", () => {
    const pick = pickHeroItems([movie(1, 4.5, 2), movie(2, 4.5, 2), movie(3, 3, 3)], fallback)
    expect(pick.source).toBe("popular")
    expect(pick.items.map((m) => m.id)).toEqual([10, 11, 12, 13, 14])
  })

  it("ignores hot titles that have no backdrop", () => {
    const pool = [movie(1, 4.5, 2, ""), movie(2, 4.5, 2), movie(3, 4.5, 2)]
    expect(pickHeroItems(pool, fallback).source).toBe("popular")
  })
})

describe("tierWithBackfill", () => {
  it("leaves a full tier untouched", () => {
    const items = Array.from({ length: 6 }, (_, i) => movie(i + 1, 4.5, 3))
    const r = tierWithBackfill(items, "hot")
    expect(r.items).toHaveLength(6)
    expect(r.backfillAt).toBe(-1)
  })

  it("fills a sparse tier with fresh then other titles after the divider", () => {
    const items = [movie(1, 4.5, 3), movie(2, 1.5, 4), movie(3, 3, 2), movie(4), movie(5, 3.5, 2), movie(6, 3.2, 2), movie(7), movie(8)]
    const r = tierWithBackfill(items, "hot")
    expect(r.items).toHaveLength(6)
    expect(r.backfillAt).toBe(1)
    expect(r.items.map((m) => m.key).slice(0, 4)).toEqual(["movie-1", "movie-5", "movie-6", "movie-3"])
    expect(new Set(r.items.map((m) => m.key)).size).toBe(6)
  })

  it("does not set a divider when there is nothing to add", () => {
    const r = tierWithBackfill([movie(1, 4.5, 3)], "hot")
    expect(r).toMatchObject({ backfillAt: -1 })
    expect(r.items).toHaveLength(1)
  })
})
