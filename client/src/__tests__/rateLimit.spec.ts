import { describe, it, expect } from "vitest"
import { isRateLimited } from "@/utils/rateLimit"

describe("isRateLimited", () => {
  it("detects 429 and the RATE_LIMITED code", () => {
    expect(isRateLimited({ response: { status: 429 } })).toBe(true)
    expect(isRateLimited({ response: { status: 400, data: { code: "RATE_LIMITED" } } })).toBe(true)
  })
  it("ignores other errors", () => {
    expect(isRateLimited({ response: { status: 401 } })).toBe(false)
    expect(isRateLimited(new Error("x"))).toBe(false)
    expect(isRateLimited(undefined)).toBe(false)
  })
})
