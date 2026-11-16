import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import { watchLogApi } from "@/api/endpoints/watchLog"
import UserWatched from "@/components/profile/UserWatched.vue"
import UserWatchlist from "@/components/profile/UserWatchlist.vue"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

const { libraryItems } = vi.hoisted(() => {
  const base = { created_at: "2024-01-01", watched_at: "2024-01-01", poster_url: "", genres: [], vote_average: 7 }
  return {
    libraryItems: [
      { ...base, id: 10, media: { ...base, id: 1, title: "Movie One", media_type: "movie" } },
      { ...base, id: 20, media: { ...base, id: 2, title: "Manga Two", media_type: "manga" } },
    ],
  }
})

vi.mock("@/api/api", () => ({
  reviewApi: { getUserReviews: vi.fn().mockResolvedValue({ data: { reviews: [] } }) },
  libraryApi: {
    getVisibleUserLibrary: vi.fn().mockResolvedValue({ data: { items: libraryItems } }),
    removeItem: vi.fn().mockResolvedValue({}),
  },
}))
vi.mock("@/api/endpoints/media", () => ({
  mediaApi: { getLikeByUserId: vi.fn().mockResolvedValue({ data: { likes: [] } }), unlikeMedia: vi.fn().mockResolvedValue({}) },
}))
vi.mock("@/api/endpoints/watchLog", () => ({
  watchLogApi: { getMyWatchLogs: vi.fn().mockResolvedValue({ data: {} }) },
}))

const i18n = useRealI18n()
const router = createRouter({
  history: createMemoryHistory(),
  routes: ["movie-detail", "tv-detail", "manga-detail", "movies"].map((name) => ({ name, path: `/${name}/:id?`, component: { render: () => null } })),
})
const VIEW_KEY = "popcorn-profile-view:watched"
const origGetClientRects = HTMLElement.prototype.getClientRects
const wrappers: VueWrapper[] = []

beforeEach(() => {
  i18n.global.locale.value = "en"
  localStorage.clear()
  vi.mocked(watchLogApi.getMyWatchLogs).mockClear()
  HTMLElement.prototype.getClientRects = function (this: HTMLElement) {
    return (this.isConnected ? [{ width: 1, height: 1 }] : []) as unknown as DOMRectList
  }
})
afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  HTMLElement.prototype.getClientRects = origGetClientRects
})

/** Mounts a profile tab as its owner (user 1), so the owner-only actions render. */
async function mountTab(comp: typeof UserWatched | typeof UserWatchlist, listType: "watched" | "watchlist") {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: 1 } as never
  const w = mount(comp, {
    props: { userId: 1, listType },
    global: { plugins: [i18n, pinia, router] },
    attachTo: document.body,
  })
  wrappers.push(w)
  await flushPromises()
  return w
}

describe("manga on the profile watchlist", () => {
  it("offers Mark watched for movies only, and labels manga Want to read", async () => {
    const w = await mountTab(UserWatchlist, "watchlist")
    expect(w.findAll('[data-action="mark-watched"]')).toHaveLength(1)
    const badges = w.findAll('[data-test="manga-status-badge"]')
    expect(badges.map((b) => b.text())).toEqual(["Want to read"])
  })
})

describe("manga on the profile watched tab", () => {
  it("hides history and log actions for manga but keeps remove", async () => {
    const w = await mountTab(UserWatched, "watched")
    // List view (default): one movie row can log or read history, the manga row cannot.
    expect(w.findAll('[data-action="log"]')).toHaveLength(1)
    expect(w.findAll('[data-action="history"]')).toHaveLength(1)
    expect(w.findAll('[data-action="remove"]')).toHaveLength(2)
  })

  it("labels manga Read on the grid badge and skips watch-log fetches for it", async () => {
    localStorage.setItem(VIEW_KEY, "grid")
    const w = await mountTab(UserWatched, "watched")
    expect(w.findAll('[data-test="manga-status-badge"]').map((b) => b.text())).toEqual(["Read"])
    expect(w.findAll('[data-action="log"]')).toHaveLength(1)
    // Only the movie gets a count fetch; the manga would be rejected by the server.
    expect(watchLogApi.getMyWatchLogs).toHaveBeenCalledTimes(1)
    expect(watchLogApi.getMyWatchLogs).toHaveBeenCalledWith("movie", 1)
  })
})
