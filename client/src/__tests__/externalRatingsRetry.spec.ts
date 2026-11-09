import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h, ref } from "vue"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"

const m = vi.hoisted(() => ({
  getById: vi.fn(),
  getVideos: vi.fn(),
  recordMediaView: vi.fn(),
}))
vi.mock("@/api/api", () => ({ movieApi: { getById: m.getById, getVideos: m.getVideos, getSeriesById: vi.fn(), getTVVideos: vi.fn() } }))
vi.mock("@/api/endpoints/media", () => ({ mediaApi: { recordMediaView: m.recordMediaView } }))

import { provideHomePopup, type HomePopup } from "@/composables/useHomePopup"
import { useMediaDetail } from "@/composables/useMediaDetail"
import type { HomeMedia } from "@/components/home/homeMedia"
import type { DetailConfig, DetailData } from "@/components/detail/detailConfig"

const src = { available: false, average: null, count: 0, scale: 10 }
const ratings = (pending: boolean, imdb = false) => ({
  popcorn: { ...src, scale: 5 },
  tmdb: src,
  ...(imdb ? { imdb: { available: true, average: 8, count: 1, scale: 10 } } : {}),
  ...(pending ? { external_pending: true } : {}),
})
const movieRes = (r: ReturnType<typeof ratings>) => ({ data: { movie: { ratings: r } } })

describe("popup ratings retry", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
    vi.stubGlobal("matchMedia", () => ({ matches: true, addEventListener: vi.fn(), removeEventListener: vi.fn() }))
    m.getById.mockReset()
    m.getVideos.mockReset().mockResolvedValue({ data: { results: [] } })
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  const item = { key: "movie-1", id: 1, type: "movie" } as HomeMedia
  function setup() {
    let popup!: HomePopup
    mount(defineComponent({ setup: () => ((popup = provideHomePopup()), () => h("div")) }))
    popup.show(item, "rail:movie-1", document.createElement("div"))
    return popup
  }

  it("retries while pending and stops once ratings arrive", async () => {
    m.getById.mockResolvedValueOnce(movieRes(ratings(true))).mockResolvedValueOnce(movieRes(ratings(false, true)))
    const popup = setup()
    await vi.advanceTimersByTimeAsync(300)
    expect(m.getById).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1600)
    expect(m.getById).toHaveBeenCalledTimes(2)
    expect(popup.detailRatings(item)?.imdb?.available).toBe(true)
    await vi.advanceTimersByTimeAsync(5000)
    expect(m.getById).toHaveBeenCalledTimes(2)
  })

  it("gives up after two retries", async () => {
    m.getById.mockResolvedValue(movieRes(ratings(true)))
    setup()
    await vi.advanceTimersByTimeAsync(10000)
    expect(m.getById).toHaveBeenCalledTimes(3)
  })

  it("does not retry when not pending", async () => {
    m.getById.mockResolvedValue(movieRes(ratings(false)))
    setup()
    await vi.advanceTimersByTimeAsync(10000)
    expect(m.getById).toHaveBeenCalledTimes(1)
  })

  it("stops retrying once the popup closed", async () => {
    m.getById.mockResolvedValue(movieRes(ratings(true)))
    const popup = setup()
    await vi.advanceTimersByTimeAsync(300)
    popup.close()
    await vi.advanceTimersByTimeAsync(10000)
    expect(m.getById).toHaveBeenCalledTimes(1)
  })
})

describe("detail ratings retry", () => {
  beforeEach(() => {
    vi.useFakeTimers()
    m.recordMediaView.mockReset().mockResolvedValue(undefined)
  })
  afterEach(() => vi.useRealTimers())

  const detail = (id: number, r: ReturnType<typeof ratings>, title = "T") => ({ id, title, ratings: r }) as unknown as DetailData

  function setup(fetch: DetailConfig["fetch"]) {
    const id = ref(1)
    let api!: ReturnType<typeof useMediaDetail>
    mount(defineComponent({ setup: () => ((api = useMediaDetail({ type: "movie", fetch } as DetailConfig, id)), () => h("div")) }))
    return { api, id }
  }

  it("refetches once when pending and merges only ratings", async () => {
    const fetch = vi.fn()
      .mockResolvedValueOnce(detail(1, ratings(true), "first"))
      .mockResolvedValueOnce(detail(1, ratings(false, true), "changed"))
    const { api } = setup(fetch)
    await flushPromises()
    expect(api.data.value?.ratings?.imdb).toBeUndefined()
    await vi.advanceTimersByTimeAsync(1600)
    expect(fetch).toHaveBeenCalledTimes(2)
    expect(api.data.value?.title).toBe("first")
    expect(api.data.value?.ratings?.imdb?.available).toBe(true)
    await vi.advanceTimersByTimeAsync(5000)
    expect(fetch).toHaveBeenCalledTimes(2)
  })

  it("stops after two retries", async () => {
    const fetch = vi.fn().mockResolvedValue(detail(1, ratings(true)))
    setup(fetch)
    await flushPromises()
    await vi.advanceTimersByTimeAsync(10000)
    expect(fetch).toHaveBeenCalledTimes(3)
  })

  it("does not retry when not pending", async () => {
    const fetch = vi.fn().mockResolvedValue(detail(1, ratings(false)))
    setup(fetch)
    await flushPromises()
    await vi.advanceTimersByTimeAsync(10000)
    expect(fetch).toHaveBeenCalledTimes(1)
  })

  it("drops a retry result when the id changed meanwhile", async () => {
    let resolveRetry!: (d: DetailData) => void
    const fetch = vi.fn()
      .mockResolvedValueOnce(detail(1, ratings(true)))
      .mockImplementationOnce(() => new Promise<DetailData>(r => (resolveRetry = r)))
      .mockResolvedValueOnce(detail(2, ratings(false)))
    const { api, id } = setup(fetch)
    await flushPromises()
    await vi.advanceTimersByTimeAsync(1600)
    id.value = 2
    await flushPromises()
    resolveRetry(detail(1, ratings(false, true)))
    await flushPromises()
    expect(api.data.value?.id).toBe(2)
    expect(api.data.value?.ratings?.imdb).toBeUndefined()
  })
})
