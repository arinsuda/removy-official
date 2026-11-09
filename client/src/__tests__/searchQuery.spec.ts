import { describe, expect, it } from "vitest"
import { DEFAULT_SEARCH_STATE, mergeSearchQuery, parseSearchQuery, serializeSearchQuery } from "@/components/search/searchQuery"

describe("searchQuery", () => {
  it("falls back to defaults for unknown values", () => {
    expect(parseSearchQuery({ q: "  dune ", tab: "bogus", sort: "x", year: "abc", genre: "-3" })).toEqual({ ...DEFAULT_SEARCH_STATE, q: "dune" })
  })

  it("round-trips non-default state and omits defaults", () => {
    const state = { q: "dune", tab: "people" as const, sort: "date" as const, year: 2021, genre: 28 }
    expect(serializeSearchQuery(state)).toEqual({ q: "dune", tab: "people", sort: "date", year: "2021", genre: "28" })
    expect(parseSearchQuery(serializeSearchQuery(state))).toEqual(state)
    expect(serializeSearchQuery({ ...DEFAULT_SEARCH_STATE, q: "a" })).toEqual({ q: "a" })
  })

  it("keeps unrelated keys when merging", () => {
    expect(mergeSearchQuery({ q: "a", tab: "tv", utm: "x" }, { ...DEFAULT_SEARCH_STATE, q: "b" })).toEqual({ utm: "x", q: "b" })
  })
})
