import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { consumeUrlSecrets } from "@/utils/urlSecrets"

const SECRET = "s3cret-value"

describe("consumeUrlSecrets", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/")
  })
  afterEach(() => vi.restoreAllMocks())

  it("reads the fragment and cleans the URL", () => {
    window.history.replaceState(null, "", `/x#token=${SECRET}&uid=7`)
    expect(consumeUrlSecrets(["token", "uid"])).toEqual({ token: SECRET, uid: "7" })
    expect(window.location.href).not.toContain(SECRET)
    expect(window.location.hash).toBe("")
  })

  it("never reads the query, even next to a fragment, and strips it", () => {
    window.history.replaceState(null, "", `/x?token=old#token=${SECRET}`)
    expect(consumeUrlSecrets(["token"]).token).toBe(SECRET)
    expect(window.location.search).toBe("")
  })

  it("drops the secret from vue-router's history.state.current", () => {
    window.history.replaceState({ current: `/x#token=${SECRET}`, back: null }, "", `/x#token=${SECRET}`)
    consumeUrlSecrets(["token"])
    expect(JSON.stringify(window.history.state)).not.toContain(SECRET)
  })

  it("ignores a query-only token but still strips it", () => {
    window.history.replaceState(null, "", `/x?token=${SECRET}`)
    expect(consumeUrlSecrets(["token"])).toEqual({})
    expect(window.location.href).not.toContain(SECRET)
  })

  it("never logs the secret", () => {
    const spies = (["log", "info", "warn", "error", "debug"] as const).map((m) =>
      vi.spyOn(console, m).mockImplementation(() => {}),
    )
    window.history.replaceState(null, "", `/x#token=${SECRET}`)
    consumeUrlSecrets(["token"])
    for (const s of spies) expect(JSON.stringify(s.mock.calls)).not.toContain(SECRET)
  })
})
