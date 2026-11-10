import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import StarRatingInput from "@/components/movie/StarRatingInput.vue"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()
i18n.global.locale.value = "en"

const build = (modelValue: number, readonly = false) =>
  mount(StarRatingInput, { props: { modelValue, readonly }, global: { plugins: [i18n] } })

describe("StarRatingInput", () => {
  it("exposes slider semantics with aria-valuenow", () => {
    const el = build(3.5).find('[role="slider"]')
    expect(el.attributes("aria-valuenow")).toBe("3.5")
    expect(el.attributes("aria-valuemin")).toBe("0.5")
    expect(el.attributes("aria-valuemax")).toBe("5")
    expect(el.attributes("tabindex")).toBe("0")
  })

  it("steps by 0.5 with arrow keys and clamps at the bounds", async () => {
    const w = build(3)
    const el = w.find('[role="slider"]')
    await el.trigger("keydown", { key: "ArrowRight" })
    await el.trigger("keydown", { key: "Home" })
    await el.trigger("keydown", { key: "End" })
    expect(w.emitted("update:modelValue")).toEqual([[3.5], [0.5], [5]])
    const low = build(0.5)
    await low.find('[role="slider"]').trigger("keydown", { key: "ArrowLeft" })
    expect(low.emitted("update:modelValue")).toEqual([[0.5]])
  })

  it("selects half values by clicking the left half of a star", async () => {
    const w = build(0)
    await w.findAll(".half-zone.left")[2]!.trigger("click")
    await w.findAll(".half-zone.right")[2]!.trigger("click")
    expect(w.emitted("update:modelValue")).toEqual([[2.5], [3]])
  })

  it("renders half stars with a half clip and is not interactive when readonly", () => {
    const w = build(2.5, true)
    expect(w.find('[role="img"]').exists()).toBe(true)
    expect(w.find(".half-zone").exists()).toBe(false)
    const clips = w.findAll("path.fill--on").map((p) => p.attributes("clip-path"))
    expect(clips).toHaveLength(3)
    expect(clips[2]).toContain("-3-half")
  })
})

describe("StarRatingInput labels", () => {
  it("falls back to a translated label and announces the value through i18n", () => {
    i18n.global.locale.value = "en"
    const w = mount(StarRatingInput, { props: { modelValue: 3.5 }, global: { plugins: [i18n] } })
    const slider = w.get('[role="slider"]')
    expect(slider.attributes("aria-label")).toBe(i18n.global.t("reviews.composer.ratingLabel"))
    expect(slider.attributes("aria-valuetext")).toBe("3.5 out of 5")
  })
})
