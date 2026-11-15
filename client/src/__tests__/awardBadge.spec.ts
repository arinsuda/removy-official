import { afterEach, beforeEach, describe, expect, it } from "vitest"
import { mount, type VueWrapper } from "@vue/test-utils"
import AwardBadge from "@/components/events/AwardBadge.vue"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()
const wrappers: VueWrapper[] = []

function badge(period: string, kind: "monthly" | "yearly" = "monthly", size: "sm" | "md" = "md") {
  const w = mount(AwardBadge, { props: { award: { kind, period }, size }, global: { plugins: [i18n] } })
  wrappers.push(w)
  return w
}

beforeEach(() => {
  i18n.global.locale.value = "en"
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
})

describe("AwardBadge", () => {
  it("shows the short start month and a full label for a monthly award", () => {
    const w = badge("2026-11")
    expect(w.text()).toContain("Nov 26")
    expect(w.attributes("aria-label")).toBe("POPCORN award, November 2026")
    expect(w.attributes("title")).toBe("POPCORN award, November 2026")
  })

  it("shows only the year for a yearly award, with a yearly label", () => {
    const w = badge("2026", "yearly")
    expect(w.text()).toContain("2026")
    expect(w.attributes("aria-label")).toBe("POPCORN yearly award, 2026")
  })

  it("renders the compact variant for poster corners", () => {
    expect(badge("2026-11", "monthly", "sm").classes()).toContain("award-badge--sm")
    expect(badge("2026-11").classes()).toContain("award-badge--md")
  })

  it("labels the badge in Thai when the locale is th", () => {
    i18n.global.locale.value = "th"
    const w = badge("2026-11")
    expect(w.attributes("aria-label")).toContain("รางวัล POPCORN ประจำ")
    expect(w.attributes("aria-label")).toContain("2569")
  })
})
