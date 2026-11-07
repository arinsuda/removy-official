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
})
