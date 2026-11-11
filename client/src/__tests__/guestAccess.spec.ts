/**
 * @vitest-environment happy-dom
 */
import { beforeEach, describe, expect, it, vi } from "vitest"
import axios, { type AxiosAdapter, type InternalAxiosRequestConfig } from "axios"

const auth = vi.hoisted(() => ({
  isInitialized: true,
  isLoggedIn: false,
  isAdmin: false,
  initError: null as string | null,
  user: null as { favorite_genres?: string | null; onboarded_at?: string | null } | null,
  initialize: vi.fn(),
  clearUser: vi.fn(),
}))
vi.mock("@/stores/auth", () => ({ useAuthStore: () => auth }))
vi.mock("@/stores/admin", () => ({ useAdminStore: () => ({ reset: vi.fn() }) }))
vi.mock("@/i18n", () => ({ updateDocumentTitle: vi.fn() }))
vi.mock("@/composables/useSocket", () => ({ resume: vi.fn() }))

// Stub lazy route views so router.push() tests the guard without compiling heavy page trees.
vi.mock("@/layouts/MainLayout.vue", () => ({ default: {} }))
vi.mock("@/views/HomeView.vue", () => ({ default: {} }))
vi.mock("@/views/SearchResultView.vue", () => ({ default: {} }))
vi.mock("@/views/movie/UpcomingView.vue", () => ({ default: {} }))
vi.mock("@/views/actor/ActorDetailView.vue", () => ({ default: {} }))
vi.mock("@/views/FeedView.vue", () => ({ default: {} }))
vi.mock("@/views/movie/MoviesView.vue", () => ({ default: {} }))
vi.mock("@/views/tv/TVSeriesView.vue", () => ({ default: {} }))
vi.mock("@/views/NotificationsView.vue", () => ({ default: {} }))
vi.mock("@/views/settings/SettingsView.vue", () => ({ default: {} }))
vi.mock("@/views/user/UserProfileView.vue", () => ({ default: {} }))
vi.mock("@/views/library/LibraryView.vue", () => ({ default: {} }))
vi.mock("@/views/achievement/AchievementsView.vue", () => ({ default: {} }))
vi.mock("@/views/auth/LoginView.vue", () => ({ default: {} }))
vi.mock("@/views/onboarding/FavoriteGenreView.vue", () => ({ default: {} }))
vi.mock("@/views/movie/MovieDetailView.vue", () => ({ default: {} }))
vi.mock("@/views/tv/TVDetailView.vue", () => ({ default: {} }))
vi.mock("@/views/aboutus/AboutUsView.vue", () => ({ default: {} }))
vi.mock("@/views/aboutus/DonateView.vue", () => ({ default: {} }))

import api from "@/api/index"
import router from "@/router"

function reject401(config: InternalAxiosRequestConfig) {
  return Promise.reject({ config, response: { status: 401, data: {}, config } })
}

describe("route guard for guests", () => {
  beforeEach(() => {
    auth.isLoggedIn = false
    auth.user = null
  })

  it.each([
    "/",
    "/search?q=dune",
    "/upcoming",
    "/actors/100",
    "/movies/603",
    "/tv/1399",
    "/about-us",
    "/donate",
  ])("lets a guest open %s", async (path) => {
    await router.push(path)
    expect(router.currentRoute.value.fullPath).toBe(path)
  })

  it.each([
    "/feed",
    "/movies",
    "/tv",
    "/users/1",
    "/users/1/library",
    "/users/1/achievements",
    "/notifications",
    "/settings",
  ])("sends a guest from private page %s to login with redirect query", async (path) => {
    await router.push(path)
    expect(router.currentRoute.value.name).toBe("login")
    expect(router.currentRoute.value.query.redirect).toBe(path)
  })

  it("sends a signed-in, not onboarded user from a public page to onboarding, but never a guest", async () => {
    auth.isLoggedIn = true
    auth.user = { favorite_genres: null, onboarded_at: null }
    await router.push("/movies/603")
    expect(router.currentRoute.value.name).toBe("onboarding")
    auth.isLoggedIn = false
    auth.user = null
    await router.push("/movies/603")
    expect(router.currentRoute.value.name).toBe("movie-detail")
  })

  it("sends a signed-in user without onboarding to onboarding, but not once onboarded_at is set", async () => {
    auth.isLoggedIn = true
    auth.user = { favorite_genres: null, onboarded_at: null }
    await router.push("/feed")
    expect(router.currentRoute.value.name).toBe("onboarding")
    auth.user = { favorite_genres: null, onboarded_at: "2026-01-01T00:00:00Z" }
    await router.push("/feed")
    expect(router.currentRoute.value.name).toBe("feed")
  })
})

describe("api interceptor for guests", () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    localStorage.clear()
    api.defaults.adapter = reject401 as unknown as AxiosAdapter
  })

  it("rejects a 401 without refreshing once the session is initialized and logged out", async () => {
    auth.isInitialized = true
    auth.isLoggedIn = false
    const post = vi.spyOn(axios, "post")
    await expect(api.get("/lists")).rejects.toBeTruthy()
    expect(post).not.toHaveBeenCalled()
    expect(auth.clearUser).not.toHaveBeenCalled()
  })

  it("still tries to refresh while the session is bootstrapping", async () => {
    auth.isInitialized = false
    const post = vi.spyOn(axios, "post").mockRejectedValue(new Error("down"))
    await api.get("/lists").catch(() => undefined)
    expect(post).toHaveBeenCalled()
  })
})
