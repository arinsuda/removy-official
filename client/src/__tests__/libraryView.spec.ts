import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter, RouterView } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import LibraryView from "@/views/library/LibraryView.vue"
import { ADD_SEARCH_DEBOUNCE_MS } from "@/views/library/config"
import { averageRating, collagePosters, ratingBuckets, reviewMonthDiff, topGenres, toEntries } from "@/views/library/libraryStats"
import { useRealI18n } from "./helpers/realI18n"
import { clickAfterRender, waitAfterDebounce, waitForEnabled } from "./helpers/waitForEnabled"

const m = vi.hoisted(() => ({
  profile: vi.fn(),
  library: vi.fn(),
  likes: vi.fn(),
  reviews: vi.fn(),
  removeItem: vi.fn(),
  addItem: vi.fn(),
  search: vi.fn(),
  searchSeries: vi.fn(),
}))

vi.mock("@/api/api", () => ({
  userApi: { getProfile: m.profile },
  libraryApi: { getVisibleUserLibrary: m.library, removeItem: m.removeItem, addItem: m.addItem },
  reviewApi: { getUserReviews: m.reviews },
  movieApi: { search: m.search, searchSeries: m.searchSeries },
}))
vi.mock("@/api/endpoints/social", async (orig) => ({
  ...(await orig<typeof import("@/api/endpoints/social")>()),
  followApi: { getFollowStatus: vi.fn().mockResolvedValue({ data: {} }), follow: vi.fn(), unfollow: vi.fn() },
}))
vi.mock("@/api/endpoints/media", () => ({ mediaApi: { getLikeByUserId: m.likes, unlikeMedia: vi.fn() } }))
vi.mock("@/api/endpoints/list", () => ({ listApi: { getOwn: vi.fn().mockResolvedValue({ data: { lists: [] } }), getForUser: vi.fn().mockResolvedValue({ data: { lists: [] } }) } }))
vi.mock("@/api/endpoints/watchLog", () => ({ watchLogApi: { getMyWatchLogs: vi.fn().mockResolvedValue({ data: {} }) } }))
vi.mock("@/views/library/components/bmol/BmolSection.vue", () => ({ default: { render: () => h("div", { "data-testid": "bmol" }) } }))

const i18n = useRealI18n()
const stub = { render: () => null }
const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { name: "user-library", path: "/users/:userId/library", component: LibraryView },
    ...["user-profile", "movies", "tv", "movie-detail", "tv-detail"].map((name) => ({ name, path: `/${name}/:id?/:userId?`, component: stub })),
  ],
})
const wrappers: VueWrapper[] = []

const media = (id: number, type = "movie") => ({ id, title: `Title ${id}`, media_type: type, poster_url: "/p.jpg", genres: [{ id: 1, name: "Drama" }], vote_average: 7, created_at: "2025-03-10T12:00:00Z" })
const items = (n: number, list: string) => Array.from({ length: n }, (_, i) => ({ id: i + 10, list_type: list, created_at: `2025-03-${10 + i}T12:00:00Z`, watched_at: null, tags: [], note: null, media: media(i + 1) }))
const review = (id: number, rating: number) => ({ id, rating, body: "b", like_count: 0, created_at: new Date().toISOString(), media: media(id) })

function seed({ watchlist = 3, watched = 2, likes = 1, reviews = [review(1, 4.5), review(2, 3)] } = {}) {
  m.profile.mockResolvedValue({ data: { user: { id: 1, username: "ann", display_name: "Ann", avatar_url: null } } })
  m.library.mockImplementation((_id: number, p: { list_type: string }) => Promise.resolve({ data: { items: items(p.list_type === "watchlist" ? watchlist : watched, p.list_type) } }))
  m.likes.mockResolvedValue({ data: { likes: items(likes, "likes").map((i) => ({ id: i.id, created_at: i.created_at, media: i.media })) } })
  m.reviews.mockResolvedValue({ data: { reviews } })
}

async function mountAt(url: string, viewerId: number) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: viewerId } as never
  await router.push(url)
  const wrapper = mount(defineComponent({ render: () => h(RouterView) }), { global: { plugins: [i18n, pinia, router, [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]] }, attachTo: document.body })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  i18n.global.locale.value = "en"
  seed()
  m.removeItem.mockResolvedValue({})
  m.addItem.mockResolvedValue({ data: {} })
  m.search.mockResolvedValue({ data: { results: [{ id: 99, title: "Found Movie", poster_path: "/f.jpg", release_date: "2024-01-01" }] } })
  m.searchSeries.mockResolvedValue({ data: { results: [] } })
})
afterEach(() => {
  vi.useRealTimers()
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  vi.clearAllMocks()
})

describe("library page shell", () => {
  it("shows the summary chips with a count per list", async () => {
    const w = await mountAt("/users/1/library", 1)
    const chips = w.findAll(".lh__chip").map((c) => c.text())
    expect(chips).toEqual(["Watchlist3", "Watched2", "Favorites1", "Reviews2"])
    expect(w.find("h1").text()).toBe("My library")
  })

  it("titles a visitor's library with the owner's name", async () => {
    const w = await mountAt("/users/1/library", 2)
    expect(w.find("h1").text()).toBe("Ann's library")
  })

  it("syncs the tab with ?tab and keeps the hash", async () => {
    const w = await mountAt("/users/1/library#top", 1)
    expect(w.find('[role="tab"][aria-selected="true"]').text()).toBe("Overview")
    await w.findAll('[role="tab"]')[1]!.trigger("click")
    await flushPromises()
    expect(router.currentRoute.value.query.tab).toBe("lists")
    expect(router.currentRoute.value.hash).toBe("#top")
    expect(w.findAll(".lc")).toHaveLength(4)
    await w.findAll('[role="tab"]')[0]!.trigger("click")
    await flushPromises()
    expect(router.currentRoute.value.query.tab).toBeUndefined()
  })

  it("opens the tab named by the URL and ignores unknown values", async () => {
    const w = await mountAt("/users/1/library?tab=bmol", 1)
    expect(w.find('[data-testid="bmol"]').exists()).toBe(true)
    await router.replace("/users/1/library?tab=nope")
    await flushPromises()
    expect(w.find('[role="tab"][aria-selected="true"]').text()).toBe("Overview")
  })

  it("renders Thai copy when the locale is th", async () => {
    i18n.global.locale.value = "th"
    const w = await mountAt("/users/1/library", 1)
    expect(w.find("h1").text()).toBe("คลังของฉัน")
  })
})

describe("dashboard", () => {
  it("shows stat cards and the up-next rail", async () => {
    const w = await mountAt("/users/1/library", 1)
    expect(w.findAll(".ds__stat")).toHaveLength(6)
    expect(w.find(".ds__stat-value").text()).toBe("2")
    expect(w.findAll(".lr")).toHaveLength(2)
    expect(w.text()).toContain("Top genres")
  })

  it("shows the onboarding empty state when the library is empty", async () => {
    seed({ watchlist: 0, watched: 0, likes: 0, reviews: [] })
    const w = await mountAt("/users/1/library", 1)
    expect(w.text()).toContain("Start your library")
    expect(w.find(".ds__stats").exists()).toBe(false)
  })

  it("shows an error state with retry when the profile fails", async () => {
    m.profile.mockRejectedValue(new Error("down"))
    const w = await mountAt("/users/1/library", 1)
    expect(w.text()).toContain("Could not load this library")
    m.profile.mockResolvedValue({ data: { user: { id: 1, username: "ann", display_name: "Ann", avatar_url: null } } })
    await w.find(".app-state button").trigger("click")
    await flushPromises()
    expect(w.find("h1").exists()).toBe(true)
  })

  it("shows the locked state instead of an error when a list answers 403", async () => {
    m.library.mockRejectedValue(Object.assign(new Error("forbidden"), { isAxiosError: true, response: { status: 403 } }))
    const w = await mountAt("/users/1/library", 2)
    expect(w.text()).toContain("This account is private")
    expect(w.text()).not.toContain("Could not load this library")
    expect(w.find(".lib__layout").exists()).toBe(false)
  })

  it("warns but still renders when one list fails", async () => {
    m.likes.mockRejectedValue(new Error("down"))
    const w = await mountAt("/users/1/library", 1)
    expect(w.find(".lib__warn").exists()).toBe(true)
    expect(w.findAll(".ds__stat")).toHaveLength(6)
  })
})

describe("lists", () => {
  it("opens a list from its card and returns with the back button", async () => {
    const w = await mountAt("/users/1/library?tab=lists", 1)
    await w.findAll(".lc")[0]!.trigger("click")
    await flushPromises()
    expect(router.currentRoute.value.query).toMatchObject({ tab: "lists", list: "watchlist" })
    expect(w.find(".ld__title").text()).toBe("Watchlist")
    clickAfterRender(w.find(".ld__head button").element as HTMLElement)
    await flushPromises()
    expect(router.currentRoute.value.query.list).toBeUndefined()
    expect(w.findAll(".lc")).toHaveLength(4)
  })

  it("lets the owner remove an item after confirming", async () => {
    const w = await mountAt("/users/1/library?tab=lists&list=watchlist", 1)
    clickAfterRender(w.find('[data-action="remove"]').element as HTMLElement)
    const confirm = await vi.waitFor(() => {
      const el = document.body.querySelector<HTMLButtonElement>(".confirm-actions button:last-child")
      if (!el) throw new Error("confirm dialog not open")
      return el
    })
    clickAfterRender(confirm)
    await flushPromises()
    expect(m.removeItem).toHaveBeenCalledWith(12)
  })

  it("hides every owner action from a visitor", async () => {
    const w = await mountAt("/users/1/library?tab=lists&list=watchlist", 2)
    expect(w.find('[data-action="remove"]').exists()).toBe(false)
    expect(w.find('[data-action="mark-watched"]').exists()).toBe(false)
    expect(w.find('[data-action="add-titles"]').exists()).toBe(false)
    expect(w.text()).not.toContain("Add titles")
  })

  it("lets the owner search and add a title to the watchlist", async () => {
    const w = await mountAt("/users/1/library?tab=lists&list=watchlist", 1)
    await w.find('[data-action="add-titles"]').trigger("click")
    await flushPromises()
    const input = document.body.querySelector<HTMLInputElement>('input[type="search"].las__input')!
    input.value = "found"
    input.dispatchEvent(new Event("input"))
    // Assert on outcomes (result rendered, request sent) instead of racing fake timers.
    const addButton = await waitForEnabled('[data-action="add"]', ADD_SEARCH_DEBOUNCE_MS)
    clickAfterRender(addButton)
    await vi.waitFor(
      () => expect(m.addItem).toHaveBeenCalledWith({ media_id: 99, media_type: "movie", list_type: "watchlist" }),
      waitAfterDebounce(ADD_SEARCH_DEBOUNCE_MS),
    )
  })

  it("shows the empty state of a list with no titles", async () => {
    seed({ watchlist: 0 })
    const w = await mountAt("/users/1/library?tab=lists&list=watchlist", 1)
    expect(w.text()).toContain("Your watchlist is empty")
  })
})

describe("library stats", () => {
  const rev = (rating: number, createdAt = new Date().toISOString()) => ({ rating, created_at: createdAt }) as never

  it("averages ratings and buckets them by half star, highest first", () => {
    expect(averageRating([rev(4), rev(5)])).toBe(4.5)
    expect(averageRating([])).toBe(0)
    const buckets = ratingBuckets([rev(5), rev(5), rev(0.5)])
    expect(buckets[0]).toMatchObject({ rating: 5, count: 2, percent: 100 })
    expect(buckets.at(-1)).toMatchObject({ rating: 0.5, count: 1, percent: 50 })
  })

  it("compares this month with last month across a year boundary", () => {
    const now = new Date(2025, 0, 15)
    expect(reviewMonthDiff([rev(4, "2025-01-03T00:00:00"), rev(4, "2025-01-04T00:00:00"), rev(4, "2024-12-20T00:00:00")], now)).toBe(1)
  })

  it("ranks genres and picks poster paths, skipping titles without one", () => {
    const entries = toEntries([...items(2, "watched"), { ...items(1, "watched")[0]!, id: 50, media: { ...media(7), poster_url: "", genres: [{ id: 2, name: "Comedy" }] } }] as never)
    expect(topGenres(entries, 5)).toEqual([{ name: "Drama", count: 2 }, { name: "Comedy", count: 1 }])
    expect(collagePosters(entries, 4)).toHaveLength(2)
  })
})
