import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { useShareImage } from "@/composables/useShareImage"

vi.mock("html-to-image", () => ({ toPng: vi.fn() }))

const TMDB = "https://image.tmdb.org/t/p/w500/poster.jpg"
const ANILIST = "https://s4.anilist.co/file/anilistcdn/media/manga/cover/medium/bx1.png"

function imageResponse() {
  return new Response(new Blob([new Uint8Array([137, 80, 78, 71])], { type: "image/png" }), { status: 200 })
}

describe("useShareImage.fetchImageAsBlob", () => {
  beforeEach(() => setActivePinia(createPinia()))
  afterEach(() => vi.unstubAllGlobals())

  it("loads TMDB images through the proxy only and returns a data URL", async () => {
    const fetchMock = vi.fn().mockResolvedValue(imageResponse())
    vi.stubGlobal("fetch", fetchMock)
    const out = await useShareImage().fetchImageAsBlob(TMDB)
    expect(out).toMatch(/^data:image\/png;base64,/)
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const called = String(fetchMock.mock.calls[0][0])
    expect(called).toBe(`/api/share/proxy-image?url=${encodeURIComponent(TMDB)}`)
    expect(called.startsWith("https://image.tmdb.org")).toBe(false)
  })

  it("loads AniList covers through the proxy too", async () => {
    const fetchMock = vi.fn().mockResolvedValue(imageResponse())
    vi.stubGlobal("fetch", fetchMock)
    await useShareImage().fetchImageAsBlob(ANILIST)
    expect(String(fetchMock.mock.calls[0][0])).toBe(`/api/share/proxy-image?url=${encodeURIComponent(ANILIST)}`)
  })

  it("fetches other hosts directly, including look-alike or plain-http hosts", async () => {
    const fetchMock = vi.fn().mockResolvedValue(imageResponse())
    vi.stubGlobal("fetch", fetchMock)
    const { fetchImageAsBlob } = useShareImage()
    for (const url of ["https://evil-s4.anilist.co/x.png", "http://s4.anilist.co/x.png", "https://cdn.example.com/x.png"]) {
      await fetchImageAsBlob(url)
    }
    expect(fetchMock.mock.calls.map((c) => String(c[0]))).toEqual([
      "https://evil-s4.anilist.co/x.png",
      "http://s4.anilist.co/x.png",
      "https://cdn.example.com/x.png",
    ])
  })

  it("never falls back to a direct TMDB request when the proxy fails", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response("nope", { status: 502 }))
    vi.stubGlobal("fetch", fetchMock)
    expect(await useShareImage().fetchImageAsBlob(TMDB)).toBeNull()
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it("returns null when the network throws", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")))
    expect(await useShareImage().fetchImageAsBlob(TMDB)).toBeNull()
  })

  it("passes data and blob URLs through without fetching", async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal("fetch", fetchMock)
    const { fetchImageAsBlob } = useShareImage()
    expect(await fetchImageAsBlob("data:image/png;base64,AAAA")).toBe("data:image/png;base64,AAAA")
    expect(await fetchImageAsBlob("blob:http://localhost/x")).toBe("blob:http://localhost/x")
    expect(fetchMock).not.toHaveBeenCalled()
  })
})
