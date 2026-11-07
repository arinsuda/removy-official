/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { isSessionRejected, runExclusiveRefresh } from "@/api/refresh"

// A minimal navigator.locks: callbacks for one name run strictly one after another.
function fakeLocks() {
  let tail: Promise<unknown> = Promise.resolve()
  return {
    request: <T>(_name: string, cb: () => Promise<T>) => {
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
