import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import type { Component } from "vue"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import UserAchievements from "@/components/profile/UserAchievements.vue"
import UserLikes from "@/components/profile/UserLikes.vue"
import UserReviews from "@/components/profile/UserReviews.vue"
import UserWatched from "@/components/profile/UserWatched.vue"
import UserWatchlist from "@/components/profile/UserWatchlist.vue"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({
  library: vi.fn(),
  likes: vi.fn(),
  reviews: vi.fn(),
  achievements: vi.fn(),
}))

vi.mock("@/api/api", () => ({
  libraryApi: { getVisibleUserLibrary: m.library, removeItem: vi.fn().mockResolvedValue({}) },
  reviewApi: { getUserReviews: m.reviews },
}))
vi.mock("@/api/endpoints/media", () => ({ mediaApi: { getLikeByUserId: m.likes, unlikeMedia: vi.fn() } }))
vi.mock("@/api/endpoints/watchLog", () => ({
  watchLogApi: { getMyWatchLogs: vi.fn().mockResolvedValue({ data: { summary: { watch_count: 3 } } }) },
}))
vi.mock("@/api/endpoints/achievement", () => ({ achievementApi: { listByUser: m.achievements } }))

const i18n = useRealI18n()
const router = createRouter({
  history: createMemoryHistory(),
  routes: ["movie-detail", "tv-detail", "movies", "user-achievements"].map((name) => ({ name, path: `/${name}/:id?/:userId?`, component: { render: () => null } })),
})
const wrappers: VueWrapper[] = []

const media = (id: number, type = "movie") => ({ id, title: `Title ${id}`, media_type: type, poster_url: "/p.jpg", genres: [], vote_average: 7 })
const libItems = (n: number) => Array.from({ length: n }, (_, i) => ({ id: i + 10, created_at: "2025-03-10T12:00:00Z", watched_at: "2025-03-10T12:00:00Z", media: media(i + 1, i === 1 ? "tv" : "movie") }))

beforeEach(async () => {
  i18n.global.locale.value = "en"
  await router.push("/movies")
  m.library.mockResolvedValue({ data: { items: libItems(3) } })
  m.likes.mockResolvedValue({ data: { likes: libItems(3).map((i) => ({ created_at: i.created_at, media: i.media })) } })
  m.reviews.mockResolvedValue({
    data: {
      reviews: [1, 2].map((id) => ({ id, rating: id === 1 ? 4.5 : 2, body: `Body ${id}`, visibility: "public", like_count: 1, comment_count: 0, helpful_count: 0, is_liked: false, is_helpful_voted: false, created_at: "2025-01-01T00:00:00Z", updated_at: "2025-01-01T00:00:00Z", media: { ...media(id), genres: [{ id: 1, name: "Drama" }] } })),
    },
  })
  m.achievements.mockResolvedValue({
    data: [
      { achievement_id: 1, achievement: { id: 1, name: "First", description: "d", exp_reward: 5, action_type: "review_count", target_count: 1 }, current_count: 1, is_unlocked: true, unlocked_at: "2025-02-01T00:00:00Z", progress_pct: 100 },
      { achievement_id: 2, achievement: { id: 2, name: "Ten", description: "d", exp_reward: 5, action_type: "review_count", target_count: 10 }, current_count: 4, is_unlocked: false, unlocked_at: null, progress_pct: 40 },
      { achievement_id: 3, achievement: { id: 3, name: "Fifty", description: "d", exp_reward: 5, action_type: "review_count", target_count: 50 }, current_count: 0, is_unlocked: false, unlocked_at: null, progress_pct: 0 },
    ],
    pagination: { page: 1, limit: 100, total: 3, total_pages: 1 },
  })
})
afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  vi.clearAllMocks()
})

function mountTab(comp: Component, viewerId: number, listType?: string) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: viewerId } as never
  const wrapper = mount(comp, { props: { userId: 1, listType }, global: { plugins: [i18n, pinia, router] }, attachTo: document.body })
  wrappers.push(wrapper)
  return wrapper
}

const ownerActions = '[data-action="remove"], [data-action="mark-watched"], [data-action="log"], [data-action="history"], [data-action="edit"], [data-action="delete"]'

describe.each<[string, Component, string | undefined]>([
  ["UserWatchlist", UserWatchlist, "watchlist"],
  ["UserLikes", UserLikes, "likes"],
  ["UserWatched", UserWatched, "watched"],
  ["UserReviews", UserReviews, undefined],
])("%s owner vs visitor", (_n, comp, listType) => {
  it("shows owner actions only to the owner", async () => {
    const owner = mountTab(comp, 1, listType)
    await flushPromises()
    expect(owner.findAll(ownerActions).length).toBeGreaterThan(0)

    const visitor = mountTab(comp, 2, listType)
    await flushPromises()
    expect(visitor.findAll(ownerActions)).toHaveLength(0)
  })
})

describe("poster tabs", () => {
  it("renders a poster card per title with a heart badge on favorites", async () => {
    const w = mountTab(UserLikes, 2, "likes")
    await flushPromises()
    expect(w.findAll(".poster-card")).toHaveLength(3)
    expect(w.findAll('[role="img"][aria-label="Favorite"]')).toHaveLength(3)
  })

  it("filters TV only and searches within the list", async () => {
    const w = mountTab(UserWatchlist, 1, "watchlist")
    await flushPromises()
    const chips = w.findAll(".tb__chip")
    await chips.find((c) => c.text().startsWith("TV"))!.trigger("click")
    expect(w.findAll(".poster-card")).toHaveLength(1)
    await chips[0]!.trigger("click")
    await w.find("input[type=search]").setValue("title 3")
    expect(w.findAll(".poster-card")).toHaveLength(1)
    await w.find("input[type=search]").setValue("nothing")
    expect(w.text()).toContain("Nothing matches")
  })

  it.each([
    ["UserWatchlist", UserWatchlist, "watchlist", "Your watchlist is empty", "Nothing on this watchlist yet"],
    ["UserLikes", UserLikes, "likes", "No favorites yet", "No favorites yet"],
  ])("%s empty state differs for owner and visitor", async (_n, comp, listType, ownerText, visitorText) => {
    m.library.mockResolvedValue({ data: { items: [] } })
    m.likes.mockResolvedValue({ data: { likes: [] } })
    const owner = mountTab(comp, 1, listType)
    await flushPromises()
    expect(owner.text()).toContain(ownerText)
    expect(owner.find("a.app-btn").exists()).toBe(true)
    const visitor = mountTab(comp, 2, listType)
    await flushPromises()
    expect(visitor.text()).toContain(visitorText)
    expect(visitor.find("a.app-btn").exists()).toBe(false)
  })

  it("shows the error state when loading fails", async () => {
    m.library.mockRejectedValue(new Error("x"))
    vi.spyOn(console, "error").mockImplementation(() => {})
    const w = mountTab(UserWatchlist, 1, "watchlist")
    await flushPromises()
    expect(w.text()).toContain("Could not load this section")
  })
})

describe("UserWatched", () => {
  it("groups the timeline by month and shows watch counts only to the owner", async () => {
    m.library.mockResolvedValue({
      data: { items: [{ id: 1, created_at: "2025-03-10T12:00:00Z", watched_at: "2025-03-10T12:00:00Z", media: media(1) }, { id: 2, created_at: "2025-01-10T12:00:00Z", watched_at: "2025-01-10T12:00:00Z", media: media(2) }] },
    })
    const owner = mountTab(UserWatched, 1, "watched")
    await flushPromises()
    expect(owner.findAll(".wt__month").map((h) => h.text())).toEqual([expect.stringContaining("March 2025"), expect.stringContaining("January 2025")])
    expect(owner.text()).toContain("×3")

    const visitor = mountTab(UserWatched, 2, "watched")
    await flushPromises()
    expect(visitor.text()).not.toContain("×3")
  })

  it("narrows to a year and can switch to the grid", async () => {
    m.library.mockResolvedValue({
      data: { items: [{ id: 1, created_at: "2025-03-10T12:00:00Z", watched_at: "2025-03-10T12:00:00Z", media: media(1) }, { id: 2, created_at: "2024-01-10T12:00:00Z", watched_at: "2024-01-10T12:00:00Z", media: media(2) }] },
    })
    const w = mountTab(UserWatched, 2, "watched")
    await flushPromises()
    await w.find("select[aria-label=Year]").setValue("2024")
    expect(w.findAll(".wt__row")).toHaveLength(1)
    await w.findAll(".tb__view-btn")[0]!.trigger("click")
    expect(w.findAll(".poster-card")).toHaveLength(1)
  })
})

describe("UserReviews", () => {
  it("summarizes and filters critical reviews", async () => {
    const w = mountTab(UserReviews, 2)
    await flushPromises()
    expect(w.find(".ur__summary").text()).toContain("3.3")
    expect(w.find(".ur__summary").text()).toContain("Drama")
    expect(w.findAll("article")).toHaveLength(2)
    await w.findAll(".tb__chip").find((c) => c.text().startsWith("Critical"))!.trigger("click")
    expect(w.findAll("article")).toHaveLength(1)
  })

  it("shows the visitor empty state without a call to action", async () => {
    m.reviews.mockResolvedValue({ data: { reviews: [] } })
    const w = mountTab(UserReviews, 2)
    await flushPromises()
    expect(w.text()).toContain("No reviews yet")
    expect(w.find("a.app-btn").exists()).toBe(false)
  })
})

describe("UserAchievements", () => {
  it("counts unlocked badges and greys out locked ones with progress", async () => {
    const w = mountTab(UserAchievements, 2)
    await flushPromises()
    expect(w.find(".ua__count").text()).toBe("1 / 3 unlocked")
    expect(w.findAll(".ab--on")).toHaveLength(1)
    expect(w.findAll(".ab")).toHaveLength(3)
    expect(w.findAll('.ab [role="progressbar"]')).toHaveLength(2)
    expect(w.text()).toContain("4 / 10")
  })

  it("filters locked and opens the detail dialog", async () => {
    const w = mountTab(UserAchievements, 2)
    await flushPromises()
    await w.findAll(".ua__chip")[2]!.trigger("click")
    expect(w.findAll(".ab")).toHaveLength(2)
    await w.findAll(".ab")[0]!.trigger("click")
    await flushPromises()
    expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain("Ten")
  })
})
