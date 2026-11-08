import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import DetailWatchProviders from "@/components/detail/DetailWatchProviders.vue"
import { i18n } from "@/i18n"
import type { WatchProviders } from "@/types/movie"

const getWatchProviders = vi.fn()
vi.mock("@/api/endpoints/movie", () => ({ movieApi: { getWatchProviders: (...a: unknown[]) => getWatchProviders(...a) } }))

const p = (id: number, name: string) => ({ provider_id: id, provider_name: name, logo_path: `/${id}.jpg`, display_priority: id })
const providers = (over: Partial<WatchProviders> = {}): WatchProviders => ({
  region: "TH", link: "https://www.themoviedb.org/movie/1/watch?locale=TH",
  flatrate: [], free: [], ads: [], rent: [], buy: [], ...over,
})

function mountIt() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return mount(DetailWatchProviders, { props: { type: "movie", id: 1 }, global: { plugins: [i18n, [VueQueryPlugin, { queryClient }]] } })
}

describe("DetailWatchProviders", () => {
  beforeEach(() => {
    getWatchProviders.mockReset()
    i18n.global.locale.value = "en"
  })

  it("shows a loading skeleton first", () => {
    getWatchProviders.mockReturnValue(new Promise(() => {}))
    const w = mountIt()
    expect(w.find('[data-test="providers-loading"]').exists()).toBe(true)
    expect(w.text()).toContain("Data from JustWatch")
  })

  it("groups providers and merges free with ads", async () => {
    getWatchProviders.mockResolvedValue({ data: { providers: providers({ flatrate: [p(1, "Netflix")], free: [p(2, "Tubi")], ads: [p(3, "Pluto")], buy: [p(4, "Apple")] }) } })
    const w = mountIt()
    await flushPromises()
    const groups = w.findAll('[data-test="providers-group"]')
    expect(groups.map(g => g.find("h3").text())).toEqual(["Stream", "Free / with ads", "Buy"])
    expect(groups[1].findAll("img").map(i => i.attributes("alt"))).toEqual(["Tubi", "Pluto"])
    expect(groups[0].find("img").attributes("src")).toContain("/w92/1.jpg")
    expect(groups[0].element.tagName).toBe("A")
    expect(groups[0].attributes("rel")).toBe("noopener noreferrer")
    expect(groups[0].attributes("target")).toBe("_blank")
  })

  it("does not link non-https URLs", async () => {
    getWatchProviders.mockResolvedValue({ data: { providers: providers({ link: "http://evil.example/x", flatrate: [p(1, "Netflix")] }) } })
    const w = mountIt()
    await flushPromises()
    expect(w.find("a").exists()).toBe(false)
    expect(w.findAll('[data-test="providers-group"]')).toHaveLength(1)
  })

  it("shows the empty state with the region", async () => {
    getWatchProviders.mockResolvedValue({ data: { providers: providers() } })
    const w = mountIt()
    await flushPromises()
    expect(w.get('[data-test="providers-empty"]').text()).toBe("Not available to stream in TH")
  })

  it("shows an error and retries", async () => {
    getWatchProviders.mockRejectedValueOnce(new Error("boom"))
    const w = mountIt()
    await flushPromises()
    expect(w.find('[role="alert"]').exists()).toBe(true)
    getWatchProviders.mockResolvedValue({ data: { providers: providers({ rent: [p(5, "Google")] }) } })
    await w.get('[role="alert"] button').trigger("click")
    await flushPromises()
    expect(w.findAll('[data-test="providers-group"]')).toHaveLength(1)
  })
})
