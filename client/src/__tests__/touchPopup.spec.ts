import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { mount } from "@vue/test-utils"
import { createPinia } from "pinia"

vi.mock("@/api/api", () => ({
  movieApi: { getById: vi.fn().mockResolvedValue({ data: {} }), getVideos: vi.fn().mockResolvedValue({ data: { results: [] } }), getSeriesById: vi.fn(), getTVVideos: vi.fn() },
}))
vi.mock("@/components/movie/PopupCard.vue", () => ({ default: { name: "PopupCard", render: () => h("div", { "data-test": "popup-card" }) } }))

import { i18n } from "../i18n"
import HomePosterCard from "../components/home/HomePosterCard.vue"
import { HOVER_FINE_QUERY } from "@/constants/input"
import { provideHomePopup, type HomePopup } from "@/composables/useHomePopup"
import type { HomeMedia } from "@/components/home/homeMedia"

const item = { key: "movie-1", id: 1, type: "movie", title: "Dune", source: { id: 1 }, poster_path: null } as unknown as HomeMedia

function stubInput(hover: boolean) {
  vi.stubGlobal("matchMedia", (q: string) => ({
    matches: q === HOVER_FINE_QUERY ? hover : false,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }))
}

function mountCard() {
  let popup!: HomePopup
  const Host = defineComponent({
    setup() {
      popup = provideHomePopup()
      return () => h(HomePosterCard, { item, rail: "r" })
    },
  })
  const w = mount(Host, { attachTo: document.body, global: { plugins: [i18n, createPinia()], stubs: { RouterLink: { template: "<a><slot /></a>" } } } })
  return { w, popup: () => popup }
}

describe("hover vs touch quick actions", () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
    document.body.innerHTML = ""
  })

  it("opens the hover popup on mouseenter when the device can hover", async () => {
    stubInput(true)
    const { w, popup } = mountCard()
    await w.find(".poster-wrap").trigger("mouseenter")
    await vi.advanceTimersByTimeAsync(300)
    expect(popup().openKey.value).toBe("r:movie-1")
  })

  it("does not open the hover popup on a coarse pointer", async () => {
    stubInput(false)
    const { w, popup } = mountCard()
    await w.find(".poster-wrap").trigger("mouseenter")
    await vi.advanceTimersByTimeAsync(300)
    expect(popup().openKey.value).toBeNull()
  })

  it.each([true, false])("the more button opens the sheet by click or keyboard (hover=%s)", async (hover) => {
    stubInput(hover)
    const { w } = mountCard()
    const btn = w.find(".more-btn")
    expect(btn.attributes("aria-label")).toBe(i18n.global.t("media.popup.more"))
    expect(btn.element.tagName).toBe("BUTTON") // native button: Enter/Space fire click
    await btn.trigger("click")
    expect(document.body.querySelector("[role=dialog] [data-test=popup-card]")).not.toBeNull()
  })
})
