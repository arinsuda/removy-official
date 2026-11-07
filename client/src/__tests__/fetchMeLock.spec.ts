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
})
