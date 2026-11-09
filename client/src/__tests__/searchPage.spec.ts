import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import SearchResultView from "@/views/SearchResultView.vue"
import { actorApi, movieApi } from "@/api/api"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", () => ({
  movieApi: {
    search: vi.fn(),
    searchSeries: vi.fn(),
    getGenres: vi.fn(),
    getSeriesGenres: vi.fn(),
    getPopular: vi.fn(),
    getPopularSeries: vi.fn(),
    getPopularDetail: vi.fn(),
  },
  actorApi: { search: vi.fn() },
}))

const i18n = useRealI18n()
const page = (results: unknown[], total = results.length) => ({ data: { page: 1, total_pages: 1, total_results: total, results } })
const movie = (id: number, title: string, date = "2021-05-01") => ({ id, title, poster_path: "/p.jpg", release_date: date, vote_average: 7, genre_ids: [28], overview: "" })
const show = (id: number, name: string) => ({ id, name, poster_path: "/s.jpg", first_air_date: "2019-01-01", vote_average: 8, genre_ids: [18], overview: "" })

async function render(url: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: "/search", name: "search-results", component: SearchResultView }, { path: "/actors/:id", name: "actor-detail", component: { template: "<div />" } }, { path: "/:p(.*)*", component: { template: "<div />" } }],
  })
  await router.push(url)
  const w = mount(SearchResultView, {
    global: { plugins: [i18n, router, [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]], stubs: { HomePosterCard: { props: ["item", "highlight", "caption"], template: "<a class='card'>{{ item.title }}|{{ caption }}</a>" } } },
  })
  await flushPromises()
  return { w, router }
}

describe("SearchResultView", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    i18n.global.locale.value = "en"
    vi.mocked(movieApi.getGenres).mockResolvedValue({ data: { genres: [{ id: 28, name: "Action" }] } } as never)
    vi.mocked(movieApi.getSeriesGenres).mockResolvedValue({ data: { genres: [{ id: 18, name: "Drama" }] } } as never)
    vi.mocked(movieApi.search).mockResolvedValue(page([movie(1, "Dune"), movie(2, "Dune 2", "2024-03-01")]) as never)
    vi.mocked(movieApi.searchSeries).mockResolvedValue(page([show(3, "Dune: Prophecy")]) as never)
    vi.mocked(actorApi.search).mockResolvedValue(page([{ id: 9, name: "Dune Person", profile_path: null, known_for_department: "Acting", popularity: 1 }]) as never)
    vi.mocked(movieApi.getPopular).mockResolvedValue(page([movie(7, "Popular")]) as never)
    vi.mocked(movieApi.getPopularSeries).mockResolvedValue(page([]) as never)
  })

  it("shows a prompt without a query", async () => {
    const { w } = await render("/search")
    expect(w.text()).toContain("Search POPCORN")
    expect(movieApi.search).not.toHaveBeenCalled()
  })

  it("lists movies and series with counts per tab", async () => {
    const { w } = await render("/search?q=dune")
    const cards = w.findAll(".card").map((c) => c.text())
    expect(cards).toHaveLength(3)
    expect(cards.some((c) => c.includes("Series"))).toBe(true)
    const tabs = w.findAll('[role="tab"]').map((t) => t.text())
    expect(tabs).toEqual(["All 3", "Movies 2", "Series 1", "People 1"])
    expect(w.text()).toContain("3 results")
  })

  it("switches tab through the URL and keeps the hash", async () => {
    const { w, router } = await render("/search?q=dune#token")
    await w.findAll('[role="tab"]')[3]!.trigger("click")
    await flushPromises()
    expect(router.currentRoute.value.query).toMatchObject({ q: "dune", tab: "people" })
    expect(router.currentRoute.value.hash).toBe("#token")
    expect(w.text()).toContain("Dune Person")
  })

  it("filters by year", async () => {
    const { w } = await render("/search?q=dune&tab=movie&year=2024")
    expect(w.findAll(".card").map((c) => c.text())).toEqual(["Dune 2|"])
  })

  it("shows the empty state with suggestions", async () => {
    vi.mocked(movieApi.search).mockResolvedValue(page([]) as never)
    vi.mocked(movieApi.searchSeries).mockResolvedValue(page([]) as never)
    const { w } = await render("/search?q=zzzz")
    expect(w.text()).toContain('No results for "zzzz"')
    expect(w.text()).toContain("Popular right now")
    expect(w.findAll(".card").map((c) => c.text())).toEqual(["Popular|"])
  })

  it("shows an error with retry when both lists fail", async () => {
    vi.mocked(movieApi.search).mockRejectedValue(new Error("x"))
    vi.mocked(movieApi.searchSeries).mockRejectedValue(new Error("x"))
    const { w } = await render("/search?q=dune")
    expect(w.text()).toContain("couldn't load the results")
    expect(w.find('[role="alert"] button').exists()).toBe(true)
  })
})
