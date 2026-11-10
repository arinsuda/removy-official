import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import Upcoming from "@/views/movie/UpcomingView.vue"
import { movieApi } from "@/api/api"
import { i18n } from "@/i18n"

vi.mock("@/api/api", () => ({
  movieApi: { getUpcoming: vi.fn(), getGenres: vi.fn() },
}))

const getUpcoming = vi.mocked(movieApi.getUpcoming)
const m = (id: number, date: string) => ({ id, title: `Movie ${id}`, release_date: date, poster_path: "/p", backdrop_path: "/b", overview: "", vote_average: 0, popularity: id })
const page = (results: ReturnType<typeof m>[] | null) =>
  ({ data: { page: 1, total_pages: 1, total_results: 0, results } }) as unknown as Awaited<ReturnType<typeof movieApi.getUpcoming>>

async function mountView() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/upcoming", name: "upcoming", component: Upcoming },
      { path: "/movies", name: "movies", component: { template: "<div />" } },
      { path: "/movies/:id", name: "movie-detail", component: { template: "<div />" } },
    ],
  })
  await router.push("/upcoming")
  const wrapper = mount(Upcoming, {
    global: {
      plugins: [i18n, router, [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]],
    },
  })
  await flushPromises()
  return { wrapper, router }
}

describe("Upcoming view", () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ["Date"] })
    vi.setSystemTime(new Date(2026, 10, 8, 12))
    i18n.global.locale.value = "en"
    vi.mocked(movieApi.getGenres).mockResolvedValue({ data: { genres: [{ id: 28, name: "Action" }] } } as never)
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
    i18n.global.locale.value = "th"
  })

  it("shows a skeleton while loading", async () => {
    getUpcoming.mockReturnValue(new Promise(() => {}))
    const { wrapper } = await mountView()
    expect(wrapper.find('[aria-busy="true"]').exists()).toBe(true)
  })

  it("shows an error and retries", async () => {
    getUpcoming.mockRejectedValueOnce(new Error("boom"))
    const { wrapper } = await mountView()
    expect(wrapper.get('[role="alert"]').text()).toContain("Could not load upcoming movies")
    getUpcoming.mockResolvedValueOnce(page([m(1, "2026-11-20")]))
    await wrapper.get('[role="alert"] button').trigger("click")
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).toContain("Movie 1")
  })

  it("shows the empty state for a null or empty result", async () => {
    getUpcoming.mockResolvedValue(page(null))
    const { wrapper } = await mountView()
    expect(wrapper.text()).toContain("Nothing scheduled yet")
  })

  it("groups titles under labelled sections and links to now playing", async () => {
    getUpcoming.mockResolvedValue(page([m(1, "2026-11-08"), m(2, "2026-11-10"), m(3, "2026-12-25")]))
    const { wrapper } = await mountView()
    const headings = wrapper.findAll("h2.timeline__heading").map((h) => h.text())
    expect(headings[0]).toContain("This week")
    expect(headings[1]).toContain("Next week")
    expect(headings[2]).toContain("December 2026")
    const sections = wrapper.findAll("section[aria-labelledby]")
    expect(sections.length).toBeGreaterThanOrEqual(3)
    expect(wrapper.find('a[href="/movies?tab=now"]').exists()).toBe(true)
  })

  it("keeps the hash when the genre changes", async () => {
    getUpcoming.mockResolvedValue(page([m(1, "2026-11-20")]))
    const { wrapper, router } = await mountView()
    await router.replace({ path: "/upcoming", hash: "#token" })
    await wrapper.get("select").setValue("28")
    await flushPromises()
    expect(router.currentRoute.value.query.genre).toBe("28")
    expect(router.currentRoute.value.hash).toBe("#token")
    expect(getUpcoming).toHaveBeenLastCalledWith(1, 28)
  })
})
