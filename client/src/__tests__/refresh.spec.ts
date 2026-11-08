/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import {
  LOCK_WAIT_MS,
  REFRESH_POST_TIMEOUT_MS,
  isSessionRejected,
  runExclusiveRefresh,
  withRefreshLock,
} from "@/api/refresh"

const post = vi.hoisted(() => vi.fn())
vi.mock("@/api/index", () => ({ default: { post } }))

// A minimal navigator.locks: callbacks for one name run strictly one after another.
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

describe("isSessionRejected", () => {
  it("logs out only for a rejected session, never for 503 or a network error", () => {
    expect(isSessionRejected({ response: { status: 401 } })).toBe(true)
    expect(isSessionRejected({ response: { status: 403 } })).toBe(true)
    expect(isSessionRejected({ response: { status: 503 } })).toBe(false)
    expect(isSessionRejected({ response: { status: 500 } })).toBe(false)
    expect(isSessionRejected(new Error("Network Error"))).toBe(false)
  })
})

describe("runExclusiveRefresh", () => {
  beforeEach(() => {
    localStorage.clear()
    localStorage.setItem("access_token", "old")
    Object.defineProperty(navigator, "locks", { value: fakeLocks(), configurable: true })
  })
  afterEach(() => {
    Object.defineProperty(navigator, "locks", { value: undefined, configurable: true })
  })

  it("lets the losing tab skip the refresh instead of reusing a spent token", async () => {
    const refresh = vi.fn(async () => {
      await Promise.resolve()
      localStorage.setItem("access_token", "new") // the winner stores the rotated pair
    })
    const [a, b] = await Promise.all([
      runExclusiveRefresh("old", refresh),
      runExclusiveRefresh("old", refresh),
    ])
    expect(refresh).toHaveBeenCalledTimes(1)
    expect([a, b].sort()).toEqual(["already-refreshed", "refreshed"])
  })

  it("still refreshes when no lock API exists", async () => {
    Object.defineProperty(navigator, "locks", { value: undefined, configurable: true })
    const refresh = vi.fn(async () => {})
    expect(await runExclusiveRefresh("old", refresh)).toBe("refreshed")
    expect(refresh).toHaveBeenCalledTimes(1)
  })
})

describe("refresh timeouts", () => {
  afterEach(() => {
    vi.restoreAllMocks()
    Object.defineProperty(navigator, "locks", { value: undefined, configurable: true })
  })

  it("stops waiting for the lock after LOCK_WAIT_MS, and that abort is not a rejected session", async () => {
    // AbortSignal.timeout is native and ignores fake timers: capture the delay, abort by hand.
    const controller = new AbortController()
    const timeoutSpy = vi.spyOn(AbortSignal, "timeout").mockReturnValue(controller.signal)
    const signal = controller.signal
    Object.defineProperty(navigator, "locks", {
      configurable: true,
      value: {
        // A lock held forever by a hung tab: only the abort signal can end the wait.
        request: (_n: string, opts: { signal: AbortSignal }) =>
          new Promise((_res, rej) => opts.signal.addEventListener("abort", () => rej(opts.signal.reason))),
      },
    })
    const waiting = withRefreshLock(async () => "never").catch((e) => e)
    expect(timeoutSpy).toHaveBeenCalledWith(LOCK_WAIT_MS)
    expect(LOCK_WAIT_MS).toBe(20000)
    expect(signal.aborted).toBe(false)
    controller.abort(new DOMException("timed out", "TimeoutError"))
    const err = await waiting
    expect(isSessionRejected(err)).toBe(false)
  })

  it("requests the lock without a signal when AbortSignal.timeout is missing", async () => {
    const original = AbortSignal.timeout
    Object.defineProperty(AbortSignal, "timeout", { value: undefined, configurable: true, writable: true })
    try {
      const request = vi.fn(async (_n: string, _o: unknown, cb: () => Promise<string>) => cb())
      Object.defineProperty(navigator, "locks", { configurable: true, value: { request } })
      await expect(withRefreshLock(async () => "ok")).resolves.toBe("ok")
      expect(request.mock.calls[0][1]).toEqual({})
    } finally {
      Object.defineProperty(AbortSignal, "timeout", { value: original, configurable: true, writable: true })
    }
  })

  it("sends the refresh POST with a timeout, and a timeout is not a rejected session", async () => {
    post.mockResolvedValue({})
    const { authApi } = await import("@/api/endpoints/auth")
    await authApi.refreshToken("rt")
    expect(post).toHaveBeenCalledWith("/auth/refresh", { refresh_token: "rt" }, { timeout: REFRESH_POST_TIMEOUT_MS })
    expect(REFRESH_POST_TIMEOUT_MS).toBe(15000)
    expect(isSessionRejected({ code: "ECONNABORTED", message: "timeout of 15000ms exceeded" })).toBe(false)
  })

  it("logout sends the stored refresh token in the body", async () => {
    post.mockResolvedValue({})
    localStorage.setItem("refresh_token", "rt-stored")
    const { authApi } = await import("@/api/endpoints/auth")
    await authApi.logout()
    expect(post).toHaveBeenCalledWith("/auth/logout", { refresh_token: "rt-stored" })
    localStorage.clear()
  })
})
