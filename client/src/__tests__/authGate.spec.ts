/**
 * @vitest-environment happy-dom
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { ref } from "vue"

const socketMocks = vi.hoisted(() => ({
  connect: vi.fn(),
  disconnect: vi.fn(),
  resume: vi.fn(),
  on: vi.fn(),
  off: vi.fn(),
  emit: vi.fn(),
}))

const apiMocks = vi.hoisted(() => ({
  getUnreadCount: vi.fn().mockResolvedValue({ data: { unread_count: 0 } }),
  notificationList: vi.fn().mockResolvedValue({ data: { notifications: [], unread_count: 0, pagination: {} } }),
  getFeed: vi.fn().mockResolvedValue({ data: { items: [] } }),
  getPopular: vi.fn().mockResolvedValue({ data: { results: [] } }),
  getNowPlaying: vi.fn().mockResolvedValue({ data: { results: [] } }),
  getPopularSeries: vi.fn().mockResolvedValue({ data: { results: [] } }),
  getNowAiringSeries: vi.fn().mockResolvedValue({ data: { results: [] } }),
  getTopRated: vi.fn().mockResolvedValue({ data: { results: [] } }),
  getRecommended: vi.fn().mockResolvedValue({ data: { results: [] } }),
  getMediaReviews: vi.fn().mockResolvedValue({
    data: {
      reviews: [
        {
          id: 42,
          user_id: 99,
          media_id: 603,
          media_type: "movie",
          rating: 4.5,
          body: "Great film",
          visibility: "public",
          like_count: 2,
          helpful_count: 1,
          comment_count: 0,
          is_liked: false,
          is_helpful_voted: false,
          created_at: "2026-01-01T00:00:00Z",
          user: { id: 99, username: "reviewer", display_name: "Reviewer", avatar_url: null },
        },
      ],
      pagination: { page: 1, has_more: false, total: 1 },
    },
  }),
  getMyReview: vi.fn(),
  likeReview: vi.fn(),
  unlikeReview: vi.fn(),
  markHelpful: vi.fn(),
  unmarkHelpful: vi.fn(),
  getFollowStatus: vi.fn(),
  getOwnMediaStatus: vi.fn(),
  addLibraryItem: vi.fn(),
  removeLibraryItem: vi.fn(),
  likeMedia: vi.fn(),
  unlikeMedia: vi.fn(),
  getMediaStats: vi.fn(),
}))

vi.mock("@/composables/useSocket", () => ({
  resume: socketMocks.resume,
  useSocket: () => ({
    ...socketMocks,
    isConnected: ref(false),
    isConnecting: ref(false),
  }),
}))

vi.mock("@/api/endpoints/notification", () => ({
  notiApi: {
    getUnreadCount: apiMocks.getUnreadCount,
    list: apiMocks.notificationList,
    markAsRead: vi.fn(),
    markAllAsRead: vi.fn(),
    deleteNotification: vi.fn(),
    clearAll: vi.fn(),
  },
}))

vi.mock("@/api/endpoints/feed", () => ({
  feedApi: {
    getFeed: apiMocks.getFeed,
  },
}))

vi.mock("@/api/endpoints/library", () => ({
  libraryApi: {
    getOwnMediaStatus: apiMocks.getOwnMediaStatus,
    addItem: apiMocks.addLibraryItem,
    removeItem: apiMocks.removeLibraryItem,
  },
}))

vi.mock("@/api/endpoints/media", () => ({
  mediaApi: {
    getMediaStats: apiMocks.getMediaStats,
    likeMedia: apiMocks.likeMedia,
    unlikeMedia: apiMocks.unlikeMedia,
  },
}))

vi.mock("@/api/endpoints/link", () => ({
  linkApi: {
    createHandoff: vi.fn(),
  },
}))

vi.mock("@/api/api", () => ({
  authApi: {
    getSession: vi.fn(),
    refreshToken: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  },
  userApi: { me: vi.fn() },
  movieApi: {
    getPopular: apiMocks.getPopular,
    getNowPlaying: apiMocks.getNowPlaying,
    getPopularSeries: apiMocks.getPopularSeries,
    getNowAiringSeries: apiMocks.getNowAiringSeries,
    getTopRated: apiMocks.getTopRated,
    getRecommended: apiMocks.getRecommended,
    getVideos: vi.fn().mockResolvedValue({ data: { results: [] } }),
    getTVVideos: vi.fn().mockResolvedValue({ data: { results: [] } }),
    getById: vi.fn(),
    getSeriesById: vi.fn(),
  },
  reviewApi: {
    getMediaReviews: apiMocks.getMediaReviews,
    getMyReview: apiMocks.getMyReview,
    likeReview: apiMocks.likeReview,
    unlikeReview: apiMocks.unlikeReview,
    markHelpful: apiMocks.markHelpful,
    unmarkHelpful: apiMocks.unmarkHelpful,
  },
  followApi: {
    getFollowStatus: apiMocks.getFollowStatus,
  },
}))

import AuthPromptModal from "@/components/auth/AuthPromptModal.vue"
import MobileNavSheet from "@/components/layout/MobileNavSheet.vue"
import MovieReviews from "@/components/movie/MovieReviews.vue"
import PopupCard from "@/components/movie/PopupCard.vue"
import MainLayout from "@/layouts/MainLayout.vue"
import HomeView from "@/views/HomeView.vue"
import { useAuthGate, type AuthGateReason } from "@/composables/useAuthGate"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()

const ALL_REASONS: AuthGateReason[] = [
  "rate",
  "review",
  "like",
  "helpful",
  "watchlist",
  "watchLog",
  "list",
  "bmol",
  "follow",
  "comment",
  "shareFeed",
  "generic",
]

function parseHref(href: string | null | undefined) {
  const url = new URL(href ?? "", "http://localhost")
  return { path: url.pathname, redirect: url.searchParams.get("redirect") }
}

function makeRouter(initialPath = "/movies/603?tab=reviews#top") {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/", name: "home", component: { template: "<div />" } },
      { path: "/movies", name: "movies", component: { template: "<div />" } },
      { path: "/tv", name: "tv", component: { template: "<div />" } },
      { path: "/upcoming", name: "upcoming", component: { template: "<div />" } },
      { path: "/about-us", name: "aboutUs", component: { template: "<div />" } },
      { path: "/movies/:id", name: "movie-detail", component: { template: "<div />" } },
      { path: "/tv/:id", name: "tv-detail", component: { template: "<div />" } },
      { path: "/actors/:id", name: "actor-detail", component: { template: "<div />" } },
      { path: "/login", name: "login", component: { template: "<div />" } },
      { path: "/register", name: "register", component: { template: "<div />" } },
    ],
  })
  return { router, ready: router.push(initialPath).then(() => router.isReady()) }
}

describe("useAuthGate & AuthPromptModal", () => {
  let activeWrappers: VueWrapper[] = []

  beforeEach(() => {
    setActivePinia(createPinia())
    useAuthGate().close()
    vi.clearAllMocks()
  })

  afterEach(() => {
    useAuthGate().close()
    for (const w of activeWrappers) w.unmount()
    activeWrappers = []
    document.body.innerHTML = ""
  })

  it("defines non-empty en and th prompt copy for every AuthGateReason", () => {
    for (const locale of ["en", "th"] as const) {
      i18n.global.locale.value = locale
      for (const r of ALL_REASONS) {
        const title = i18n.global.t(`auth.prompt.reasons.${r}.title`)
        const body = i18n.global.t(`auth.prompt.reasons.${r}.body`)
        expect(title, `${locale}:${r}.title`).toBeTruthy()
        expect(body, `${locale}:${r}.body`).toBeTruthy()
        expect(title).not.toContain("auth.prompt.")
        expect(body).not.toContain("auth.prompt.")
      }
    }
  })

  it("opens shared modal for guests and allows signed-in users without opening", () => {
    const auth = useAuthStore()
    const gate = useAuthGate()

    expect(gate.requireAuth("like")).toBe(false)
    expect(gate.isOpen.value).toBe(true)
    expect(gate.reason.value).toBe("like")

    gate.close()
    auth.user = { id: 7 } as never
    expect(gate.requireAuth("watchlist")).toBe(true)
    expect(gate.isOpen.value).toBe(false)
  })

  it("renders AuthPromptModal with reason copy and preserves safe redirect query", async () => {
    const { router, ready } = makeRouter("/movies/603?tab=cast#reviews")
    await ready

    const w = mount(AuthPromptModal, {
      attachTo: document.body,
      global: { plugins: [i18n, router] },
    })
    activeWrappers.push(w)

    const gate = useAuthGate()
    expect(gate.requireAuth("watchlist")).toBe(false)
    await flushPromises()

    expect(document.body.textContent).toContain(i18n.global.t("auth.prompt.reasons.watchlist.title"))
    expect(document.body.textContent).toContain(i18n.global.t("auth.prompt.reasons.watchlist.body"))

    const loginHref = document.body.querySelector('[data-test="auth-prompt-login"]')?.getAttribute("href")
    const signupHref = document.body.querySelector('[data-test="auth-prompt-signup"]')?.getAttribute("href")
    expect(parseHref(loginHref)).toEqual({ path: "/login", redirect: "/movies/603?tab=cast" })
    expect(parseHref(signupHref)).toEqual({ path: "/register", redirect: "/movies/603?tab=cast" })

    await router.push("/")
    await flushPromises()
    expect(document.body.querySelector('[data-test="auth-prompt-login"]')?.getAttribute("href")).toBe("/login")
  })
})

describe("guest shell, headers, and write entry points", () => {
  let activeWrappers: VueWrapper[] = []

  beforeEach(() => {
    setActivePinia(createPinia())
    useAuthGate().close()
    vi.clearAllMocks()
  })

  afterEach(() => {
    useAuthGate().close()
    for (const w of activeWrappers) w.unmount()
    activeWrappers = []
    document.body.innerHTML = ""
  })

  it("MainLayout shows Log in / Sign up, mounts AuthPromptModal, and triggers no socket or notification fetches for guests", async () => {
    const { router, ready } = makeRouter("/upcoming")
    await ready

    const w = mount(MainLayout, {
      attachTo: document.body,
      global: {
        plugins: [i18n, router],
        stubs: { RouterView: { template: "<div class='page-stub' />" } },
      },
    })
    activeWrappers.push(w)
    await flushPromises()

    expect(apiMocks.getUnreadCount).not.toHaveBeenCalled()
    expect(socketMocks.connect).not.toHaveBeenCalled()
    expect(w.find(".hdr-bell").exists()).toBe(false)
    expect(w.find(".hdr-account").exists()).toBe(false)

    const guestLinks = w.findAll(".hdr-guest a").map((a) => parseHref(a.attributes("href")))
    expect(guestLinks).toEqual([
      { path: "/login", redirect: "/upcoming" },
      { path: "/register", redirect: "/upcoming" },
    ])
  })

  it("MobileNavSheet hides account/library links and shows guest login/signup buttons with redirect", async () => {
    const { router, ready } = makeRouter("/actors/100")
    await ready

    const w = mount(MobileNavSheet, {
      attachTo: document.body,
      props: { modelValue: true, user: null, isAdmin: false },
      global: { plugins: [i18n, router] },
    })
    activeWrappers.push(w)
    await flushPromises()

    const panel = document.body.querySelector(".mns-panel")
    expect(panel).not.toBeNull()
    expect(panel?.querySelector(".mns-card")).toBeNull()
    const parsed = [...(panel?.querySelectorAll("a") ?? [])].map((a) => parseHref(a.getAttribute("href")))
    expect(parsed.map((p) => p.path)).not.toContain("/feed")
    expect(parsed.map((p) => p.path)).not.toContain("/settings")
    expect(parsed).toContainEqual({ path: "/login", redirect: "/actors/100" })
    expect(parsed).toContainEqual({ path: "/register", redirect: "/actors/100" })
  })

  it("PopupCard gates watched, like, and watchlist clicks through useAuthGate without network calls or toasts", async () => {
    const { router, ready } = makeRouter("/")
    await ready

    const w = mount(PopupCard, {
      global: { plugins: [i18n, router] },
      props: {
        movie: {
          id: 603,
          title: "The Matrix",
          vote_average: 8.2,
          vote_count: 100,
          overview: "Neo wakes up.",
        } as never,
        currentTrailer: null,
        trailerUnavailable: true,
        isIframeMounted: false,
        isIframeLoaded: false,
        showSkeleton: false,
        showFallback: true,
        attachPlayer: () => {},
      },
    })
    activeWrappers.push(w)
    await flushPromises()

    expect(apiMocks.getOwnMediaStatus).not.toHaveBeenCalled()

    const gate = useAuthGate()
    await w.get(".action-btn--watched").trigger("click")
    expect(gate.isOpen.value).toBe(true)
    expect(gate.reason.value).toBe("watchLog")

    gate.close()
    await w.get(".action-btn--favorite").trigger("click")
    expect(gate.isOpen.value).toBe(true)
    expect(gate.reason.value).toBe("like")

    gate.close()
    await w.get(".action-btn--watchlist").trigger("click")
    expect(gate.isOpen.value).toBe(true)
    expect(gate.reason.value).toBe("watchlist")

    expect(apiMocks.addLibraryItem).not.toHaveBeenCalled()
    expect(apiMocks.likeMedia).not.toHaveBeenCalled()
    expect(useUIStore().toasts).toHaveLength(0)
  })

  it("MovieReviews loads public reviews without auth-only calls and opens AuthPromptModal on like/helpful clicks", async () => {
    const { router, ready } = makeRouter("/movies/603")
    await ready

    const w = mount(MovieReviews, {
      global: {
        plugins: [i18n, router],
        stubs: { ShareImageModal: true },
      },
      props: { movieId: 603, mediaType: "movie" },
    })
    activeWrappers.push(w)
    await flushPromises()

    expect(apiMocks.getMediaReviews).toHaveBeenCalledWith("movie", 603, { page: 1 })
    expect(apiMocks.getMyReview).not.toHaveBeenCalled()
    expect(apiMocks.getFollowStatus).not.toHaveBeenCalled()

    const signInLink = w.find(".reviews__link")
    expect(parseHref(signInLink.attributes("href"))).toEqual({ path: "/login", redirect: "/movies/603" })

    const actionButtons = w.findAll(".card__btn")
    const gate = useAuthGate()
    await actionButtons[0]!.trigger("click")
    expect(gate.isOpen.value).toBe(true)
    expect(gate.reason.value).toBe("like")

    gate.close()
    await actionButtons[1]!.trigger("click")
    expect(gate.isOpen.value).toBe(true)
    expect(gate.reason.value).toBe("helpful")

    expect(apiMocks.likeReview).not.toHaveBeenCalled()
    expect(apiMocks.markHelpful).not.toHaveBeenCalled()
  })

  it("HomeView omits HomeFriendsStrip and HomeGenreCta and skips private queries for guests", async () => {
    const { router, ready } = makeRouter("/")
    await ready
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

    const w = mount(HomeView, {
      global: {
        plugins: [i18n, router, [VueQueryPlugin, { queryClient }]],
        stubs: {
          HomeHero: true,
          HomeTierRail: true,
          HomeRail: true,
          HomeFriendsStrip: { template: "<div data-test='friends-strip' />" },
          HomeGenreCta: { template: "<div data-test='genre-cta' />" },
        },
      },
    })
    activeWrappers.push(w)
    await flushPromises()

    expect(w.find('[data-test="friends-strip"]').exists()).toBe(false)
    expect(w.find('[data-test="genre-cta"]').exists()).toBe(false)
    expect(apiMocks.getFeed).not.toHaveBeenCalled()
    expect(apiMocks.getRecommended).not.toHaveBeenCalled()
  })
})
