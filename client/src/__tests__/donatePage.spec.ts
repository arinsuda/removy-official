import { beforeEach, describe, expect, it, vi } from "vitest"
import { mount } from "@vue/test-utils"
import { createPinia } from "pinia"
import DonateView from "@/views/aboutus/DonateView.vue"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()
beforeEach(() => (i18n.global.locale.value = "en"))

describe("DonateView", () => {
  it("renders hero, tiers, payment and faq", async () => {
    const w = mount(DonateView, { global: { plugins: [createPinia(), i18n] } })
    expect(w.get("h1").text()).toBe("Support POPCORN Development")
    expect(w.findAll(".tier")).toHaveLength(3)
    expect(w.find('img[src="/mePromptPay.png"]').exists()).toBe(true)
    expect(w.text()).toContain("Questions")
  })

  it("shows the suggested amount and the bank account after switching tab", async () => {
    const w = mount(DonateView, { global: { plugins: [createPinia(), i18n] } })
    await w.findAll(".tier")[1]!.trigger("click")
    expect(w.text()).toContain("Suggested amount: ฿100")
    await w.findAll('[role="tab"]')[1]!.trigger("click")
    expect(w.text()).toContain("045-3-31543-6")
  })

  it("copies the account number", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true })
    const w = mount(DonateView, { global: { plugins: [createPinia(), i18n] } })
    await w.findAll('[role="tab"]')[1]!.trigger("click")
    await w.findAll("button").find((b) => b.text() === "Copy")!.trigger("click")
    expect(writeText).toHaveBeenCalledWith("0453315436")
  })

  it("expands a faq answer", async () => {
    const w = mount(DonateView, { global: { plugins: [createPinia(), i18n] } })
    const btn = w.get(".faq__btn")
    expect(btn.attributes("aria-expanded")).toBe("false")
    await btn.trigger("click")
    expect(btn.attributes("aria-expanded")).toBe("true")
  })
})
