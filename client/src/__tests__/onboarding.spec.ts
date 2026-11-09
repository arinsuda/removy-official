import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { i18n } from "@/i18n"
import { useAuthStore } from "@/stores/auth"
import FavoriteGenreView from "@/views/onboarding/FavoriteGenreView.vue"
import { extractGenres, pickDiverse } from "@/components/onboarding/onboardingConfig"
import type { Movie } from "@/types"

const push = vi.fn()
const updateFavoriteGenres = vi.fn()
const getPopular = vi.fn()
const getTopRated = vi.fn()

vi.mock("vue-router", () => ({ useRouter: () => ({ push }) }))
vi.mock("@/api/api", () => ({
  movieApi: { getPopular: (p: number) => getPopular(p), getTopRated: (p: number) => getTopRated(p) },
  userApi: { updateFavoriteGenres: (...a: unknown[]) => updateFavoriteGenres(...a) },
}))

function movie(id: number, genre_ids: number[]): Movie {
  return { id, title: `m${id}`, poster_path: `/p${id}.jpg`, vote_average: 7, vote_count: 100, genre_ids } as unknown as Movie
}

// 6 movies; 28 appears on 1-4, 35 on 1-2, 18 on 5.
const POOL = [movie(1, [28, 35]), movie(2, [28, 35]), movie(3, [28]), movie(4, [28]), movie(5, [18]), movie(6, [27])]

function stubMatchMedia(reduce: boolean) {
  vi.stubGlobal("matchMedia", () => ({ matches: reduce, addEventListener: vi.fn(), removeEventListener: vi.fn() }))
}

async function mountView(reduce = true) {
  stubMatchMedia(reduce)
  const pinia = createPinia()
  setActivePinia(pinia)
  const auth = useAuthStore()
  auth.setUser({ id: 7, username: "u", favorite_genres: null } as never)
  const w = mount(FavoriteGenreView, { global: { plugins: [i18n, pinia] } })
  await flushPromises()
  return { w, auth }
}

const tiles = (w: Awaited<ReturnType<typeof mountView>>["w"]) => w.findAll(".tile__btn")
const go = (w: Awaited<ReturnType<typeof mountView>>["w"]) => w.get(".bar__go button")

describe("onboarding view", () => {
  beforeEach(() => {
    i18n.global.locale.value = "en"
    push.mockReset()
    updateFavoriteGenres.mockReset().mockResolvedValue({})
    getPopular.mockReset().mockResolvedValue({ data: { results: POOL } })
    getTopRated.mockReset().mockResolvedValue({ data: { results: [] } })
  })

  it("renders pressable tiles and keeps Continue disabled below the minimum", async () => {
    const { w } = await mountView()
    expect(tiles(w)).toHaveLength(POOL.length)
    expect(tiles(w)[0]!.attributes("aria-pressed")).toBe("false")
    expect(go(w).attributes("disabled")).toBeDefined()
    expect(w.text()).toContain("Pick 5 more")
  })

  it("selects and deselects tiles, updating the count and live region", async () => {
    const { w } = await mountView()
    await tiles(w)[0]!.trigger("click")
    await tiles(w)[1]!.trigger("click")
    expect(tiles(w).filter((b) => b.attributes("aria-pressed") === "true")).toHaveLength(2)
    expect(w.get('[role="status"]').text()).toBe("2 of 5 selected")
    await tiles(w)[0]!.trigger("click")
    expect(w.get('[role="status"]').text()).toBe("1 of 5 selected")
    expect(w.get('[role="progressbar"]').attributes("aria-valuenow")).toBe("1")
  })

  it("enables Continue at the minimum and submits the top genres, then leaves", async () => {
    const { w, auth } = await mountView()
    const picked = tiles(w)
    const ids = [1, 2, 3, 4, 5]
    for (const b of picked) {
      if (ids.some((id) => b.text().includes(`m${id}`) && !b.text().includes(`m${id}0`))) await b.trigger("click")
    }
    expect(go(w).attributes("disabled")).toBeUndefined()
    await go(w).trigger("click")
    await flushPromises()
    expect(updateFavoriteGenres).toHaveBeenCalledWith(7, [28, 35, 18])
    expect(auth.user?.favorite_genres).toBe("[28,35,18]")
    expect(push).toHaveBeenCalledWith({ name: "home" })
  })

  it("skips with an empty payload and stamps onboarded_at", async () => {
    const { w, auth } = await mountView()
    await w.get(".bar__skip").trigger("click")
    await flushPromises()
    expect(updateFavoriteGenres).toHaveBeenCalledWith(7, [])
    expect(auth.user?.onboarded_at).toBeTruthy()
    expect(push).toHaveBeenCalledWith({ name: "home" })
  })

  it("still skips when saving the skip fails", async () => {
    updateFavoriteGenres.mockRejectedValue(new Error("down"))
    const { w, auth } = await mountView()
    await w.get(".bar__skip").trigger("click")
    await flushPromises()
    expect(auth.user?.onboarded_at).toBeTruthy()
    expect(push).toHaveBeenCalledWith({ name: "home" })
  })

  it("stays on the page and shows an error when saving fails", async () => {
    updateFavoriteGenres.mockRejectedValue(new Error("down"))
    const { w } = await mountView()
    for (const b of tiles(w).slice(0, 5)) await b.trigger("click")
    await go(w).trigger("click")
    await flushPromises()
    expect(push).not.toHaveBeenCalled()
    expect(w.find('[role="alert"]').text()).toContain("could not save")
  })

  it("shows a retryable error when movies fail to load", async () => {
    getPopular.mockRejectedValueOnce(new Error("down"))
    const { w } = await mountView()
    expect(w.find(".tile__btn").exists()).toBe(false)
    getPopular.mockResolvedValue({ data: { results: POOL } })
    await w.get(".app-state button").trigger("click")
    await flushPromises()
    expect(tiles(w)).toHaveLength(POOL.length)
  })

  it("marks the shell as reduced only when the user prefers reduced motion", async () => {
    expect((await mountView(true)).w.get(".onb").classes()).toContain("onb--reduced")
    expect((await mountView(false)).w.get(".onb").classes()).not.toContain("onb--reduced")
  })
})

describe("onboarding config", () => {
  it("extracts the five most common genres", () => {
    expect(extractGenres(POOL)).toEqual([28, 35, 18, 27])
    expect(extractGenres([])).toEqual([])
  })

  it("builds a deduped pool capped at the total", () => {
    const out = pickDiverse([...POOL, ...POOL, { ...movie(9, [28]), poster_path: null } as unknown as Movie], 4, () => 0.5)
    expect(out).toHaveLength(4)
    expect(new Set(out.map((m) => m.id)).size).toBe(4)
    expect(out.every((m) => m.poster_path)).toBe(true)
  })
})
