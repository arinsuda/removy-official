/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect, vi } from "vitest"
import { readFileSync } from "node:fs"
import { join } from "node:path"
import type { AxiosAdapter, InternalAxiosRequestConfig } from "axios"

vi.mock("@/stores/auth", () => ({ useAuthStore: () => ({ clearUser: vi.fn() }) }))

import api, { CSRF_HEADERS } from "@/api/index"

describe("CSRF header", () => {
  it("is sent on every unsafe method, including multipart uploads", async () => {
    const seen: InternalAxiosRequestConfig[] = []
    api.defaults.adapter = ((config: InternalAxiosRequestConfig) => {
      seen.push(config)
      return Promise.resolve({ data: {}, status: 200, statusText: "OK", headers: {}, config })
    }) as unknown as AxiosAdapter

    await api.post("/x", {})
    await api.put("/x", {})
    await api.patch("/x", new FormData())
    await api.delete("/x")

    expect(seen).toHaveLength(4)
    for (const config of seen) {
      expect(config.headers.get("X-Requested-With")).toBe("XMLHttpRequest")
    }
  })

  it("is applied to the bare axios refresh call", () => {
    const src = readFileSync(join(__dirname, "../api/index.ts"), "utf8")
    expect(src).toMatch(/axios\.post\([^)]*headers: CSRF_HEADERS/s)
    expect(CSRF_HEADERS["X-Requested-With"]).toBe("XMLHttpRequest")
  })
})
