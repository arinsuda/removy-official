import { describe, expect, it } from "vitest"
import { handoffErrorKey } from "@/utils/handoffError"

const withStatus = (status: number) => ({ response: { status } })

describe("handoffErrorKey", () => {
  it("maps statuses to keys", () => {
    expect(handoffErrorKey(withStatus(429))).toBe("navigation.link.handoff.tooMany")
    expect(handoffErrorKey(withStatus(502))).toBe("navigation.link.handoff.emailFailed")
    expect(handoffErrorKey(withStatus(503))).toBe("navigation.link.handoff.disabled")
    expect(handoffErrorKey(withStatus(401))).toBe("navigation.link.handoff.wrong")
    expect(handoffErrorKey(withStatus(400))).toBe("navigation.link.handoff.wrong")
  })
  it("treats a missing response as a network error", () => {
    expect(handoffErrorKey(new Error("Network Error"))).toBe("navigation.link.handoff.network")
    expect(handoffErrorKey(undefined)).toBe("navigation.link.handoff.network")
  })
})
