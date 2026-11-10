import { beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h, ref } from "vue"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import { useMediaActivity } from "@/composables/useMediaActivity"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import { i18n } from "@/i18n"

const m = vi.hoisted(() => ({
  getMediaStats: vi.fn(),
  likeMedia: vi.fn(),
  unlikeMedia: vi.fn(),
  getOwnMediaStatus: vi.fn(),
  addItem: vi.fn(),
  removeItem: vi.fn(),
  getMyWatchLogs: vi.fn(),
}))
vi.mock("@/api/endpoints/media", () => ({ mediaApi: { getMediaStats: m.getMediaStats, likeMedia: m.likeMedia, unlikeMedia: m.unlikeMedia } }))
vi.mock("@/api/endpoints/library", () => ({ libraryApi: { getOwnMediaStatus: m.getOwnMediaStatus, addItem: m.addItem, removeItem: m.removeItem } }))
vi.mock("@/api/endpoints/social", () => ({ reviewApi: { getMediaRating: vi.fn().mockResolvedValue({ data: { rating: { distribution: [] } } }) } }))
vi.mock("@/api/endpoints/watchLog", () => ({ watchLogApi: { getMyWatchLogs: m.getMyWatchLogs } }))

const stats = { media_id: 1, media_type: "movie", like_count: 3, view_count: 0, review_count: 0, watchlist_count: 2, average_rating: 0, has_rating: false, liked_at: null, watchlisted_at: null }

async function setup(loggedIn: boolean) {
  const pinia = createPinia()
  setActivePinia(pinia)
  if (loggedIn) useAuthStore().user = { id: 7 } as never
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: "/", component: { template: "<div />" } }, { path: "/login", name: "login", component: { template: "<div />" } }] })
  await router.push("/")
  const id = ref(1)
  let api!: ReturnType<typeof useMediaActivity>
  mount(defineComponent({ setup: () => ((api = useMediaActivity("movie", id)), () => h("div")) }), { global: { plugins: [i18n, pinia, router] } })
  await flushPromises()
  return { api, id, router }
}

describe("useMediaActivity", () => {
  beforeEach(() => {
    i18n.global.locale.value = "en"
    vi.resetAllMocks()
    m.likeMedia.mockResolvedValue({})
    m.unlikeMedia.mockResolvedValue({})
    m.getMediaStats.mockResolvedValue({ data: { stats } })
    m.getOwnMediaStatus.mockResolvedValue({ data: { in_lists: [] } })
    m.getMyWatchLogs.mockResolvedValue({ data: { logs: [], summary: { watch_count: 0, first_watched_on: null, last_watched_on: null, has_watched: false } } })
    m.addItem.mockResolvedValue({ data: { item: { id: 55 } } })
  })

  it("sends guests to login without calling the API", async () => {
    const { api, router } = await setup(false)
    await api.toggleLike()
    await api.toggleWatchlist()
    await flushPromises()
    expect(m.likeMedia).not.toHaveBeenCalled()
    expect(m.addItem).not.toHaveBeenCalled()
    expect(router.currentRoute.value.name).toBe("login")
    expect(useUIStore().toasts.at(-1)?.message).toBe("Sign in to do that.")
  })

  it("likes and unlikes, updating the count", async () => {
    const { api } = await setup(true)
    await api.toggleLike()
    expect(m.likeMedia).toHaveBeenCalledWith("movie", 1)
    expect(api.isLiked.value).toBe(true)
    expect(api.stats.value?.like_count).toBe(4)
    await api.toggleLike()
    expect(m.unlikeMedia).toHaveBeenCalledWith("movie", 1)
    expect(api.stats.value?.like_count).toBe(3)
  })

  it("ignores a second tap while the first is in flight", async () => {
    const { api } = await setup(true)
    m.likeMedia.mockReturnValue(new Promise(() => {}))
    void api.toggleLike()
    await api.toggleLike()
    expect(m.likeMedia).toHaveBeenCalledTimes(1)
  })

  it("acts on the current id", async () => {
    const { api, id } = await setup(true)
    id.value = 9
    await flushPromises()
    await api.toggleLike()
    expect(m.likeMedia).toHaveBeenCalledWith("movie", 9)
  })

  it("logging adds the title to watched and refreshes the summary", async () => {
    const { api } = await setup(true)
    await api.markLogged()
    expect(m.addItem).toHaveBeenCalledWith(expect.objectContaining({ media_id: 1, media_type: "movie", list_type: "watched" }))
    expect(api.isWatched.value).toBe(true)
    expect(m.getMyWatchLogs).toHaveBeenCalledTimes(2)
  })

  it("treats 409 already-exists as success", async () => {
    const { api } = await setup(true)
    m.addItem.mockRejectedValue(Object.assign(new Error("dup"), { isAxiosError: true, response: { status: 409 } }))
    m.getOwnMediaStatus.mockResolvedValue({ data: { in_lists: [{ list_type: "watched", item_id: 12 }] } })
    await api.markLogged()
    expect(api.isWatched.value).toBe(true)
    expect(useUIStore().toasts.some(t => t.type === "error")).toBe(false)
  })

  it("does not re-add a title already marked watched", async () => {
    m.getOwnMediaStatus.mockResolvedValue({ data: { in_lists: [{ list_type: "watched", item_id: 4 }] } })
    const { api } = await setup(true)
    await api.markLogged()
    expect(m.addItem).not.toHaveBeenCalled()
  })

  it("drops a like response that lands after the title changed", async () => {
    const { api, id } = await setup(true)
    let release!: () => void
    m.likeMedia.mockReturnValue(new Promise<void>(r => (release = r)))
    const p = api.toggleLike()
    id.value = 9
    await flushPromises()
    const before = api.stats.value
    release()
    await p
    expect(api.stats.value).toBe(before)
    expect(api.isLiked.value).toBe(false)
  })

  it("drops a watchlist removal that lands after the title changed", async () => {
    m.getOwnMediaStatus.mockResolvedValue({ data: { in_lists: [{ list_type: "watchlist", item_id: 4 }] } })
    const { api, id } = await setup(true)
    let release!: () => void
    m.removeItem.mockReturnValue(new Promise<void>(r => (release = r)))
    const p = api.toggleWatchlist()
    id.value = 9
    await flushPromises()
    const before = api.stats.value
    release()
    await p
    expect(api.stats.value).toBe(before)
  })
})
