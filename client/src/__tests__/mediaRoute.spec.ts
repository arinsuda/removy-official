import { describe, expect, it } from "vitest"
import { mediaPath, mediaRoute, mediaTypeLabelKey } from "@/utils/mediaRoute"

describe("mediaRoute", () => {
  it.each([
    ["movie", "movie-detail"],
    ["tv", "tv-detail"],
    ["manga", "manga-detail"],
  ] as const)("routes %s to the %s detail route", (type, name) => {
    expect(mediaRoute(type, 42)).toEqual({ name, params: { id: 42 } })
  })

  it.each([
    ["movie", "/movies/42"],
    ["tv", "/tv/42"],
    ["manga", "/manga/42"],
  ] as const)("builds the %s path string", (type, path) => {
    expect(mediaPath(type, 42)).toBe(path)
  })

  it("keys the type label under media.types", () => {
    expect(mediaTypeLabelKey("manga")).toBe("media.types.manga")
    expect(mediaTypeLabelKey("tv")).toBe("media.types.tv")
  })
})
