import { describe, it, expect, beforeEach } from "vitest"
import { mount } from "@vue/test-utils"
import { i18n } from "../i18n"
import MediaRating from "../components/movie/MediaRating.vue"
import PopcornScoreBadge from "../components/movie/PopcornScoreBadge.vue"
import { withLiveStats } from "../utils/mediaRatings"
import { popcornTier } from "../components/movie/popcornTier"
import type { MediaRatings } from "../types/movie"

const src = (average: number | null, count = 0, scale = 10) => ({
  average,
  count,
  available: average !== null && average > 0,
  scale,
})

const full: MediaRatings = {
  popcorn: src(4.25, 12, 5),
  tmdb: src(7.84, 900),
  imdb: src(8.1, 1000),
  rotten_tomatoes: src(94, 0, 100),
}

const mountIt = (props: Record<string, unknown>) =>
  mount(MediaRating, { props, global: { plugins: [i18n] } })

describe("popcornTier", () => {
  it.each([
    [0, 0, "none"],
    [null, 0, "none"],
    [undefined, 5, "none"],
    [4.5, 0, "none"],
    [1, 3, "stale"],
    [2.4, 3, "stale"],
    [2.5, 3, "fresh"],
    [3.9, 3, "fresh"],
    [3.96, 3, "hot"],
    [4, 1, "hot"],
    [5, 9, "hot"],
  ])("avg %s count %s -> %s", (avg, count, tier) => {
    expect(popcornTier(avg as number | null | undefined, count)).toBe(tier)
  })
})

describe("PopcornScoreBadge", () => {
  beforeEach(() => {
    i18n.global.locale.value = "en"
  })

  it("overlays the score with 1 decimal and an aria-label", () => {
    const w = mount(PopcornScoreBadge, { props: { average: 3.46, count: 2 }, global: { plugins: [i18n] } })
    expect(w.get("[data-test=score]").text()).toBe("3.5")
    expect(w.get("[role=img]").attributes("aria-label")).toBe("POPCORN score 3.5 out of 5")
    expect(w.attributes("data-tier")).toBe("fresh")
  })

  it.each([
    [0, 0, "none", "No ratings yet"],
    [1.2, 4, "stale", "Stale"],
    [3, 4, "fresh", "Fresh Pop"],
    [4.6, 4, "hot", "Hot Pop"],
  ])("detail shows tier label (%s/%s -> %s)", (average, count, tier, label) => {
    const w = mount(PopcornScoreBadge, { props: { average, count, size: "detail" }, global: { plugins: [i18n] } })
    expect(w.attributes("data-tier")).toBe(tier)
    expect(w.get("[data-test=tier-label]").text()).toBe(label)
  })

  it("uses the no-ratings text as aria-label when the tier is none", () => {
    const w = mount(PopcornScoreBadge, { props: { average: 0, count: 0 }, global: { plugins: [i18n] } })
    expect(w.get("[role=img]").attributes("aria-label")).toBe("No ratings yet")
  })

  it("hides the tier label on the card size and localizes it in Thai", () => {
    const card = mount(PopcornScoreBadge, { props: { average: 4.5, count: 1 }, global: { plugins: [i18n] } })
    expect(card.find("[data-test=tier-label]").exists()).toBe(false)

    i18n.global.locale.value = "th"
    const detail = mount(PopcornScoreBadge, { props: { average: 0, count: 0, size: "detail" }, global: { plugins: [i18n] } })
    expect(detail.get("[data-test=tier-label]").text()).toBe("ยังไม่มีคะแนน")
  })
})

describe("MediaRating", () => {
  beforeEach(() => {
    i18n.global.locale.value = "en"
  })

  it("falls back to POPCORN 0.0 when there are no ratings or reviews", () => {
    const a = mountIt({ variant: "card" })
    expect(a.get("[data-test=score]").text()).toBe("0.0")
    expect(a.find("[data-source]").exists()).toBe(false)

    const b = mountIt({ variant: "detail", ratings: { popcorn: src(0, 0, 5), tmdb: src(null) } })
    expect(b.get("[data-test=score]").text()).toBe("0.0")
    expect(b.get("[data-test=review-count]").text()).toBe("out of 5 · 0 reviews")
  })

  it("card variant shows POPCORN plus TMDB only", () => {
    const w = mountIt({ variant: "card", ratings: full })
    expect(w.get("[data-test=score]").text()).toBe("4.3")
    expect(w.findAll("[data-source]").map((e) => e.attributes("data-source"))).toEqual(["tmdb"])
    expect(w.get("[data-source=tmdb]").text()).toContain("7.8")
    expect(w.find("[data-test=review-count]").exists()).toBe(false)
  })

  it("detail variant shows 'out of 5' count, TMDB, IMDb and RT with scale aria-labels", () => {
    const w = mountIt({ variant: "detail", ratings: full })
    expect(w.find("[data-test=popcorn-value]").exists()).toBe(false)
    expect(w.get("[data-test=review-count]").text()).toBe("out of 5 · 12 reviews")
    expect(w.findAll("[data-source]").map((e) => e.attributes("data-source"))).toEqual(["tmdb", "imdb", "rt"])
    expect(w.get("[data-source=tmdb]").attributes("aria-label")).toBe("TMDB score 7.8 out of 10")
    expect(w.get("[data-source=imdb]").attributes("aria-label")).toBe("IMDb score 8.1 out of 10")
    expect(w.get("[data-source=rt]").attributes("aria-label")).toBe("Rotten Tomatoes score 94 percent")
    expect(w.get("[data-source=rt]").text()).toContain("94%")
  })

  it("hides unavailable sources", () => {
    const ratings: MediaRatings = {
      ...full,
      imdb: { average: null, count: 0, available: false, scale: 10 },
      rotten_tomatoes: undefined,
      tmdb: src(null),
    }
    const w = mountIt({ variant: "detail", ratings })
    expect(w.find("[data-source]").exists()).toBe(false)
  })

  it("uses legacy voteAverage for TMDB when ratings.tmdb is absent", () => {
    const w = mountIt({ variant: "card", voteAverage: 6.66 })
    expect(w.get("[data-source=tmdb]").text()).toContain("6.7")
    expect(w.get("[data-test=score]").text()).toBe("0.0")
  })
})

describe("withLiveStats", () => {
  it("resets popcorn to 0.0 when live stats say there is no rating", () => {
    const out = withLiveStats(full, { average_rating: 0, review_count: 0, has_rating: false })
    expect(out?.popcorn).toEqual({ average: 0, count: 0, available: false, scale: 5 })
    expect(out?.tmdb).toEqual(full.tmdb)
  })

  it("uses live numbers when a rating exists and passes null through", () => {
    expect(withLiveStats(full, { average_rating: 3, review_count: 2, has_rating: true })?.popcorn.average).toBe(3)
    expect(withLiveStats(null, { average_rating: 3, review_count: 2, has_rating: true })).toBeNull()
  })
})
