import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createMemoryHistory, createRouter } from "vue-router"
import { i18n } from "../i18n"
import HeaderSearch from "../components/search/HeaderSearch.vue"
import { highlightParts } from "../components/search/highlight"
import { movieApi } from "../api/endpoints/movie"

vi.mock("../api/endpoints/movie", () => ({
  movieApi: { search: vi.fn(), searchSeries: vi.fn() },
}))

const page = (results: unknown[]) => ({ data: { page: 1, total_pages: 1, total_results: results.length, results } })
const movie = { id: 1, title: "Spider-Man", poster_path: "/a.jpg", release_date: "2002-05-01", vote_average: 7.4, ratings: { popcorn: { average: 4.2, count: 3, available: true, scale: 5 } } }
const show = { id: 2, name: "Spidey Show", poster_path: null, first_air_date: "2019-01-01", vote_average: 0 }

describe("highlightParts", () => {
  it("flags case-insensitive matches and escapes regex chars", () => {
    expect(highlightParts("Spider-Man", "SPI")).toEqual([
      { text: "Spi", match: true },
      { text: "der-Man", match: false },
    ])
    expect(highlightParts("A (b) c", "(b")).toEqual([
      { text: "A ", match: false },
      { text: "(b", match: true },
      { text: ") c", match: false },
    ])
  })
  it("returns the whole text when the query is empty", () => {
    expect(highlightParts("Dune", "  ")).toEqual([{ text: "Dune", match: false }])
  })
})

describe("HeaderSearch", () => {
  let wrapper: VueWrapper
  let router: ReturnType<typeof createRouter>

  const type = async (value: string) => {
    await wrapper.find("input").setValue(value)
    await vi.advanceTimersByTimeAsync(350)
    await flushPromises()
  }
  const key = (k: string) => wrapper.find("input").trigger("keydown", { key: k })

  beforeEach(async () => {
    vi.useFakeTimers()
    vi.mocked(movieApi.search).mockResolvedValue(page([movie]) as never)
    vi.mocked(movieApi.searchSeries).mockResolvedValue(page([show]) as never)
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/", component: { template: "<div />" } },
        { path: "/movies/:id", component: { template: "<div />" } },
        { path: "/tv/:id", component: { template: "<div />" } },
        { path: "/search", name: "search-results", component: { template: "<div />" } },
      ],
    })
    await router.push("/")
    wrapper = mount(HeaderSearch, { global: { plugins: [i18n, router] }, attachTo: document.body })
    await wrapper.find(".hs-toggle").trigger("click")
  })

  afterEach(() => {
    wrapper.unmount()
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it("does not search below 2 characters", async () => {
    await type("s")
    expect(movieApi.search).not.toHaveBeenCalled()
    expect(wrapper.find(".hs-panel").exists()).toBe(false)
  })

  it("shows skeleton rows while loading, then options with highlighted text", async () => {
    vi.mocked(movieApi.search).mockReturnValue(new Promise(() => {}))
    await wrapper.find("input").setValue("spide")
    await vi.advanceTimersByTimeAsync(350)
    expect(wrapper.findAll(".hs-skel")).toHaveLength(3)
    vi.mocked(movieApi.search).mockResolvedValue(page([movie]) as never)
    await type("spider")
    expect(wrapper.findAll('[role="option"]')).toHaveLength(3) // movie, show, view all
    expect(wrapper.find("mark").text().toLowerCase()).toBe("spider")
    expect(wrapper.find("input").attributes("aria-expanded")).toBe("true")
  })

  it("moves with ArrowUp/ArrowDown and exposes aria-activedescendant", async () => {
    await type("spide")
    const input = wrapper.find("input")
    expect(input.attributes("aria-activedescendant")).toBeUndefined()
    await key("ArrowDown")
    expect(input.attributes("aria-activedescendant")).toBe("header-search-option-0")
    expect(wrapper.findAll('[role="option"]')[0]!.attributes("aria-selected")).toBe("true")
    await key("ArrowUp")
    expect(input.attributes("aria-activedescendant")).toBe("header-search-option-2") // wraps to view all
    await key("ArrowDown")
    expect(input.attributes("aria-activedescendant")).toBe("header-search-option-0")
  })

  it("Enter opens the active item, or the full results page when none is active", async () => {
    await type("spide")
    await key("ArrowDown")
    await key("Enter")
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe("/movies/1")

    await wrapper.find(".hs-toggle").trigger("click")
    await type("spide")
    await key("Enter")
    await flushPromises()
    expect(router.currentRoute.value.name).toBe("search-results")
    expect(router.currentRoute.value.query.q).toBe("spide")
  })

  it("Escape closes and clears", async () => {
    await type("spide")
    await key("Escape")
    expect(wrapper.find("input").exists()).toBe(false)
    expect(wrapper.find(".hs-panel").exists()).toBe(false)
  })

  it("clear button empties the query and hides the panel", async () => {
    await type("spide")
    await wrapper.find('.hs-clear').trigger("click")
    expect((wrapper.find("input").element as HTMLInputElement).value).toBe("")
    expect(wrapper.find(".hs-panel").exists()).toBe(false)
  })

  it("shows the empty state", async () => {
    vi.mocked(movieApi.search).mockResolvedValue(page([]) as never)
    vi.mocked(movieApi.searchSeries).mockResolvedValue(page([]) as never)
    await type("zzzz")
    expect(wrapper.find(".hs-state").text()).toContain("zzzz")
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it("shows an error state and retries", async () => {
    vi.mocked(movieApi.search).mockRejectedValueOnce(new Error("boom"))
    await type("spide")
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    await wrapper.find(".hs-retry").trigger("click")
    await flushPromises()
    expect(wrapper.findAll('[role="option"]').length).toBeGreaterThan(0)
  })
})
