import { beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h, ref } from "vue"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import DetailPage from "@/components/detail/DetailPage.vue"
import type { DetailConfig, DetailData } from "@/components/detail/detailConfig"
import { i18n } from "@/i18n"

vi.mock("@/api/api", () => ({ movieApi: {}, reviewApi: {}, followApi: {} }))
vi.mock("@/api/endpoints/media", () => ({ mediaApi: { recordMediaView: vi.fn(() => Promise.resolve()) } }))
vi.mock("@/composables/useMediaActivity", () => ({
  useMediaActivity: () => ({
    stats: ref(null),
    watchSummary: ref(null),
    isLiked: ref(false),
    isWatchlisted: ref(false),
    isWatched: ref(false),
    pending: ref(new Set()),
    requireLogin: () => true,
    toggleLike: vi.fn(),
    toggleWatchlist: vi.fn(),
    markLogged: vi.fn(),
    refreshStats: vi.fn(),
  }),
}))

const stub = (name: string) => defineComponent({ name, render: () => h("div", { "data-stub": name }) })
const stubs = { MovieReviews: stub("MovieReviews"), HomeRail: stub("HomeRail"), HomePosterCard: true, PopupTrailer: true, ShareImageModal: true, WatchLogModal: true }

const data = (id: number, title: string): DetailData => ({
  id, type: "movie", title, englishTitle: null, tagline: "", overview: "Plot", year: "2010", runtimeMin: 100,
  seasonCount: null, episodeCount: null, status: null, genres: [], ratings: null, voteAverage: 0,
  posterPath: null, backdropPath: null, homepage: "", facts: [], cast: [], crewHighlights: [], videos: [], tvSeasons: [], networks: [],
})

function deferred<T>() {
  let resolve!: (v: T) => void
  let reject!: (e: unknown) => void
  const promise = new Promise<T>((res, rej) => ((resolve = res), (reject = rej)))
  return { promise, resolve, reject }
}

function setup(fetch: DetailConfig["fetch"]) {
  const config: DetailConfig = { type: "movie", fetch, fetchSimilar: async () => [], fallbackRoute: { name: "movies" } }
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/movies", name: "movies", component: { template: "<div />" } },
      { path: "/login", name: "login", component: { template: "<div />" } },
    ],
  })
  const wrapper = mount(DetailPage, { props: { config, id: 1 }, global: { plugins: [i18n, createPinia(), router], stubs } })
  return wrapper
}

describe("DetailPage", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    i18n.global.locale.value = "en"
  })

  it("shows a skeleton while loading", () => {
    const w = setup(() => new Promise(() => {}))
    expect(w.find('[aria-busy="true"]').exists()).toBe(true)
  })

  it("renders the 404 state with a way back", async () => {
    const err = Object.assign(new Error("nf"), { isAxiosError: true, response: { status: 404 } })
    const w = setup(() => Promise.reject(err))
    await flushPromises()
    const nf = w.get('[data-test="detail-not-found"]')
    expect(nf.text()).toContain("We could not find that movie")
    expect(nf.text()).toContain("Browse movies")
  })

  it("shows an error and retries", async () => {
    const fetch = vi.fn().mockRejectedValueOnce(new Error("boom")).mockResolvedValueOnce(data(1, "Inception"))
    const w = setup(fetch)
    await flushPromises()
    expect(w.get('[role="alert"]').text()).toContain("Could not load this title")
    await w.get('[role="alert"] button').trigger("click")
    await flushPromises()
    expect(w.find('[role="alert"]').exists()).toBe(false)
    expect(w.get("h1").text()).toBe("Inception")
    expect(fetch).toHaveBeenCalledTimes(2)
  })

  it("drops a stale response when the id changes mid-flight", async () => {
    const first = deferred<DetailData>()
    const second = deferred<DetailData>()
    const fetch = vi.fn((id: number) => (id === 1 ? first.promise : second.promise))
    const w = setup(fetch)
    await w.setProps({ id: 2 })
    second.resolve(data(2, "Second"))
    await flushPromises()
    first.resolve(data(1, "First"))
    await flushPromises()
    expect(w.get("h1").text()).toBe("Second")
  })

  it("keeps the old title and shows a progress bar while the next one loads", async () => {
    const next = deferred<DetailData>()
    const fetch = vi.fn((id: number) => (id === 1 ? Promise.resolve(data(1, "First")) : next.promise))
    const w = setup(fetch)
    await flushPromises()
    await w.setProps({ id: 2 })
    expect(w.get("h1").text()).toBe("First")
    expect(w.find(".detail__progress").exists()).toBe(true)
    next.resolve(data(2, "Second"))
    await flushPromises()
    expect(w.get("h1").text()).toBe("Second")
    expect(w.find(".detail__progress").exists()).toBe(false)
  })
})
