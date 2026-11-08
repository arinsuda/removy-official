/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect, vi, beforeEach } from "vitest"
import type { AxiosAdapter, InternalAxiosRequestConfig } from "axios"

const runExclusiveRefresh = vi.hoisted(() => vi.fn())
vi.mock("@/api/refresh", () => ({
  REFRESH_POST_TIMEOUT_MS: 1,
  isSessionRejected: () => false,
  runExclusiveRefresh,
}))
vi.mock("@/stores/auth", () => ({ useAuthStore: () => ({ clearUser: vi.fn() }) }))

import api from "@/api/index"

function reject401(config: InternalAxiosRequestConfig) {
  return Promise.reject({ config, response: { status: 401, data: {}, config } })
}

describe("api 401 interceptor", () => {
  beforeEach(() => {
    runExclusiveRefresh.mockReset()
    api.defaults.adapter = reject401 as unknown as AxiosAdapter
  })

  it.each(["/auth/handoff", "/auth/login", "/auth/refresh", "/auth/logout"])(
    "does not try to refresh after a 401 from %s",
    async (url) => {
      await expect(api.post(url)).rejects.toBeDefined()
      expect(runExclusiveRefresh).not.toHaveBeenCalled()
    },
  )

  it("still refreshes after a 401 from a normal route", async () => {
    runExclusiveRefresh.mockRejectedValue({ response: { status: 503 } })
    await expect(api.get("/movies")).rejects.toBeDefined()
    expect(runExclusiveRefresh).toHaveBeenCalledTimes(1)
  })
})
