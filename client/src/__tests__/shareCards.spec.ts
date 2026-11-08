import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createI18n } from "vue-i18n"
import en from "@/i18n/locales/en/share.json"
import ShareImageModal from "@/components/share/ShareImageModal.vue"
import TakeCard from "@/components/share/cards/TakeCard.vue"
import PosterStoryCard from "@/components/share/cards/PosterStoryCard.vue"
import { availableTemplates, publicExcerpt, shareTier, truncateExcerpt } from "@/components/share/shareCards"
import { useShareImage } from "@/composables/useShareImage"
import type { ShareMediaContext, ShareReviewContext, ShareStatsContext } from "@/types/share"

vi.mock("html-to-image", () => ({ toPng: vi.fn().mockResolvedValue("data:image/png;base64,AAAA") }))

const i18n = createI18n({
  legacy: false,
  locale: "en",
  messages: { en: { share: en, common: { close: "Close" } } },
})

const media: ShareMediaContext = {
  id: 7, mediaType: "movie", title: "Dune", posterPath: "/p.jpg", posterUrl: null,
  releaseYear: "2021", genres: [{ id: 878, name: "Sci-Fi" }], voteAverage: 8, popcornRating: 4.4, popcornCount: 12,
}
const stats: ShareStatsContext = {
  rating: 4.5, watchCount: 2, lastWatchedOn: "2026-10-12", reviewBody: "Secret thoughts", reviewVisibility: "private",
}
const review: ShareReviewContext = {
  id: 1, authorDisplayName: "A", authorUsername: "a", authorAvatarUrl: null, rating: 4, body: "Great", visibility: "public", createdAt: "",
}

describe("template selection", () => {
  it("offers My take only when the owner has stats", () => {
    expect(availableTemplates("poster", null, null).map((d) => d.key)).toEqual(["poster", "sticker", "square"])
    expect(availableTemplates("poster", stats, null).map((d) => d.key)).toEqual(["poster", "take", "sticker", "square"])
    expect(availableTemplates("poster", { ...stats, rating: null, watchCount: 0 }, null).map((d) => d.key)).not.toContain("take")
  })

  it("offers the review card only for public reviews", () => {
    expect(availableTemplates("review", null, review)[0]!.key).toBe("review")
    expect(availableTemplates("review", null, { ...review, visibility: "followers" }).map((d) => d.key)).toEqual(["poster", "square"])
    expect(availableTemplates("review", null, { ...review, visibility: "private" }).map((d) => d.key)).not.toContain("review")
  })

  it("exports sticker transparent and stories at 1080x1920", () => {
    const list = availableTemplates("poster", stats, null)
    expect(list.find((d) => d.key === "sticker")).toMatchObject({ transparent: true, width: 1080, height: 1080 })
    expect(list.find((d) => d.key === "poster")).toMatchObject({ transparent: false, width: 1080, height: 1920 })
    expect(list.find((d) => d.key === "square")).toMatchObject({ width: 1080, height: 1080 })
  })

  it("derives the tier from the rounded score", () => {
    expect(shareTier(media)).toBe("hot")
    expect(shareTier({ ...media, popcornRating: 3, popcornCount: 3 })).toBe("fresh")
    expect(shareTier({ ...media, popcornRating: null, popcornCount: null })).toBe("none")
  })
})

describe("privacy", () => {
  it("publicExcerpt only returns public text", () => {
    expect(publicExcerpt("public", " hi ")).toBe("hi")
    expect(publicExcerpt("followers", "hi")).toBeNull()
    expect(publicExcerpt("private", "hi")).toBeNull()
    expect(publicExcerpt(null, "hi")).toBeNull()
    expect(publicExcerpt("public", "   ")).toBeNull()
  })

  it("truncateExcerpt cuts on a word boundary and adds an ellipsis", () => {
    expect(truncateExcerpt("short text", 90)).toBe("short text")
    const long = "Visually stunning and emotionally huge. The sandworm ride alone is worth the ticket, and the score gave me goosebumps."
    const out = truncateExcerpt(long, 90)
    expect(out.endsWith("…")).toBe(true)
    expect(Array.from(out).length).toBeLessThanOrEqual(91)
    expect(out).toBe("Visually stunning and emotionally huge. The sandworm ride alone is worth the ticket, and…")
    expect(truncateExcerpt("ก".repeat(200), 90)).toBe(`${"ก".repeat(90)}…`)
  })

  it("My take card shows an ellipsis for long public reviews", () => {
    const w = mount(TakeCard, {
      props: { media, stats: { ...stats, reviewVisibility: "public", reviewBody: "word ".repeat(60) }, posterUrl: null, theme: "dark", shortUrl: "x/y" },
      global: { plugins: [i18n] },
    })
    expect(w.find("[data-test=share-excerpt]").text().endsWith("…")).toBe(true)
  })

  it("My take card never prints non-public review text", () => {
    for (const visibility of ["private", "followers"] as const) {
      const w = mount(TakeCard, {
        props: { media, stats: { ...stats, reviewVisibility: visibility }, posterUrl: null, theme: "dark", shortUrl: "x/y" },
        global: { plugins: [i18n] },
      })
      expect(w.find("[data-test=share-excerpt]").exists()).toBe(false)
      expect(w.text()).not.toContain("Secret thoughts")
    }
    const pub = mount(TakeCard, {
      props: { media, stats: { ...stats, reviewVisibility: "public" }, posterUrl: null, theme: "dark", shortUrl: "x/y" },
      global: { plugins: [i18n] },
    })
    expect(pub.find("[data-test=share-excerpt]").text()).toBe("Secret thoughts")
    expect(pub.find("[data-test=share-watch-count]").text()).toBe("×2")
    expect(pub.find("[data-test=share-my-rating]").exists()).toBe(true)
  })

  it("poster card has no user stats", () => {
    const w = mount(PosterStoryCard, {
      props: { media, posterUrl: null, theme: "light", shortUrl: "x/y" },
      global: { plugins: [i18n] },
    })
    expect(w.find("[data-test=share-score]").text()).toContain("4.4")
    expect(w.classes()).toContain("sc-theme-light")
  })
})

describe("modal", () => {
  beforeEach(() => setActivePinia(createPinia()))
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ""
  })

  function stubProxy(ok: boolean) {
    const fetchMock = vi.fn().mockResolvedValue(
      ok ? new Response(new Blob([new Uint8Array([1, 2])], { type: "image/jpeg" })) : new Response("x", { status: 502 }),
    )
    vi.stubGlobal("fetch", fetchMock)
    return fetchMock
  }

  it("shows a skeleton, then the carousel, and only fetches through the proxy", async () => {
    const fetchMock = stubProxy(true)
    const w = mount(ShareImageModal, {
      props: { modelValue: true, shareType: "poster", media, stats: null },
      global: { plugins: [i18n] },
      attachTo: document.body,
    })
    expect(document.body.querySelector(".skeleton-block")).not.toBeNull()
    await vi.waitFor(() => expect(document.body.querySelector("[data-template=poster]")).not.toBeNull())
    expect(fetchMock.mock.calls.every((c) => String(c[0]).startsWith("/api/share/proxy-image"))).toBe(true)
    expect(document.body.querySelectorAll("[role=tab]").length).toBe(3)
    w.unmount()
  })

  it("shows an error with retry when the proxy fails", async () => {
    const fetchMock = stubProxy(false)
    const w = mount(ShareImageModal, {
      props: { modelValue: true, shareType: "poster", media },
      global: { plugins: [i18n] },
      attachTo: document.body,
    })
    await vi.waitFor(() => expect(document.body.querySelector("[role=alert]")).not.toBeNull())
    expect(document.body.querySelector("[data-template]")).toBeNull()
    fetchMock.mockResolvedValue(new Response(new Blob([new Uint8Array([1])], { type: "image/jpeg" })))
    ;(document.body.querySelector("[role=alert] button") as HTMLButtonElement).click()
    await flushPromises()
    await vi.waitFor(() => expect(document.body.querySelector("[data-template=poster]")).not.toBeNull())
    w.unmount()
  })

  it("ignores stats when sharing someone's review", async () => {
    stubProxy(true)
    const w = mount(ShareImageModal, {
      props: { modelValue: true, shareType: "review", media, review, stats },
      global: { plugins: [i18n] },
      attachTo: document.body,
    })
    await vi.waitFor(() => expect(document.body.querySelector("[role=tab]")).not.toBeNull())
    const tabs = Array.from(document.body.querySelectorAll("[role=tab]")).map((t) => t.textContent?.trim())
    expect(tabs).toEqual(["Review", "Poster", "Square"])
    w.unmount()
  })
})

describe("native share / clipboard fallbacks", () => {
  beforeEach(() => setActivePinia(createPinia()))
  afterEach(() => vi.unstubAllGlobals())
  const blob = new Blob(["x"], { type: "image/png" })

  it("canShareFiles is false without the Web Share API", () => {
    vi.stubGlobal("navigator", { ...navigator, share: undefined, canShare: undefined })
    expect(useShareImage().canShareFiles()).toBe(false)
  })

  it("canShareFiles is false when files are unsupported", () => {
    vi.stubGlobal("navigator", { ...navigator, share: vi.fn(), canShare: vi.fn().mockReturnValue(false) })
    expect(useShareImage().canShareFiles()).toBe(false)
  })

  it("shares the file natively when supported, and treats dismissal as quiet", async () => {
    const share = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal("navigator", { ...navigator, share, canShare: vi.fn().mockReturnValue(true) })
    const s = useShareImage()
    expect(s.canShareFiles()).toBe(true)
    expect(await s.shareImageFile(blob, "a.png", "hi")).toBe(true)
    expect(share.mock.calls[0]![0].files[0].name).toBe("a.png")

    share.mockRejectedValueOnce(Object.assign(new Error("x"), { name: "AbortError" }))
    expect(await s.shareImageFile(blob, "a.png", "hi")).toBe(false)
  })

  it("shareImageFile does nothing when unsupported", async () => {
    vi.stubGlobal("navigator", { ...navigator, share: undefined, canShare: undefined })
    expect(await useShareImage().shareImageFile(blob, "a.png", "hi")).toBe(false)
  })

  it("copyImage uses ClipboardItem when available and reports unsupported otherwise", async () => {
    const write = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal("ClipboardItem", class { constructor(public items: unknown) {} })
    vi.stubGlobal("navigator", { ...navigator, clipboard: { write } })
    const s = useShareImage()
    expect(s.canCopyImage()).toBe(true)
    expect(await s.copyImage(blob)).toBe(true)
    expect(write).toHaveBeenCalledTimes(1)

    vi.stubGlobal("ClipboardItem", undefined)
    expect(s.canCopyImage()).toBe(false)
    expect(await s.copyImage(blob)).toBe(false)
  })
})
