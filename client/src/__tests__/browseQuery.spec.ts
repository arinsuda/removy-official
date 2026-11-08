import { describe, expect, it } from "vitest"
import { DEFAULT_BROWSE_STATE, mergeBrowseQuery, parseBrowseQuery, serializeBrowseQuery } from "@/components/browse/browseQuery"

describe("parseBrowseQuery", () => {
  it("returns defaults for an empty query", () => {
    expect(parseBrowseQuery({})).toEqual(DEFAULT_BROWSE_STATE)
  })

  it("reads valid values", () => {
    expect(parseBrowseQuery({ tab: "top", genre: "28", sort: "popcorn", tier: "hot", q: " dune " })).toEqual({
      tab: "top",
      genre: 28,
      sort: "popcorn",
      tier: "hot",
      q: "dune",
    })
  })

  it("falls back to defaults for unknown values", () => {
    const s = parseBrowseQuery({ tab: "best", genre: "abc", sort: "random", tier: "gold" })
    expect(s).toEqual(DEFAULT_BROWSE_STATE)
    expect(parseBrowseQuery({ genre: "-3" }).genre).toBeNull()
    expect(parseBrowseQuery({ genre: "1.5" }).genre).toBeNull()
  })

  it("takes the first of repeated keys", () => {
    expect(parseBrowseQuery({ tab: ["now", "top"] }).tab).toBe("now")
  })
})

describe("serializeBrowseQuery", () => {
  it("omits defaults", () => {
    expect(serializeBrowseQuery(DEFAULT_BROWSE_STATE)).toEqual({})
  })

  it("round-trips non-default state", () => {
    const state = { tab: "now", genre: 35, sort: "date", tier: "stale", q: "up" } as const
    expect(parseBrowseQuery(serializeBrowseQuery(state))).toEqual(state)
  })

  it("mergeBrowseQuery keeps unrelated keys and drops stale browse keys", () => {
    const merged = mergeBrowseQuery({ q: "old", tab: "top", ref: "x" }, DEFAULT_BROWSE_STATE)
    expect(merged).toEqual({ ref: "x" })
  })
})
