/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { useAuthStore } from "@/stores/auth"

vi.mock("@/api/api", () => ({
  authApi: { refreshToken: vi.fn(), logout: vi.fn() },
  userApi: { me: vi.fn() },
}))

import { authApi, userApi } from "@/api/api"

// navigator.locks stand-in: callbacks for one name run strictly one after another.
function fakeLocks() {
  let tail: Promise<unknown> = Promise.resolve()
  return {
    request: <T>(_name: string, ...rest: unknown[]) => {
      const cb = rest[rest.length - 1] as () => Promise<T>
      const next = tail.then(cb)
      tail = next.catch(() => undefined)
      return next
    },
  }
}

const refreshReply = (n: number) =>
  ({
    data: { user: { id: 1, email: "u@x.test", is_verified: true }, access_token: `at${n}`, refresh_token: `rt${n}` },
  }) as Awaited<ReturnType<typeof authApi.refreshToken>>

describe("fetchMe refresh lock", () => {
  beforeEach(() => {
    localStorage.clear()
    localStorage.setItem("refresh_token", "rt0")
    vi.clearAllMocks()
    Object.defineProperty(navigator, "locks", { value: fakeLocks(), configurable: true })
    vi.mocked(userApi.me).mockResolvedValue({
      data: { user: { id: 1, username: "u", favorite_genres: "[]" } },
    } as Awaited<ReturnType<typeof userApi.me>>)
  })
  afterEach(() => {
    Object.defineProperty(navigator, "locks", { value: undefined, configurable: true })
  })

  it("a second tab refreshes with the token the first tab stored, not the spent one", async () => {
    let n = 0
    vi.mocked(authApi.refreshToken).mockImplementation(async (token?: string) => {
      // The server accepts each refresh token once, in order.
      if (token !== `rt${n}`) throw { response: { status: 401 } }
      await Promise.resolve()
      n++
      return refreshReply(n)
    })

    setActivePinia(createPinia())
    const tabA = useAuthStore()
    setActivePinia(createPinia())
    const tabB = useAuthStore()

    await Promise.all([tabA.fetchMe(), tabB.fetchMe()])

    expect(tabA.isLoggedIn).toBe(true)
    expect(tabB.isLoggedIn).toBe(true)
    expect(tabB.initError).toBeNull()
    expect(localStorage.getItem("refresh_token")).toBe("rt2")
  })

  it("keeps the rotated tokens when login/logout made the call stale while it waited", async () => {
    let release!: () => void
    vi.mocked(authApi.refreshToken).mockImplementation(async () => {
      await new Promise<void>((r) => (release = r))
      return refreshReply(1)
    })

    setActivePinia(createPinia())
    const store = useAuthStore()
    const pending = store.fetchMe()
    await vi.waitFor(() => expect(release).toBeTypeOf("function"))

    // Bumps the generation without touching the stored tokens.
    store.applySession({ user: { id: 2 } } as Parameters<typeof store.applySession>[0])
    release()
    await pending

    expect(localStorage.getItem("refresh_token")).toBe("rt1")
    expect(localStorage.getItem("access_token")).toBe("at1")
    expect(store.user?.id).toBe(2)
  })

  it.each([
    ["logout", () => localStorage.removeItem("refresh_token"), null],
    ["another account's login", () => localStorage.setItem("refresh_token", "other-rt"), "other-rt"],
  ])("does not overwrite the stored tokens after %s during a refresh", async (_name, change, expected) => {
    let release!: () => void
    vi.mocked(authApi.refreshToken).mockImplementation(async () => {
      await new Promise<void>((r) => (release = r))
      return refreshReply(1)
    })

    setActivePinia(createPinia())
    const store = useAuthStore()
    const pending = store.fetchMe()
    await vi.waitFor(() => expect(release).toBeTypeOf("function"))

    change()
    store.applySession({ user: { id: 2 } } as Parameters<typeof store.applySession>[0])
    release()
    await pending

    expect(localStorage.getItem("refresh_token")).toBe(expected)
    expect(store.user?.id).toBe(2)
  })

  // Tab 1's first refresh waits for the test; any later refresh answers for whatever
  // session is stored by then: B's token rotates, anything else is rejected.
  const replyB = { data: { user: { id: 2, email: "b@x.test", is_verified: true }, access_token: "at-B2", refresh_token: "rt-B2" } } as Awaited<ReturnType<typeof authApi.refreshToken>>
  const meB = { data: { user: { id: 2, username: "b", favorite_genres: "[]" } } } as Awaited<ReturnType<typeof userApi.me>>

  it.each([
    ["logs out", () => localStorage.removeItem("refresh_token"), null, false],
    ["logs in as B", () => localStorage.setItem("refresh_token", "rt-B"), "rt-B2", true],
  ])("a refresh in tab 1 never writes A's tokens after tab 2 %s", async (_n, change, expected, loggedIn) => {
    let release!: () => void
    vi.mocked(authApi.refreshToken).mockImplementation(async (token?: string) => {
      if (token === "rt0") {
        await new Promise<void>((r) => (release = r))
        return refreshReply(1)
      }
      if (token === "rt-B") return replyB
      throw { response: { status: 401 } }
    })
    vi.mocked(userApi.me).mockResolvedValue(meB)
    setActivePinia(createPinia())
    const tab1 = useAuthStore()
    const pending = tab1.fetchMe()
    await vi.waitFor(() => expect(release).toBeTypeOf("function"))

    change() // tab 2 acts on shared localStorage; tab 1's generation is untouched
    release()
    await pending

    expect(localStorage.getItem("refresh_token")).toBe(expected)
    expect(localStorage.getItem("access_token")).toBe(loggedIn ? "at-B2" : null)
    expect(tab1.isInitialized).toBe(true)
    expect(tab1.user?.id).toBe(loggedIn ? 2 : undefined)
  })

  it("a stale 401 in tab 1 does not clear the session tab 2 just created", async () => {
    let fail!: () => void
    vi.mocked(authApi.refreshToken).mockImplementation(async (token?: string) => {
      if (token === "rt-B") return replyB
      await new Promise<void>((_r, rej) => (fail = () => rej({ response: { status: 401 } })))
      return refreshReply(1)
    })
    vi.mocked(userApi.me).mockResolvedValue(meB)
    setActivePinia(createPinia())
    const tab1 = useAuthStore()
    const pending = tab1.fetchMe()
    await vi.waitFor(() => expect(fail).toBeTypeOf("function"))

    localStorage.setItem("access_token", "at-B")
    localStorage.setItem("refresh_token", "rt-B")
    fail()
    await pending

    // B's session survives and tab 1 adopts it instead of showing logged out.
    expect(localStorage.getItem("refresh_token")).toBe("rt-B2")
    expect(tab1.user?.id).toBe(2)
    expect(tab1.initError).toBeNull()
  })

  it.each([
    ["signs in as the other tab's account", "other-rt", true],
    ["ends logged out", null, false],
  ])("finishes bootstrap when another tab changes the session mid-refresh and it %s", async (_n, otherRt, loggedIn) => {
    let release!: () => void
    vi.mocked(authApi.refreshToken).mockImplementation(async (token?: string) => {
      if (token === "rt0") {
        await new Promise<void>((r) => (release = r))
        throw { response: { status: 401 } }
      }
      if (token === "other-rt") {
        return { data: { user: { id: 2, email: "b@x.test", is_verified: true }, access_token: "atB", refresh_token: "rtB" } } as Awaited<ReturnType<typeof authApi.refreshToken>>
      }
      throw { response: { status: 401 } }
    })
    vi.mocked(userApi.me).mockResolvedValue({
      data: { user: { id: 2, username: "b", favorite_genres: "[]" } },
    } as Awaited<ReturnType<typeof userApi.me>>)

    setActivePinia(createPinia())
    const store = useAuthStore()
    const pending = store.fetchMe()
    await vi.waitFor(() => expect(release).toBeTypeOf("function"))

    // Another tab: no generation bump in this tab.
    if (otherRt) localStorage.setItem("refresh_token", otherRt)
    else localStorage.removeItem("refresh_token")
    release()
    await pending

    expect(store.isInitialized).toBe(true)
    expect(store.isLoggedIn).toBe(loggedIn)
    if (loggedIn) {
      expect(store.user?.id).toBe(2)
      expect(store.initError).toBeNull()
    } else {
      expect(store.initError).toBe("unauthenticated")
    }
  })

  it("logs only the status of a failed refresh, never the error object or its tokens", async () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => undefined)
    vi.mocked(authApi.refreshToken).mockRejectedValue({
      response: { status: 503 },
      config: { data: '{"refresh_token":"SECRET-RT"}', headers: { Authorization: "Bearer SECRET-AT" } },
    })
    setActivePinia(createPinia())
    await useAuthStore().fetchMe()

    const logged = log.mock.calls.flat()
    expect(logged.every((a) => typeof a === "string" || typeof a === "number" || typeof a === "boolean" || a === undefined)).toBe(true)
    expect(JSON.stringify(logged)).not.toContain("SECRET")
    log.mockRestore()
  })

  it("reads OAuth tokens from URL hash, stores them, and cleans the URL before network calls", async () => {
    localStorage.clear()
    window.history.replaceState({}, "", "/movies?sort=top#access_token=hash-at&refresh_token=hash-rt")
    const replaceSpy = vi.spyOn(window.history, "replaceState")

    let hashDuringRefresh = "not-checked"
    let searchDuringRefresh = "not-checked"
    let accessDuringRefresh: string | null = null
    vi.mocked(authApi.refreshToken).mockImplementation(async (token?: string) => {
      hashDuringRefresh = window.location.hash
      searchDuringRefresh = window.location.search
      accessDuringRefresh = localStorage.getItem("access_token")
      expect(token).toBe("hash-rt")
      return refreshReply(1)
    })

    setActivePinia(createPinia())
    await useAuthStore().fetchMe()

    expect(accessDuringRefresh).toBe("hash-at")
    expect(hashDuringRefresh).toBe("")
    expect(searchDuringRefresh).toBe("?sort=top")
    expect(replaceSpy).toHaveBeenCalledWith({}, document.title, "/movies?sort=top")
    expect(localStorage.getItem("access_token")).toBe("at1")
    expect(localStorage.getItem("refresh_token")).toBe("rt1")

    // Also preserves any non-token hash params alongside the tokens
    window.history.replaceState({}, "", "/movies?sort=top#access_token=hash-at2&refresh_token=hash-rt2&tab=reviews")
    replaceSpy.mockClear()
    vi.mocked(authApi.refreshToken).mockImplementation(async (token?: string) => {
      expect(token).toBe("hash-rt2")
      return refreshReply(2)
    })
    await useAuthStore().fetchMe()
    expect(replaceSpy).toHaveBeenCalledWith({}, document.title, "/movies?sort=top#tab=reviews")
    replaceSpy.mockRestore()
    window.history.replaceState({}, "", "/")
  })

  it("ignores OAuth tokens in the query string without storing them, while still stripping them from the URL", async () => {
    localStorage.clear()
    window.history.replaceState({}, "", "/movies?access_token=query-at&refresh_token=query-rt&sort=top#cast")
    const replaceSpy = vi.spyOn(window.history, "replaceState")

    let accessDuringRefresh: string | null = "not-checked"
    let refreshDuringRefresh: string | null = "not-checked"
    let searchDuringRefresh = "not-checked"
    vi.mocked(authApi.refreshToken).mockImplementation(async (token?: string) => {
      accessDuringRefresh = localStorage.getItem("access_token")
      refreshDuringRefresh = localStorage.getItem("refresh_token")
      searchDuringRefresh = window.location.search
      expect(token).toBeUndefined()
      throw { response: { status: 401 } }
    })

    setActivePinia(createPinia())
    await useAuthStore().fetchMe()

    expect(accessDuringRefresh).toBeNull()
    expect(refreshDuringRefresh).toBeNull()
    expect(localStorage.getItem("access_token")).toBeNull()
    expect(localStorage.getItem("refresh_token")).toBeNull()
    expect(searchDuringRefresh).toBe("?sort=top")
    expect(replaceSpy).toHaveBeenCalledWith({}, document.title, "/movies?sort=top#cast")
    expect(window.location.search).toBe("?sort=top")
    expect(window.location.hash).toBe("#cast")
    replaceSpy.mockRestore()
    window.history.replaceState({}, "", "/")
  })
})
