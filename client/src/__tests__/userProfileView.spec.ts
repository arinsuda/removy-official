import { beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import UserProfileView from "@/views/user/UserProfileView.vue"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({
  getProfile: vi.fn(),
  getStats: vi.fn(),
  getUserReviews: vi.fn(),
  getUserBMOL: vi.fn(),
  getFollowStatus: vi.fn(),
  follow: vi.fn(),
  unfollow: vi.fn(),
}))

vi.mock("@/api/api", () => ({
  userApi: { getProfile: m.getProfile, getStats: m.getStats },
  reviewApi: { getUserReviews: m.getUserReviews },
  bmolApi: { getUserBMOL: m.getUserBMOL },
  followApi: { getFollowStatus: m.getFollowStatus, follow: m.follow, unfollow: m.unfollow },
}))
// useFollow imports followApi straight from the endpoint module.
vi.mock("@/api/endpoints/social", () => ({
  followApi: { getFollowStatus: m.getFollowStatus, follow: m.follow, unfollow: m.unfollow },
}))
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

async function mountAt(path: string) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: 1 } as never
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: "/users/:userId", component: UserProfileView }] })
  await router.push(path)
  await router.isReady()
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = mount(UserProfileView, {
    global: { plugins: [i18n, pinia, router, [VueQueryPlugin, { queryClient: client }]], stubs },
    attachTo: document.body,
  })
  await flushPromises()
  return { wrapper, router }
}

beforeEach(() => {
  vi.clearAllMocks()
  m.getProfile.mockResolvedValue({ data: { user: profile() } })
  m.getStats.mockResolvedValue({ data: { stats: { review_count: 1, watchlist_count: 2, like_count: 3, watched_count: 4, level: 2 } } })
  m.getUserReviews.mockResolvedValue({ data: { reviews: [] } })
  m.getUserBMOL.mockResolvedValue({ data: { items: [] } })
  m.getFollowStatus.mockResolvedValue({ data: { is_following: false, is_followed_by: false } })
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
