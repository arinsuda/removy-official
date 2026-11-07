import { defineStore } from "pinia"
import { ref, computed } from "vue"

import { authApi, userApi } from "@/api/api"
import { withRefreshLock } from "@/api/refresh"
import { resume as resumeSocket } from "@/composables/useSocket"

import type { AuthResponse, AuthUser, LoginRequest, RegisterRequest } from "@/types"

export type InitErrorKind = "unauthenticated" | "network" | "server" | null

export const useAuthStore = defineStore("auth", () => {
  const user = ref<AuthUser | null>(null)

  const isLoading = ref(false)

  const isInitialized = ref(false)

  const initError = ref<InitErrorKind>(null)

  const isLoggedIn = computed(() => !!user.value)

  const isAdmin = computed(() => user.value?.role === "admin")

  // --- Stale-request protection ---
  let _generation = 0

  // --- Request deduplication (non-reactive private variable) ---
  let _initPromise: Promise<void> | null = null

  function setUser(u: AuthUser) {
    user.value = u
  }

  function clearUser() {
    user.value = null
    localStorage.removeItem("access_token")
    localStorage.removeItem("refresh_token")
  }

  async function login(data: LoginRequest) {
    isLoading.value = true

    try {
      const res = await authApi.login(data)
      if (res.data.access_token) {
        localStorage.setItem("access_token", res.data.access_token)
      }
      if (res.data.refresh_token) {
        localStorage.setItem("refresh_token", res.data.refresh_token)
      }

      _generation++
      setUser(res.data.user)
      isInitialized.value = true
      initError.value = null

      return res.data.user
    } finally {
      isLoading.value = false
    }
  }

  const needsOnboarding = computed(
    () => isLoggedIn.value && !user.value?.favorite_genres,
  )

  async function register(data: RegisterRequest) {
    isLoading.value = true

    try {
      const res = await authApi.register(data)
      if (res.data.access_token) {
        localStorage.setItem("access_token", res.data.access_token)
      }
      if (res.data.refresh_token) {
        localStorage.setItem("refresh_token", res.data.refresh_token)
      }

      _generation++
      setUser(res.data.user)
      isInitialized.value = true
      initError.value = null

      return res.data.user
    } finally {
      isLoading.value = false
    }
  }

  /** Adopt a session issued by the server outside the login form (e.g. account handoff). */
  function applySession(data: AuthResponse) {
    if (data.access_token) localStorage.setItem("access_token", data.access_token)
    if (data.refresh_token) localStorage.setItem("refresh_token", data.refresh_token)
    resumeSocket()
    _generation++
    setUser(data.user)
    isInitialized.value = true
    initError.value = null
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      _generation++
      clearUser()
      isInitialized.value = true
      initError.value = null
    }
  }

  async function fetchMe() {
    const myGen = ++_generation
    initError.value = null
    try {
      const urlParams = new URLSearchParams(window.location.search)
      const urlAccess = urlParams.get("access_token")
      const urlRefresh = urlParams.get("refresh_token")
      console.log(
        "[fetchMe] urlAccess:",
        !!urlAccess,
        "urlRefresh:",
        !!urlRefresh,
      )

      if (urlAccess && urlRefresh) {
        localStorage.setItem("access_token", urlAccess)
        localStorage.setItem("refresh_token", urlRefresh)
        window.history.replaceState(
          {},
          document.title,
          window.location.pathname,
        )
      }

      // Read the refresh token inside the lock: another tab may have rotated it while
      // we waited, and replaying the spent one would end the session everywhere.
      // Store the rotated tokens inside the lock: the server already spent the old refresh
      // token. A stale call may store them only while that token is still the current one,
      // so it can never undo a logout or overwrite another account's login.
      const refreshRes = await withRefreshLock(async () => {
        const storedRefresh = localStorage.getItem("refresh_token")
        console.log("[fetchMe] storedRefresh exists:", !!storedRefresh)
        const res = await authApi.refreshToken(storedRefresh || undefined)
        const sessionUnchanged =
          myGen === _generation ||
          (storedRefresh !== null && localStorage.getItem("refresh_token") === storedRefresh)
        if (sessionUnchanged) {
          if (res.data.access_token) localStorage.setItem("access_token", res.data.access_token)
          if (res.data.refresh_token) localStorage.setItem("refresh_token", res.data.refresh_token)
          resumeSocket()
        }
        return res
      })

      // Stale-request protection: bail out if login/logout happened while waiting
      if (myGen !== _generation) return

      const userId = refreshRes.data.user.id

      const meRes = await userApi.me(userId)

      // Stale-request protection after second network call
      if (myGen !== _generation) return

      const profile = meRes.data.user

      setUser({
        ...profile,
        email: refreshRes.data.user.email,
        is_verified: refreshRes.data.user.is_verified,
      })
      isInitialized.value = true
    } catch (e: unknown) {
      // Stale-request protection
      if (myGen !== _generation) return

      const axiosErr = e as { response?: { status?: number }; code?: string }
      const status = axiosErr.response?.status
      // Status only: the error object carries request headers and bodies with tokens.
      console.log("[fetchMe] caught error, status:", status)
      if (status === 401 || status === 403) {
        clearUser()
        initError.value = "unauthenticated"
      } else if (
        !axiosErr.response ||
        axiosErr.code === "ERR_NETWORK" ||
        axiosErr.code === "ECONNABORTED"
      ) {
        initError.value = "network"
      } else if (status !== undefined && status >= 500) {
        initError.value = "server"
      } else {
        clearUser()
        initError.value = "unauthenticated"
      }
      isInitialized.value = true
    }
  }

  /**
   * Shared entry point for authentication initialization.
   * - Deduplicates concurrent calls (returns the same promise).
   * - Treats 401 as final (no retry).
   * - Allows retry after network or server errors.
   * - Clears _initPromise in finally so future retries can proceed.
   */
  async function initialize(): Promise<void> {
    if (
      isInitialized.value &&
      initError.value !== "network" &&
      initError.value !== "server"
    ) {
      return
    }

    if (_initPromise) return _initPromise

    _initPromise = fetchMe().finally(() => {
      _initPromise = null
    })

    return _initPromise
  }

  return {
    user,
    isLoading,
    isInitialized,
    initError,
    isLoggedIn,
    isAdmin,
    needsOnboarding,
    setUser,
    clearUser,
    login,
    register,
    applySession,
    logout,
    fetchMe,
    initialize,
  }
})
