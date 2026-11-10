/**
 * @vitest-environment happy-dom
 */

import { describe, it, expect, vi, beforeEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { mount } from "@vue/test-utils"
import { createRouter, createWebHistory, type Router } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import App from "@/App.vue"
import SessionBootstrapView from "@/components/common/SessionBootstrapView.vue"

// ── Mocks ────────────────────────────────────────────────────────────────────

vi.mock("@/api/api", () => ({
  authApi: {
    refreshToken: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    getGoogleConfig: vi.fn(),
    getGoogleLoginUrl: vi.fn(() => "#"),
  },
  userApi: {
    me: vi.fn(),
  },
}))

vi.mock("gsap", () => ({
  gsap: {
    timeline: () => ({
      to: vi.fn(function (this: unknown) {
        return this
      }),
    }),
    set: vi.fn(),
    to: vi.fn(() => Promise.resolve()),
    fromTo: vi.fn(),
  },
}))

import { authApi, userApi } from "@/api/api"

// ── Helper: create a test router matching production guard ────────────────────

function createTestRouter(authStore: ReturnType<typeof useAuthStore>): Router {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      {
        path: "/login",
        name: "login",
        component: { template: '<div class="login-view">LoginView</div>' },
        meta: { guestOnly: true },
      },
      {
        path: "/register",
        name: "register",
        component: { template: '<div class="register-view">RegisterView</div>' },
        meta: { guestOnly: true },
      },
      {
        path: "/",
        name: "home",
        component: { template: '<div class="home-view">HomeView</div>' },
        meta: { requiresAuth: true },
      },
    ],
  })

  router.beforeEach(async (to) => {
    await authStore.initialize()

    if (authStore.initError === "network" || authStore.initError === "server") {
      return
    }

    const needsAuth = to.matched.some((r) => r.meta.requiresAuth)

    if (needsAuth && !authStore.isLoggedIn) {
      return { name: "login", query: { redirect: to.fullPath } }
    }

    if (to.meta.guestOnly && authStore.isLoggedIn) {
      return { name: "home" }
    }
  })

  return router
}

const dummyI18nPlugin = {
  install(app: { config: { globalProperties: Record<string, unknown> } }) {
    app.config.globalProperties.$t = (key: string) => {
      const translations: Record<string, string> = {
        "session.checking": "Checking your session…",
        "session.error": "Unable to reach server",
        "session.errorDetail": "The server is not responding. Please try again.",
        "session.retry": "Retry",
      }
      return translations[key] || key
    }
  },
}

// ── Tests ────────────────────────────────────────────────────────────────────

describe("Session Bootstrap and Authorization", () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    localStorage.clear()
    vi.clearAllMocks()
  })

  describe("Vue Mounting & Initial Loading UI", () => {
    it("Vue mounts immediately before session API resolves", () => {
      const authStore = useAuthStore()
      vi.mocked(authApi.refreshToken).mockReturnValue(new Promise(() => {}))

      const wrapper = mount(App, {
        global: {
          plugins: [pinia, dummyI18nPlugin],
          stubs: { RouterView: true },
        },
      })

      expect(wrapper.exists()).toBe(true)
      expect(authStore.isInitialized).toBe(false)
    })

    it("loading bootstrap UI is visible while pending", () => {
      const authStore = useAuthStore()
      authStore.isInitialized = false

      const wrapper = mount(App, {
        global: {
          plugins: [pinia, dummyI18nPlugin],
          stubs: { RouterView: true },
        },
      })

      expect(wrapper.findComponent(SessionBootstrapView).exists()).toBe(true)
      expect(wrapper.text()).toContain("Checking your session…")
    })
  })

  describe("Session Verification & Routing Outcomes", () => {
    it("valid session routes to Home", async () => {
      const authStore = useAuthStore()

      vi.mocked(authApi.refreshToken).mockResolvedValueOnce({
        data: {
          user: { id: 1, email: "user@test.com" },
          access_token: "at",
          refresh_token: "rt",
        },
      } as Awaited<ReturnType<typeof authApi.refreshToken>>)

      vi.mocked(userApi.me).mockResolvedValueOnce({
        data: { user: { id: 1, username: "validUser", favorite_genres: '["action"]' } },
      } as Awaited<ReturnType<typeof userApi.me>>)

      const router = createTestRouter(authStore)
      router.push("/")
      await router.isReady()

      expect(authStore.isLoggedIn).toBe(true)
      expect(router.currentRoute.value.path).toBe("/")
    })

    it("expired session routes to Login", async () => {
      const authStore = useAuthStore()

      vi.mocked(authApi.refreshToken).mockRejectedValueOnce({
        response: { status: 401 },
      })

      const router = createTestRouter(authStore)
      router.push("/")
      await router.isReady()

      expect(authStore.isLoggedIn).toBe(false)
      expect(router.currentRoute.value.name).toBe("login")
    })

    it("no LoginView flash occurs for an authenticated user opening /login directly", async () => {
      const authStore = useAuthStore()

      // Delay token resolution slightly
      let resolveRefresh!: (val: unknown) => void
      vi.mocked(authApi.refreshToken).mockReturnValue(
        new Promise((r) => {
          resolveRefresh = r
        }),
      )
      vi.mocked(userApi.me).mockResolvedValue({
        data: { user: { id: 1, username: "user", favorite_genres: '["action"]' } },
      } as Awaited<ReturnType<typeof userApi.me>>)

      const router = createTestRouter(authStore)
      router.push("/login")

      const wrapper = mount(App, {
        global: {
          plugins: [pinia, router, dummyI18nPlugin],
        },
      })

      // While session check is pending, App displays SessionBootstrapView (NOT LoginView!)
      expect(wrapper.findComponent(SessionBootstrapView).exists()).toBe(true)
      expect(wrapper.find(".login-view").exists()).toBe(false)

      // Resolve session check as logged in
      resolveRefresh({
        data: {
          user: { id: 1, email: "user@test.com" },
          access_token: "at",
          refresh_token: "rt",
        },
      })

      await router.isReady()
      await wrapper.vm.$nextTick()

      // Router redirects to HomeView without showing LoginView
      expect(router.currentRoute.value.path).toBe("/")
      expect(wrapper.find(".login-view").exists()).toBe(false)
    })
  })

  describe("Backend Error & Retry Behavior", () => {
    it("backend error shows Retry state", async () => {
      const authStore = useAuthStore()

      vi.mocked(authApi.refreshToken).mockRejectedValueOnce({
        code: "ERR_NETWORK",
      })

      const wrapper = mount(SessionBootstrapView, {
        global: {
          plugins: [pinia, dummyI18nPlugin],
        },
      })

      await authStore.initialize()
      await wrapper.vm.$nextTick()

      expect(authStore.initError).toBe("network")
      expect(wrapper.text()).toContain("Unable to reach server")
      expect(wrapper.find(".btn-retry").exists()).toBe(true)
    })

    it("clicking Retry triggers session initialization again", async () => {
      const authStore = useAuthStore()

      vi.mocked(authApi.refreshToken).mockRejectedValueOnce({
        code: "ERR_NETWORK",
      })

      await authStore.initialize()
      expect(authStore.initError).toBe("network")

      const wrapper = mount(SessionBootstrapView, {
        global: {
          plugins: [pinia, dummyI18nPlugin],
        },
      })

      // Mock successful response for retry
      vi.mocked(authApi.refreshToken).mockResolvedValueOnce({
        data: {
          user: { id: 1, email: "user@test.com" },
          access_token: "at",
          refresh_token: "rt",
        },
      } as Awaited<ReturnType<typeof authApi.refreshToken>>)

      vi.mocked(userApi.me).mockResolvedValueOnce({
        data: { user: { id: 1, username: "retryUser" } },
      } as Awaited<ReturnType<typeof userApi.me>>)

      const retryBtn = wrapper.find(".btn-retry")
      await retryBtn.trigger("click")

      expect(authStore.isLoggedIn).toBe(true)
      expect(authStore.initError).toBeNull()
    })
  })

  describe("URL fragment secrets", () => {
    it("initialize() leaves #token= intact for the verify/reset pages", async () => {
      const authStore = useAuthStore()
      window.history.replaceState(null, "", "/auth/verify-email#token=KEEPME&access_token=a&refresh_token=r")
      vi.mocked(authApi.refreshToken).mockRejectedValue({ response: { status: 401 } })
      await authStore.initialize()
      const params = new URLSearchParams(window.location.hash.slice(1))
      expect(params.get("token")).toBe("KEEPME")
      expect(params.has("access_token")).toBe(false)
      window.history.replaceState(null, "", "/")
    })
  })

  describe("Deduplication & Staleness Protection", () => {
    it("concurrent initialization sends one request", async () => {
      const authStore = useAuthStore()
      let resolveRefresh!: (val: unknown) => void

      vi.mocked(authApi.refreshToken).mockReturnValue(
        new Promise((r) => {
          resolveRefresh = r
        }),
      )
      vi.mocked(userApi.me).mockResolvedValue({
        data: { user: { id: 1, username: "test" } },
      } as Awaited<ReturnType<typeof userApi.me>>)

      const p1 = authStore.initialize()
      const p2 = authStore.initialize()
      const p3 = authStore.initialize()

      resolveRefresh({
        data: {
          user: { id: 1, email: "test@test.com" },
          access_token: "at",
          refresh_token: "rt",
        },
      })

      await Promise.all([p1, p2, p3])

      expect(authApi.refreshToken).toHaveBeenCalledTimes(1)
    })

    it("stale initialization cannot overwrite a newer login", async () => {
      const authStore = useAuthStore()

      let resolveSlowFetch!: (val: unknown) => void
      vi.mocked(authApi.refreshToken).mockReturnValue(
        new Promise((r) => {
          resolveSlowFetch = r
        }),
      )

      // Start slow background initialize
      const slowPromise = authStore.initialize()

      // User performs explicit login while fetch is pending
      vi.mocked(authApi.login).mockResolvedValueOnce({
        data: {
          user: { id: 99, username: "newlyLoggedInUser" },
          access_token: "new_at",
          refresh_token: "new_rt",
        },
      } as Awaited<ReturnType<typeof authApi.login>>)

      await authStore.login({ identifier: "new", password: "pwd" })

      expect(authStore.user?.username).toBe("newlyLoggedInUser")

      // Now slow fetch resolves late
      resolveSlowFetch({
        data: {
          user: { id: 1, email: "old@test.com" },
          access_token: "old_at",
          refresh_token: "old_rt",
        },
      })

      await slowPromise

      // Newly logged in user must NOT be overwritten by the stale response
      expect(authStore.user?.username).toBe("newlyLoggedInUser")
      expect(authStore.user?.id).toBe(99)
    })
  })
})
