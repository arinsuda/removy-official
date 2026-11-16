import { describe, expect, it } from "vitest"
import {
  DEFAULT_MANGA_STATE,
  mergeMangaQuery,
  parseMangaQuery,
  serializeMangaQuery,
  toMangaApiQuery,
  type MangaBrowseState,
} from "@/components/manga/mangaQuery"
import { MANGA_GENRE_MAX_LENGTH } from "@/components/manga/mangaConfig"

describe("parseMangaQuery", () => {
  it("returns the defaults for an empty query", () => {
    expect(parseMangaQuery({})).toEqual(DEFAULT_MANGA_STATE)
    expect(DEFAULT_MANGA_STATE).toMatchObject({ country: "JP", format: "all", sort: "popular", status: "all", genre: null })
  })

  it("reads every valid key", () => {
    expect(parseMangaQuery({ country: "KR", format: "novel", sort: "trending", genre: "Action", status: "finished" })).toEqual({
      adult: false,
      country: "KR",
      format: "novel",
      sort: "trending",
      genre: "Action",
      status: "finished",
    })
  })

  it("falls back to the default for unknown values", () => {
    expect(parseMangaQuery({ country: "FR", format: "comic", sort: "random", status: "cancelled" })).toEqual(DEFAULT_MANGA_STATE)
  })

  it("trims the genre and drops empty or overlong ones", () => {
    expect(parseMangaQuery({ genre: "  Romance  " }).genre).toBe("Romance")
    expect(parseMangaQuery({ genre: "   " }).genre).toBeNull()
    expect(parseMangaQuery({ genre: "x".repeat(MANGA_GENRE_MAX_LENGTH + 1) }).genre).toBeNull()
  })

  it("takes the first value when a key repeats", () => {
    expect(parseMangaQuery({ country: ["KR", "CN"] }).country).toBe("KR")
  })
})

describe("serializeMangaQuery and mergeMangaQuery", () => {
  it("omits values equal to their default", () => {
    expect(serializeMangaQuery(DEFAULT_MANGA_STATE)).toEqual({})
    expect(serializeMangaQuery({ ...DEFAULT_MANGA_STATE, country: "CN", genre: "Drama" })).toEqual({ country: "CN", genre: "Drama" })
  })

  it("round-trips through the URL", () => {
    const state: MangaBrowseState = { country: "all", format: "one_shot", sort: "top_rated", genre: "Horror", status: "releasing", adult: false }
    const query = serializeMangaQuery(state)
    expect(parseMangaQuery(query)).toEqual(state)
    expect(parseMangaQuery({ ...query, adult: "1" }).adult).toBe(true)
  })

  it("replaces manga keys and keeps the others", () => {
    const merged = mergeMangaQuery({ q: "x", tab: "manga", genre: "Old", country: "KR" }, DEFAULT_MANGA_STATE)
    expect(merged).toEqual({ q: "x", tab: "manga" })
  })
})

describe("adult flag in the URL", () => {
  it("reads adult=1 only and writes it only when on", () => {
    expect(parseMangaQuery({ adult: "1" }).adult).toBe(true)
    expect(parseMangaQuery({ adult: "true" }).adult).toBe(false)
    expect(serializeMangaQuery({ ...DEFAULT_MANGA_STATE, adult: true })).toEqual({ adult: "1" })
  })
})

describe("toMangaApiQuery", () => {
  it("sends the filters the endpoint takes, with genre left out when unset", () => {
    expect(toMangaApiQuery(DEFAULT_MANGA_STATE, false)).toEqual({
      country: "JP",
      format: "all",
      sort: "popular",
      status: "all",
      genre: undefined,
      include_adult: false,
    })
    expect(toMangaApiQuery({ ...DEFAULT_MANGA_STATE, genre: "Action" }, false).genre).toBe("Action")
  })

  it("passes include_adult through", () => {
    expect(toMangaApiQuery(DEFAULT_MANGA_STATE, true).include_adult).toBe(true)
  })
})
