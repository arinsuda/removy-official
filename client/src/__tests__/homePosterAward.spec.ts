import { afterEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { mount, type VueWrapper } from "@vue/test-utils"
import { createPinia } from "pinia"
import type { Movie } from "@/types/movie"
import type { MediaAward } from "@/types/event"
import { fromMovie } from "@/components/home/homeMedia"
import { provideHomePopup } from "@/composables/useHomePopup"
import { i18n } from "../i18n"
import HomePosterCard from "../components/home/HomePosterCard.vue"

vi.mock("@/api/api", () => ({
  movieApi: { getById: vi.fn().mockResolvedValue({ data: {} }), getVideos: vi.fn().mockResolvedValue({ data: { results: [] } }), getSeriesById: vi.fn(), getTVVideos: vi.fn() },
}))
vi.mock("@/components/movie/PopupCard.vue", () => ({ default: { name: "PopupCard", render: () => h("div") } }))

const movie = (awards?: MediaAward[]) =>
  fromMovie({
    id: 1,
    title: "Dune",
    original_title: "Dune",
    overview: "",
    poster_path: "",
    backdrop_path: "",
    release_date: "2021-09-15",
    vote_average: 0,
    vote_count: 0,
    popularity: 0,
    genre_ids: [],
    adult: false,
    awards,
  } satisfies Movie)

const wrappers: VueWrapper[] = []

function mountCard(awards?: MediaAward[]) {
  const Host = defineComponent({
    setup() {
      provideHomePopup()
      return () => h(HomePosterCard, { item: movie(awards), rail: "r" })
    },
  })
  const w = mount(Host, {
    attachTo: document.body,
    global: { plugins: [i18n, createPinia()], stubs: { RouterLink: { template: "<a><slot /></a>" } } },
  })
  wrappers.push(w)
  return w
}

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  i18n.global.locale.value = "en"
})

describe("HomePosterCard award badge", () => {
  it("shows no badge for a title without awards", () => {
    expect(mountCard().find(".award-badge").exists()).toBe(false)
  })

  it("shows one compact monthly badge", () => {
    const badge = mountCard([{ kind: "monthly", period: "2026-11" }]).find(".award-badge")
    expect(badge.exists()).toBe(true)
    expect(badge.classes()).toContain("award-badge--sm")
    expect(badge.attributes("aria-label")).toBe("POPCORN award, November 2026")
  })

  it("prefers the yearly award and renders only one badge", () => {
    const w = mountCard([
      { kind: "monthly", period: "2027-01" },
      { kind: "yearly", period: "2026" },
    ])
    expect(w.findAll(".award-badge")).toHaveLength(1)
    expect(w.find(".award-badge").attributes("aria-label")).toBe("POPCORN yearly award, 2026")
  })
})
