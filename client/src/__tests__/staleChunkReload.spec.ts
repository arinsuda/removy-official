import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import {
  STALE_CHUNK_RELOAD_COOLDOWN_MS,
  isStaleChunkError,
  reloadForStaleChunk,
} from "@/monitoring/staleChunkReload"

describe("staleChunkReload", () => {
  let assign: ReturnType<typeof vi.fn>

  beforeEach(() => {
    sessionStorage.clear()
    assign = vi.fn()
    vi.spyOn(window, "location", "get").mockReturnValue({ ...window.location, assign } as Location)
  })
  afterEach(() => vi.restoreAllMocks())

  it("recognises stale chunk errors only", () => {
    expect(isStaleChunkError(new Error("Unable to preload CSS for /assets/LoginView-x.css"))).toBe(true)
    expect(isStaleChunkError(new TypeError("Failed to fetch dynamically imported module: /assets/a.js"))).toBe(true)
    expect(isStaleChunkError(new Error("Request failed with status code 500"))).toBe(false)
  })

  it("reloads once, then waits for the cooldown", () => {
    const start = 1_000_000
    expect(reloadForStaleChunk("/login", start)).toBe(true)
    expect(reloadForStaleChunk("/login", start + 1)).toBe(false)
    expect(reloadForStaleChunk("/login", start + STALE_CHUNK_RELOAD_COOLDOWN_MS)).toBe(true)
    expect(assign).toHaveBeenCalledTimes(2)
    expect(assign).toHaveBeenCalledWith("/login")
  })
})
