import { afterEach, describe, expect, it, vi } from "vitest"
import { createApp, defineComponent, ref, type Ref } from "vue"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { fromMovie } from "@/components/home/homeMedia"
import type { BrowseConfig } from "@/components/browse/browseConfig"
import { DEFAULT_BROWSE_STATE, type BrowseState } from "@/components/browse/browseQuery"
import { useBrowseData } from "@/composables/useBrowseData"
import type { Movie } from "@/types"

const flush = async (ms = 20) => new Promise((r) => setTimeout(r, ms))

const movie = (id: number, avg = 0, count = 0) =>
  fromMovie({ id, title: `m${id}`, poster_path: "/p", ratings: { popcorn: { average: avg, count, available: !!count, scale: 5 } } } as unknown as Movie)

function setup(config: BrowseConfig, state: Ref<BrowseState>) {
  let result!: ReturnType<typeof useBrowseData>
  const app = createApp(
    defineComponent({
      setup() {
        result = useBrowseData(config, state)
        return () => null
      },
    }),
  )
  app.use(VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) })
  app.mount(document.createElement("div"))
  return { result, app }
}

function config(pages: Record<number, number[]>, hot: Set<number> = new Set()) {
  const fetchList = vi.fn(async (_tab: string, page: number, _genre: number | null) => ({
    page,
    total_pages: Object.keys(pages).length,
    total_results: 0,
    results: (pages[page] ?? []).map((id) => (hot.has(id) ? movie(id, 4.5, 3) : movie(id))),
  }))
  const cfg: BrowseConfig = {
    type: "movie",
    routeName: "movies",
    genreServerSide: true,
    fetchList: fetchList as unknown as BrowseConfig["fetchList"],
    fetchSearch: vi.fn(async (_q: string, page: number) => ({ page, total_pages: 1, total_results: 0, results: [movie(99)] })),
    fetchGenres: async () => [{ id: 28, name: "Action" }],
  }
  return { cfg, fetchList }
}

describe("useBrowseData", () => {
  let app: ReturnType<typeof setup>["app"] | null = null
  afterEach(() => app?.unmount())

  it("loads pages, dedupes and passes tab and genre to the API", async () => {
    const { cfg, fetchList } = config({ 1: [1, 2], 2: [2, 3] })
    const state = ref<BrowseState>({ ...DEFAULT_BROWSE_STATE, tab: "top", genre: 28 })
    const s = setup(cfg, state)
    app = s.app
    await flush()
    expect(fetchList).toHaveBeenCalledWith("top", 1, 28)
    expect(s.result.items.value.map((m) => m.id)).toEqual([1, 2])
    expect(s.result.hasNextPage.value).toBe(true)
    await s.result.fetchNextPage()
    await flush()
    expect(s.result.items.value.map((m) => m.id)).toEqual([1, 2, 3])
    expect(s.result.hasNextPage.value).toBe(false)
  })

  it("refetches when genre changes (genre is in the key)", async () => {
    const { cfg, fetchList } = config({ 1: [1] })
    const state = ref<BrowseState>({ ...DEFAULT_BROWSE_STATE })
    const s = setup(cfg, state)
    app = s.app
    await flush()
    state.value = { ...state.value, genre: 35 }
    await flush()
    expect(fetchList).toHaveBeenLastCalledWith("popular", 1, 35)
  })

  it("uses search for 2+ chars and ignores 1 char", async () => {
    const { cfg, fetchList } = config({ 1: [1] })
    const state = ref<BrowseState>({ ...DEFAULT_BROWSE_STATE, q: "d" })
    const s = setup(cfg, state)
    app = s.app
    await flush()
    expect(s.result.searching.value).toBe(false)
    state.value = { ...state.value, q: "du" }
    await flush()
    expect(cfg.fetchSearch).toHaveBeenCalledWith("du", 1)
    expect(s.result.items.value.map((m) => m.id)).toEqual([99])
    expect(fetchList).toHaveBeenCalledTimes(1)
  })

  it("auto-loads more pages while a tier has too few matches", async () => {
    const { cfg, fetchList } = config({ 1: [1, 2], 2: [3, 4], 3: [5, 6] }, new Set([5]))
    const state = ref<BrowseState>({ ...DEFAULT_BROWSE_STATE, tier: "hot" })
    const s = setup(cfg, state)
    app = s.app
    await flush(60)
    expect(fetchList).toHaveBeenCalledTimes(3)
    expect(s.result.items.value.map((m) => m.id)).toEqual([5])
    expect(s.result.tierCounts.value.hot).toBe(1)
  })

  afterEach(() => vi.clearAllMocks())
})
