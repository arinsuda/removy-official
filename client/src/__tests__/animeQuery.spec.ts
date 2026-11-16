import { describe, expect, it } from "vitest"
import {
  DEFAULT_ANIME_STATE,
  applyAnimeChange,
  mergeAnimeQuery,
  parseAnimeQuery,
  serializeAnimeQuery,
  toAnimeApiQuery,
  type AnimeBrowseState,
} from "@/components/anime/animeQuery"
import { animeDecades, animeYears } from "@/components/anime/animeConfig"

const NOW = 2026

describe("animeDecades and animeYears", () => {
  it("lists decades from the current one down to the minimum, newest first", () => {
    expect(animeDecades(NOW)).toEqual([2020, 2010, 2000, 1990, 1980, 1970, 1960])
    expect(animeDecades(1999)).toEqual([1990, 1980, 1970, 1960])
  })

  it("lists years from the current year down to the minimum", () => {
    const years = animeYears(1961)
    expect(years).toEqual([1961, 1960])
  })
})

describe("parseAnimeQuery", () => {
  it("returns defaults for an empty query", () => {
    expect(parseAnimeQuery({}, NOW)).toEqual(DEFAULT_ANIME_STATE)
  })

  it("reads valid values", () => {
    expect(
      parseAnimeQuery({ type: "tv", sort: "airing", genre: "28", keyword: "isekai", country: "all", decade: "1990" }, NOW),
    ).toEqual({ ...DEFAULT_ANIME_STATE, sort: "airing", genre: 28, keyword: "isekai", country: "all", decade: 1990 })
  })

  it("falls back to defaults for unknown values", () => {
    const s = parseAnimeQuery({ type: "book", sort: "random", keyword: "cars", country: "US", season: "monsoon" }, NOW)
    expect(s).toEqual(DEFAULT_ANIME_STATE)
  })

  it("drops the Animation genre, negative and fractional genres", () => {
    expect(parseAnimeQuery({ genre: "16" }, NOW).genre).toBeNull()
    expect(parseAnimeQuery({ genre: "-3" }, NOW).genre).toBeNull()
    expect(parseAnimeQuery({ genre: "1.5" }, NOW).genre).toBeNull()
    expect(parseAnimeQuery({ genre: "" }, NOW).genre).toBeNull()
  })

  it("rejects sorts that the type does not have", () => {
    expect(parseAnimeQuery({ type: "movie", sort: "airing" }, NOW).sort).toBe("popular")
    expect(parseAnimeQuery({ type: "tv", sort: "airing" }, NOW).sort).toBe("airing")
  })

  it("rejects decades and years outside the offered range", () => {
    expect(parseAnimeQuery({ decade: "1950" }, NOW).decade).toBeNull()
    expect(parseAnimeQuery({ decade: "2030" }, NOW).decade).toBeNull()
    expect(parseAnimeQuery({ year: "2027" }, NOW).year).toBeNull()
  })

  it("lets year win over decade", () => {
    expect(parseAnimeQuery({ decade: "1990", year: "1995" }, NOW)).toMatchObject({ decade: null, year: 1995 })
  })

  it("keeps season only for TV with a year", () => {
    expect(parseAnimeQuery({ season: "fall" }, NOW).season).toBeNull()
    expect(parseAnimeQuery({ type: "movie", year: "2020", season: "fall" }, NOW).season).toBeNull()
    expect(parseAnimeQuery({ year: "2020", season: "fall" }, NOW)).toMatchObject({ year: 2020, season: "fall" })
  })

  it("takes the first of repeated keys", () => {
    expect(parseAnimeQuery({ type: ["movie", "tv"] }, NOW).type).toBe("movie")
  })
})

describe("serializeAnimeQuery", () => {
  it("omits defaults", () => {
    expect(serializeAnimeQuery(DEFAULT_ANIME_STATE)).toEqual({})
  })

  it("round-trips non-default state", () => {
    const state: AnimeBrowseState = {
      type: "tv",
      sort: "top_rated",
      genre: 18,
      keyword: "slice_of_life",
      country: "CN",
      decade: null,
      year: 2019,
      season: "winter",
    }
    expect(parseAnimeQuery(serializeAnimeQuery(state), NOW)).toEqual(state)
  })

  it("mergeAnimeQuery keeps unrelated keys and drops stale anime keys", () => {
    const merged = mergeAnimeQuery({ sort: "old", year: "1990", ref: "x" }, DEFAULT_ANIME_STATE)
    expect(merged).toEqual({ ref: "x" })
  })
})

describe("applyAnimeChange", () => {
  it("switching type clears the genre and the TV-only season", () => {
    const tv = applyAnimeChange(DEFAULT_ANIME_STATE, { genre: 18, year: 2020, season: "spring" })
    const movie = applyAnimeChange(tv, { type: "movie" })
    expect(movie).toMatchObject({ type: "movie", genre: null, year: 2020, season: null })
  })

  it("switching type to movie moves an airing sort back to popular", () => {
    const airing = applyAnimeChange(DEFAULT_ANIME_STATE, { sort: "airing" })
    expect(applyAnimeChange(airing, { type: "movie" }).sort).toBe("popular")
  })

  it("picking a decade clears year and season", () => {
    const withYear = applyAnimeChange(DEFAULT_ANIME_STATE, { year: 2020, season: "fall" })
    expect(applyAnimeChange(withYear, { decade: 1990 })).toMatchObject({ decade: 1990, year: null, season: null })
  })

  it("picking a year clears the decade and keeps the season", () => {
    const withDecade = applyAnimeChange(DEFAULT_ANIME_STATE, { decade: 2010 })
    expect(applyAnimeChange(withDecade, { year: 2015, season: "summer" })).toMatchObject({
      decade: null,
      year: 2015,
      season: "summer",
    })
  })

  it("clearing the year clears the season", () => {
    const s = applyAnimeChange(DEFAULT_ANIME_STATE, { year: 2020, season: "fall" })
    expect(applyAnimeChange(s, { year: null })).toMatchObject({ year: null, season: null })
  })

  it("ignores a season without a year", () => {
    expect(applyAnimeChange(DEFAULT_ANIME_STATE, { season: "winter" }).season).toBeNull()
  })
})

describe("toAnimeApiQuery", () => {
  it("sends the API params and leaves out empty filters", () => {
    expect(toAnimeApiQuery(DEFAULT_ANIME_STATE)).toEqual({
      sort: "popular",
      country: "JP",
      genre: undefined,
      keyword: undefined,
      decade: undefined,
      year: undefined,
      season: undefined,
    })
    expect(toAnimeApiQuery({ ...DEFAULT_ANIME_STATE, genre: 18, year: 2020, season: "fall" })).toMatchObject({
      genre: 18,
      year: 2020,
      season: "fall",
    })
  })
})
