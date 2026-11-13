import { beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import UserProfileView from "@/views/user/UserProfileView.vue"
import { useAuthGate } from "@/composables/useAuthGate"
import { useAuthStore } from "@/stores/auth"
import { clickAfterRender } from "./helpers/waitForEnabled"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({
  getProfile: vi.fn(),
  getStats: vi.fn(),
  getUserReviews: vi.fn(),
  getUserBMOL: vi.fn(),
  getVisibleUserLibrary: vi.fn(),
  getMyWatchLogs: vi.fn(),
  getFollowStatus: vi.fn(),
  getFollowers: vi.fn(),
  getFollowing: vi.fn(),
  getPendingRequests: vi.fn(),
  follow: vi.fn(),
  unfollow: vi.fn(),
}))

vi.mock("@/api/api", () => ({
  userApi: { getProfile: m.getProfile, getStats: m.getStats },
  reviewApi: { getUserReviews: m.getUserReviews },
  bmolApi: { getUserBMOL: m.getUserBMOL },
  libraryApi: { getVisibleUserLibrary: m.getVisibleUserLibrary },
  followApi: {
    getFollowStatus: m.getFollowStatus,
    getFollowers: m.getFollowers,
    getFollowing: m.getFollowing,
    getPendingRequests: m.getPendingRequests,
    follow: m.follow,
    unfollow: m.unfollow,
  },
}))
// useFollow imports followApi straight from the endpoint module.
vi.mock("@/api/endpoints/social", () => ({
  followApi: {
    getFollowStatus: m.getFollowStatus,
    getFollowers: m.getFollowers,
    getFollowing: m.getFollowing,
    getPendingRequests: m.getPendingRequests,
    follow: m.follow,
    unfollow: m.unfollow,
  },
}))
vi.mock("@/api/endpoints/review", () => ({ reviewApi: { getUserReviews: m.getUserReviews } }))
vi.mock("@/api/endpoints/library", () => ({ libraryApi: { getVisibleUserLibrary: m.getVisibleUserLibrary } }))
vi.mock("@/api/endpoints/watchLog", () => ({ watchLogApi: { getMyWatchLogs: m.getMyWatchLogs } }))
vi.mock("@/api/endpoints/achievement", () => ({ achievementApi: { listByUser: vi.fn().mockResolvedValue({ data: [] }) } }))

const i18n = useRealI18n()

const profile = (over = {}) => ({
  id: 5, username: "alice", display_name: "Alice", bio: null, avatar_url: null, is_private: false,
  level: 2, review_count: 1, follower_count: 3, following_count: 4, created_at: "2024-01-01T00:00:00Z", ...over,
})
const httpError = (status: number) => ({ isAxiosError: true, response: { status } })

const tabStub = (name: string) => defineComponent({ name, render: () => h("div", { "data-test": name }) })
const stubs = Object.fromEntries(
  ["UserReviews", "UserWatchlist", "UserWatched", "UserLikes", "UserAchievements", "UserBmol", "FollowListModal", "ProfileTastePanel", "ProfileBmolHighlight"].map((n) => [n, tabStub(n)]),
)

async function mountAt(path: string, loggedIn = true, withStubs = true) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const auth = useAuthStore()
  auth.user = loggedIn ? ({ id: 1 } as never) : null
  auth.isInitialized = true
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/users/:userId", component: UserProfileView },
      { path: "/movies/:id", name: "movie-detail", component: tabStub("MovieDetail") },
      { path: "/tv/:id", name: "tv-detail", component: tabStub("TVDetail") },
    ],
  })
  await router.push(path)
  await router.isReady()
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = mount(UserProfileView, {
    global: {
      plugins: [i18n, pinia, router, [VueQueryPlugin, { queryClient: client }]],
      ...(withStubs ? { stubs } : {}),
    },
    attachTo: document.body,
  })
  await flushPromises()
  return { wrapper, router }
}

beforeEach(() => {
  vi.clearAllMocks()
  m.getProfile.mockResolvedValue({ data: { user: profile() } })
  m.getStats.mockResolvedValue({ data: { stats: { review_count: 1, watchlist_count: 2, like_count: 3, watched_count: 4, level: 2 } } })
  m.getUserReviews.mockResolvedValue({ data: { reviews: [], pagination: { page: 1, page_size: 20, total: 0, has_more: false }, summary: { count: 0, average_rating: 0, filter_counts: { all: 0, high: 0, low: 0, with_text: 0 } } } })
  m.getUserBMOL.mockResolvedValue({ data: { items: [] } })
  m.getVisibleUserLibrary.mockResolvedValue({ data: { items: [{ id: 10, media: { id: 603, media_type: "movie", title: "The Matrix", poster_path: "/m.jpg" }, list_type: "watched", created_at: "2025-01-01T00:00:00Z" }] } })
  m.getMyWatchLogs.mockResolvedValue({ data: { logs: [] } })
  m.getFollowStatus.mockResolvedValue({ data: { is_following: false, is_followed_by: false } })
  m.getFollowers.mockResolvedValue({ data: { followers: [] } })
  m.getFollowing.mockResolvedValue({ data: { following: [] } })
  m.getPendingRequests.mockResolvedValue({ data: { requests: [] } })
})

describe("UserProfileView states", () => {
  it("shows the not-found state on 404", async () => {
    m.getProfile.mockRejectedValue(httpError(404))
    const { wrapper } = await mountAt("/users/5")
    expect(wrapper.text()).toContain(i18n.global.t("profile.states.notFoundTitle"))
  })

  it("shows a retryable error on other failures", async () => {
    m.getProfile.mockRejectedValueOnce(httpError(500))
    const { wrapper } = await mountAt("/users/5")
    expect(wrapper.text()).toContain(i18n.global.t("profile.states.errorTitle"))
    await wrapper.find("[role=alert] button").trigger("click")
    await flushPromises()
    expect(m.getProfile).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain("Alice")
  })

  it("locks a private profile: no tabs and no section requests", async () => {
    m.getProfile.mockResolvedValue({ data: { user: profile({ is_private: true }) } })
    const { wrapper } = await mountAt("/users/5")
    expect(wrapper.text()).toContain(i18n.global.t("profile.states.lockedTitle"))
    expect(wrapper.find("[role=tablist]").exists()).toBe(false)
    expect(m.getStats).not.toHaveBeenCalled()
    expect(m.getUserReviews).not.toHaveBeenCalled()
    expect(m.getUserBMOL).not.toHaveBeenCalled()
  })
})

describe("UserProfileView tabs", () => {
  it("opens the tab from ?tab and writes it back on change", async () => {
    const { wrapper, router } = await mountAt("/users/5?tab=likes")
    expect(wrapper.find('[data-test="UserLikes"]').exists()).toBe(true)
    const watched = wrapper.findAll("[role=tab]").find((b) => b.text().startsWith(i18n.global.t("profile.tabs.watched")))!
    await watched.trigger("click")
    await flushPromises()
    expect(router.currentRoute.value.query.tab).toBe("watched")
    expect(wrapper.find('[data-test="UserWatched"]').exists()).toBe(true)
  })
})

describe("UserProfileView follow flow", () => {
  it("follow -> requested -> cancel", async () => {
    m.getProfile.mockResolvedValue({ data: { user: profile({ is_private: true }) } })
    let state: "pending" | undefined
    m.getFollowStatus.mockImplementation(() => Promise.resolve({ data: { is_following: false, follow_status: state, is_followed_by: false } }))
    m.follow.mockImplementation(() => { state = "pending"; return Promise.resolve({ data: { status: "pending" } }) })
    m.unfollow.mockImplementation(() => { state = undefined; return Promise.resolve({}) })

    const { wrapper } = await mountAt("/users/5")
    const label = (k: string) => i18n.global.t(k)
    const btn = (text: string) => wrapper.findAll("button").find((b) => b.text().trim().startsWith(text))!

    await btn(label("profile.follow.follow")).trigger("click")
    await flushPromises()
    expect(m.follow).toHaveBeenCalledWith(5)
    expect(btn(label("profile.follow.requested"))).toBeTruthy()

    await btn(label("profile.follow.requested")).trigger("click")
    await flushPromises()
    expect(m.unfollow).toHaveBeenCalledWith(5)
    expect(wrapper.text()).toContain(label("profile.states.lockedTitle"))
  })
})

describe("UserProfileView guest mode", () => {
  beforeEach(() => {
    useAuthGate().close()
  })

  it("renders a public profile with real child components for a guest without auth-only calls and gates follow through useAuthGate", async () => {
    const { wrapper } = await mountAt("/users/5", false, false)
    expect(wrapper.find(".hero__name").text()).toBe("Alice")
    expect(wrapper.find("[role=tablist]").exists()).toBe(true)
    expect(wrapper.find("aside.taste").exists()).toBe(true)
    expect(m.getFollowStatus).not.toHaveBeenCalled()
    expect(m.getPendingRequests).not.toHaveBeenCalled()

    const watchedTab = wrapper.findAll("[role=tab]").find((b) => b.text().startsWith(i18n.global.t("profile.tabs.watched")))!
    clickAfterRender(watchedTab.element as HTMLElement)
    await flushPromises()

    expect(m.getVisibleUserLibrary).toHaveBeenCalledWith(5, { list_type: "watched" })
    expect(m.getMyWatchLogs).not.toHaveBeenCalled()

    const followBtn = wrapper.findAll("button").find((b) => b.text().includes(i18n.global.t("profile.follow.logInToFollow")))!
    expect(followBtn).toBeTruthy()
    clickAfterRender(followBtn.element as HTMLElement)

    expect(m.follow).not.toHaveBeenCalled()
    expect(useAuthGate().isOpen.value).toBe(true)
    expect(useAuthGate().reason.value).toBe("follow")
    wrapper.unmount()
  })

  it("renders a private profile with ProfileHero, counts, ProfileLockedState, and Log in to follow for a guest", async () => {
    m.getProfile.mockResolvedValue({ data: { user: profile({ is_private: true }) } })
    const { wrapper } = await mountAt("/users/5", false, false)

    expect(wrapper.find(".hero__name").text()).toBe("Alice")
    const counts = wrapper.findAll(".hero__counts .hero__count")
    expect(counts).toHaveLength(3)
    expect(counts[1]!.find(".hero__num").text()).toBe("3")
    expect(counts[1]!.find(".hero__lbl").text()).toBe(i18n.global.t("profile.hero.followers"))
    expect(counts[2]!.find(".hero__num").text()).toBe("4")
    expect(counts[2]!.find(".hero__lbl").text()).toBe(i18n.global.t("profile.hero.following"))

    expect(wrapper.text()).toContain(i18n.global.t("profile.states.lockedTitle"))
    expect(wrapper.find("[role=tablist]").exists()).toBe(false)
    expect(m.getFollowStatus).not.toHaveBeenCalled()
    expect(m.getStats).not.toHaveBeenCalled()
    expect(m.getUserReviews).not.toHaveBeenCalled()
    expect(m.getUserBMOL).not.toHaveBeenCalled()
    expect(m.getMyWatchLogs).not.toHaveBeenCalled()

    const followBtns = wrapper.findAll("button").filter((b) => b.text().includes(i18n.global.t("profile.follow.logInToFollow")))
    expect(followBtns.length).toBeGreaterThanOrEqual(1)
    clickAfterRender(followBtns[0]!.element as HTMLElement)

    expect(m.follow).not.toHaveBeenCalled()
    expect(useAuthGate().isOpen.value).toBe(true)
    expect(useAuthGate().reason.value).toBe("follow")
    wrapper.unmount()
  })
})
