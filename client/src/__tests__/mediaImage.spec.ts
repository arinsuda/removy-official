import { describe, expect, it } from "vitest"
import { getTmdbImageUrl, getTmdbImageUrlOrPlaceholder, mediaImage } from "@/utils/image"

describe("mediaImage", () => {
  it("prefixes a TMDB path with the TMDB base and size", () => {
    expect(mediaImage("/abc.jpg", "w500")).toBe("https://image.tmdb.org/t/p/w500/abc.jpg")
  })

  it("passes https URLs on an allowed host through unchanged", () => {
    const cover = "https://s4.anilist.co/file/cover.jpg"
    expect(mediaImage(cover, "w92")).toBe(cover)
  })

  it("drops absolute URLs that are not https on an allowed host", () => {
    expect(mediaImage("http://s4.anilist.co/file/cover.jpg")).toBeNull()
    expect(mediaImage("https://evil.example.com/cover.jpg")).toBeNull()
    expect(mediaImage("https://s4.anilist.co.evil.example/cover.jpg")).toBeNull()
    expect(mediaImage("https://image.tmdb.org/t/p/w500/abc.jpg")).toBe("https://image.tmdb.org/t/p/w500/abc.jpg")
  })

  it("returns null when there is no image", () => {
    expect(mediaImage(null)).toBeNull()
    expect(mediaImage("")).toBeNull()
  })

  it("sends a disallowed absolute URL to the placeholder", () => {
    expect(getTmdbImageUrlOrPlaceholder("https://evil.example.com/x.jpg")).toBe("/placeholder.jpg")
  })

  it("keeps the TMDB helpers in step for absolute URLs", () => {
    const cover = "https://s4.anilist.co/file/cover.jpg"
    expect(getTmdbImageUrl(cover, "w342")).toBe(cover)
    expect(getTmdbImageUrlOrPlaceholder(cover)).toBe(cover)
    expect(getTmdbImageUrlOrPlaceholder(null)).toBe("/placeholder.jpg")
  })
})
