import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { mount } from "@vue/test-utils"

const m = vi.hoisted(() => ({ getById: vi.fn(), getVideos: vi.fn() }))
vi.mock("@/api/api", () => ({
  movieApi: { getById: m.getById, getVideos: m.getVideos, getSeriesById: vi.fn(), getTVVideos: vi.fn() },
}))

import { i18n } from "../i18n"
import MediaRating from "../components/movie/MediaRating.vue"
import { provideHomePopup, type HomePopup } from "@/composables/useHomePopup"
import type { HomeMedia } from "@/components/home/homeMedia"

const src = { available: false, average: null, count: 0, scale: 10 }
const base = { popcorn: { ...src, scale: 5 }, tmdb: src }
const res = (extra: object) => ({ data: { movie: { ratings: { ...base, ...extra } } } })
const pendingRes = res({ external_pending: true })
const imdbRes = res({ imdb: { available: true, average: 8, count: 1, scale: 10 } })

describe("popup ratings loading", () => {
  beforeEach(() => {
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
  function hover() {
    let popup!: HomePopup
    mount(defineComponent({ setup: () => ((popup = provideHomePopup()), () => h("div")) }))
    popup.show(item, "rail:movie-1", document.createElement("div"))
    return popup
  }

  it("starts the request on hover, before the popup opens", async () => {
    m.getById.mockResolvedValue(imdbRes)
    const popup = hover()
    await vi.advanceTimersByTimeAsync(0)
    expect(popup.openKey.value).toBeNull()
    expect(m.getById).toHaveBeenCalledTimes(1)
    expect(popup.detailRatings(item)?.imdb?.available).toBe(true)
  })

  it("reports loading while pending, then the value after a retry", async () => {
    m.getById.mockResolvedValueOnce(pendingRes).mockResolvedValueOnce(imdbRes)
    const popup = hover()
    expect(popup.ratingsLoading(item)).toBe(true)
    await vi.advanceTimersByTimeAsync(300)
    expect(popup.openKey.value).toBe("rail:movie-1")
    expect(popup.ratingsLoading(item)).toBe(true)
    await vi.advanceTimersByTimeAsync(1600)
    expect(popup.ratingsLoading(item)).toBe(false)
    expect(popup.detailRatings(item)?.imdb?.available).toBe(true)
  })

  it("settles as unavailable when OMDb has nothing", async () => {
    m.getById.mockResolvedValue(res({}))
    const popup = hover()
    await vi.advanceTimersByTimeAsync(300)
    expect(popup.ratingsLoading(item)).toBe(false)
    expect(m.getById).toHaveBeenCalledTimes(1)
  })

  it("stops loading after retries are exhausted", async () => {
    m.getById.mockResolvedValue(pendingRes)
    const popup = hover()
    await vi.advanceTimersByTimeAsync(10000)
    expect(m.getById).toHaveBeenCalledTimes(3)
    expect(popup.ratingsLoading(item)).toBe(false)
  })

  it("keeps retrying while the card is hovered even if the popup is not open", async () => {
    m.getById.mockResolvedValueOnce(pendingRes).mockResolvedValueOnce(imdbRes)
    const popup = hover()
    await vi.advanceTimersByTimeAsync(100)
    popup.leave("rail:movie-1") // leave before SHOW_DELAY, then re-enter
    popup.show(item, "rail:movie-1", document.createElement("div"))
    await vi.advanceTimersByTimeAsync(1700)
    expect(m.getById).toHaveBeenCalledTimes(2)
  })
})

describe("MediaRating external loading", () => {
  const mountRating = (externalLoading: boolean, ratings = base) =>
    mount(MediaRating, { global: { plugins: [i18n] }, props: { ratings, variant: "detail", externalLoading } })

  it("shows skeletons in IMDb and RT slots while loading", () => {
    const w = mountRating(true)
    expect(w.find('[data-source="imdb"] .skeleton-block').exists()).toBe(true)
    expect(w.find('[data-source="rt"] .skeleton-block').exists()).toBe(true)
    expect(w.find('[data-source="imdb"]').text()).not.toContain("—")
  })

  it("shows an em dash once settled unavailable", () => {
    const w = mountRating(false)
    expect(w.find(".skeleton-block").exists()).toBe(false)
    expect(w.find('[data-source="imdb"]').text()).toContain("—")
    expect(w.find('[data-source="rt"]').text()).toContain("—")
  })

  it("keeps real values while loading", () => {
    const w = mountRating(true, { ...base, imdb: { available: true, average: 8, count: 1, scale: 10 } } as typeof base)
    expect(w.find('[data-source="imdb"]').text()).toContain("8.0/10")
    expect(w.find('[data-source="rt"] .skeleton-block').exists()).toBe(true)
  })
})
