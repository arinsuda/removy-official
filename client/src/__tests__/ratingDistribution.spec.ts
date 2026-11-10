import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import RatingDistribution from "@/components/detail/RatingDistribution.vue"
import { ratingDistribution, ratingDistributionFromBuckets } from "@/components/detail/ratingDistribution"
import { i18n } from "@/i18n"
import type { Visibility } from "@/types"

const r = (rating: number, visibility: Visibility = "public") => ({ rating, visibility })

describe("ratingDistribution", () => {
  it("counts public reviews only", () => {
    const res = ratingDistribution([r(5), r(5), r(4.5), r(5, "private"), r(1, "followers")])
    expect(res.total).toBe(3)
    expect(res.buckets).toHaveLength(10)
    expect(res.buckets.find(b => b.value === 5)).toMatchObject({ count: 2, pct: 100 })
    expect(res.buckets.find(b => b.value === 4.5)).toMatchObject({ count: 1, pct: 50 })
    expect(res.buckets.find(b => b.value === 1)?.count).toBe(0)
  })

  it("ignores out-of-range ratings and handles an empty list", () => {
    expect(ratingDistribution([r(0), r(7)]).total).toBe(0)
    expect(ratingDistribution([]).buckets.every(b => b.pct === 0)).toBe(true)
  })

  it("builds 10 steps from server distribution buckets", () => {
    const res = ratingDistributionFromBuckets([
      { rating: 0.5, count: 0 },
      { rating: 3.5, count: 2 },
      { rating: 5.0, count: 4 },
    ])
    expect(res.total).toBe(6)
    expect(res.buckets).toHaveLength(10)
    expect(res.buckets.find(b => b.value === 5)).toMatchObject({ count: 4, pct: 100 })
    expect(res.buckets.find(b => b.value === 3.5)).toMatchObject({ count: 2, pct: 50 })
    expect(res.buckets.find(b => b.value === 0.5)).toMatchObject({ count: 0, pct: 0 })
  })
})

describe("RatingDistribution", () => {
  const mountDist = (reviews: ReturnType<typeof r>[]) => {
    i18n.global.locale.value = "en"
    return mount(RatingDistribution, { props: { reviews }, global: { plugins: [i18n] } })
  }

  it("says how many public reviews it is based on", () => {
    const w = mountDist([r(4), r(3), r(5, "private")])
    expect(w.findAll(".dist__row")).toHaveLength(10)
    expect(w.get('[data-test="dist-based"]').text()).toBe("Based on 2 public reviews")
  })

  it("renders server distribution buckets directly when provided", () => {
    i18n.global.locale.value = "en"
    const w = mount(RatingDistribution, {
      props: {
        distribution: [
          { rating: 4.0, count: 3 },
          { rating: 5.0, count: 7 },
        ],
      },
      global: { plugins: [i18n] },
    })
    expect(w.findAll(".dist__row")).toHaveLength(10)
    expect(w.get('[data-test="dist-based"]').text()).toBe("Based on 10 public reviews")
  })

  it("shows an empty message when only non-public reviews exist", () => {
    const w = mountDist([r(5, "private")])
    expect(w.find(".dist__rows").exists()).toBe(false)
    expect(w.text()).toContain("No public reviews yet")
  })
})
