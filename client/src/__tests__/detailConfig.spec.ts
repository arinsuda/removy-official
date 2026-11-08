import { describe, expect, it, vi } from "vitest"
import { adaptMovieDetail, adaptTvDetail, statusKey } from "@/components/detail/detailConfig"
import type { MovieDetail, TVSeriesDetail } from "@/types/movie"

vi.mock("@/api/api", () => ({ movieApi: {} }))

const series = {
  id: 1399,
  name: "Game of Thrones",
  original_name: "Game of Thrones",
  english_title: "Game of Thrones",
  first_air_date: "2011-04-17",
  last_air_date: "2019-05-19",
  overview: "Noble families fight.",
  poster_path: "/p.jpg",
  backdrop_path: "/b.jpg",
  vote_average: 8.4,
  status: "Ended",
  number_of_seasons: 8,
  number_of_episodes: 73,
  episode_run_time: [0, 60],
  genres: [{ id: 18, name: "Drama" }],
  networks: [{ id: 49, name: "HBO", logo_path: "/h.png" }],
  created_by: [{ id: 9813, name: "David Benioff", profile_path: "/d.jpg" }],
  seasons: [
    { id: 2, name: "Season 1", season_number: 1, episode_count: 10, air_date: "2011-04-17", poster_path: "", overview: "" },
    { id: 1, name: "Specials", season_number: 0, episode_count: 5, air_date: "", poster_path: "", overview: "" },
  ],
} as unknown as TVSeriesDetail

describe("adaptTvDetail", () => {
  const d = adaptTvDetail({ series })

  it("uses name and first_air_date, not the movie fields", () => {
    expect(d.title).toBe("Game of Thrones")
    expect(d.year).toBe("2011")
    expect(d.type).toBe("tv")
  })

  it("derives runtime, counts, creators and seasons", () => {
    expect(d.runtimeMin).toBe(60)
    expect([d.seasonCount, d.episodeCount]).toEqual([8, 73])
    expect(d.crewHighlights).toEqual([{ id: 9813, name: "David Benioff", role: "creator", profilePath: "/d.jpg" }])
    expect(d.tvSeasons.map(s => s.name)).toEqual(["Season 1", "Specials"])
  })

  it("lists facts that have values only", () => {
    expect(d.facts.map(f => f.key)).toEqual(["status", "firstAirDate", "lastAirDate", "networks", "seasons", "episodes"])
  })
})

describe("adaptMovieDetail", () => {
  const movie = {
    id: 27205,
    title: "Inception",
    original_title: "Inception",
    release_date: "2010-07-15",
    runtime: 148,
    status: "Released",
    budget: 160000000,
    revenue: 0,
    genres: [],
  } as unknown as MovieDetail

  it("adapts a movie and caps cast at 12", () => {
    const cast = Array.from({ length: 20 }, (_, i) => ({ id: i, name: `A${i}`, character: "", profile_path: "", order: i }))
    const crew = [
      { id: 1, name: "Chris", job: "Director", department: "Directing", profile_path: "" },
      { id: 1, name: "Chris", job: "Director", department: "Directing", profile_path: "" },
    ]
    const d = adaptMovieDetail({ movie, credits: { cast, crew } })
    expect(d.year).toBe("2010")
    expect(d.runtimeMin).toBe(148)
    expect(d.cast).toHaveLength(12)
    expect(d.crewHighlights).toHaveLength(1)
    expect(d.facts.map(f => f.key)).toEqual(["status", "releaseDate", "budget"])
    expect(d.facts.find(f => f.key === "budget")?.value).toBe("$160,000,000")
  })

  it("normalizes status for i18n keys", () => {
    expect(statusKey("Post Production")).toBe("PostProduction")
  })
})
