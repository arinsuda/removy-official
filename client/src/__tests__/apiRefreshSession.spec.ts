/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect, vi, beforeEach } from "vitest"
import axios, { type AxiosAdapter, type InternalAxiosRequestConfig } from "axios"

const clearUser = vi.hoisted(() => vi.fn())
const resumeSocket = vi.hoisted(() => vi.fn())
vi.mock("@/stores/auth", () => ({ useAuthStore: () => ({ clearUser }) }))
vi.mock("@/composables/useSocket", () => ({ resume: resumeSocket }))

import api from "@/api/index"

function reject401(config: InternalAxiosRequestConfig) {
  return Promise.reject({ config, response: { status: 401, data: {}, config } })
}

// Holds the refresh POST open so the test can change the session while it is in flight.
function pendingRefresh() {
  let settle!: (v: { data: unknown } | { error: unknown }) => void
  const spy = vi.spyOn(axios, "post").mockImplementation(
    () =>
      new Promise((resolve, reject) => {
        settle = (v) => ("error" in v ? reject(v.error) : resolve(v))
      }),
  )
  return { spy, settle: (v: { data: unknown } | { error: unknown }) => settle(v) }
}

describe("api refresh vs a changed session", () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    clearUser.mockReset()
    resumeSocket.mockReset()
    localStorage.clear()
    localStorage.setItem("access_token", "at-A")
    localStorage.setItem("refresh_token", "rt-A")
    api.defaults.adapter = reject401 as unknown as AxiosAdapter
  })

  async function startRefresh() {
    const p = pendingRefresh()
    const request = api.get("/movies").catch((e) => e)
    await vi.waitFor(() => expect(p.spy).toHaveBeenCalled())
    return { ...p, request }
  }

  it("stays logged out when the user logs out during the refresh", async () => {
    const { settle, request } = await startRefresh()
    localStorage.removeItem("access_token")
    localStorage.removeItem("refresh_token")
    settle({ data: { access_token: "at-A2", refresh_token: "rt-A2" } })
    await request
    expect(localStorage.getItem("access_token")).toBeNull()
    expect(localStorage.getItem("refresh_token")).toBeNull()
    expect(resumeSocket).not.toHaveBeenCalled()
    expect(clearUser).not.toHaveBeenCalled()
  })

  it("keeps account B's tokens when B logs in during the refresh", async () => {
    const { settle, request } = await startRefresh()
    localStorage.setItem("access_token", "at-B")
    localStorage.setItem("refresh_token", "rt-B")
    settle({ data: { access_token: "at-A2", refresh_token: "rt-A2" } })
    await request
    expect(localStorage.getItem("access_token")).toBe("at-B")
    expect(localStorage.getItem("refresh_token")).toBe("rt-B")
    expect(resumeSocket).not.toHaveBeenCalled()
  })

  it("a stale 401 from the old session does not clear account B", async () => {
    const { settle, request } = await startRefresh()
    localStorage.setItem("access_token", "at-B")
    localStorage.setItem("refresh_token", "rt-B")
    settle({ error: { response: { status: 401 } } })
    await request
    expect(localStorage.getItem("refresh_token")).toBe("rt-B")
    expect(localStorage.getItem("access_token")).toBe("at-B")
    expect(clearUser).not.toHaveBeenCalled()
  })
})
