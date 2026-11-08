import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import PopcornMeter from "@/components/detail/PopcornMeter.vue"
import { kernelFills } from "@/components/detail/popcornMeter"
import { i18n } from "@/i18n"

describe("kernelFills", () => {
  it("fills half a point per kernel", () => {
    expect(kernelFills(5)).toEqual(Array(10).fill(1))
    expect(kernelFills(0)).toEqual(Array(10).fill(0))
    const f = kernelFills(3.7)
    expect(f.slice(0, 7)).toEqual(Array(7).fill(1))
    expect(f[7]).toBeCloseTo(0.4)
    expect(f.slice(8)).toEqual([0, 0])
  })

  it("clamps bad input", () => {
    expect(kernelFills(9)).toEqual(Array(10).fill(1))
    expect(kernelFills(undefined)).toEqual(Array(10).fill(0))
    expect(kernelFills(Number.NaN)).toEqual(Array(10).fill(0))
  })
})

describe("PopcornMeter", () => {
  const mountMeter = (average: number, count: number) => {
    i18n.global.locale.value = "en"
    return mount(PopcornMeter, { props: { average, count }, global: { plugins: [i18n] } })
  }

  it("shows the empty state when there are no reviews", () => {
    const w = mountMeter(0, 0)
    expect(w.attributes("data-tier")).toBe("none")
    expect(w.get('[data-test="meter-score"]').text()).toBe("—")
    expect(w.get('[data-test="meter-count"]').text()).toContain("No reviews yet")
  })

  it.each([
    [1.5, "stale"],
    [3.2, "fresh"],
    [4.6, "hot"],
  ])("colors %s as the %s tier", (avg, tier) => {
    const w = mountMeter(avg, 4)
    expect(w.attributes("data-tier")).toBe(tier)
    expect(w.get('[data-test="meter-score"]').text()).toBe(avg.toFixed(1))
    expect(w.findAll(".meter__kernel")).toHaveLength(10)
  })

  it("exposes the like toggle state", async () => {
    const w = mount(PopcornMeter, { props: { average: 4, count: 2, liked: true }, global: { plugins: [i18n] } })
    const btn = w.get(".meter__like")
    expect(btn.attributes("aria-pressed")).toBe("true")
    await btn.trigger("click")
    expect(w.emitted("toggle-like")).toHaveLength(1)
  })
})
