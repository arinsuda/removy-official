import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createI18n } from "vue-i18n"
import { useShareImage } from "@/composables/useShareImage"
import PosterShareTemplate from "@/components/share/PosterShareTemplate.vue"
import ReviewShareTemplate from "@/components/share/ReviewShareTemplate.vue"
import ShareImageModal from "@/components/share/ShareImageModal.vue"
import type { ShareMediaContext, ShareReviewContext } from "@/types/share"

// Mock html-to-image
vi.mock("html-to-image", () => ({
  toPng: vi.fn().mockResolvedValue("data:image/png;base64,fakePNGData"),
}))

const i18n = createI18n({
  legacy: false,
  locale: "en",
  messages: {
    en: {
      share: {
        share: "Share",
        shareAsImage: "Share as Image",
        downloadImage: "Download Image",
        copyCaption: "Copy Caption",
        copyLink: "Copy Link",
        sharedFrom: "Shared from POPCORN",
        reviewedOn: "Reviewed on POPCORN",
        movie: "Movie",
        tvSeries: "TV Series",
        downloadSuccess: "Image downloaded",
        captionCopied: "Caption copied!",
        linkCopied: "Link copied!",
        generating: "Generating...",
        storyFormat: "Story",
        posterShare: "Poster Share",
        reviewShare: "Review Share",
      },
      common: {
        close: "Close",
      },
    },
  },
})

describe("Share as Image Feature", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe("useShareImage Composable Logic", () => {
    it("truncates review text correctly for long text", () => {
      const { truncateReview } = useShareImage()
      const longText = "This is a very long review text that exceeds the max characters limit set for the share card template layout. It should be safely truncated with an ellipsis at the end without breaking words or surrogate pairs."
      const truncated = truncateReview(longText, 50)
      expect(truncated.length).toBeLessThanOrEqual(51)
      expect(truncated.endsWith("…")).toBe(true)
    })

    it("does not truncate short review text", () => {
      const { truncateReview } = useShareImage()
      const shortText = "Great movie!"
      expect(truncateReview(shortText, 50)).toBe("Great movie!")
    })

    it("sanitizes filenames safely", () => {
      const { sanitizeFilename } = useShareImage()
      const filename = sanitizeFilename("Spider-Man: Across the Spider-Verse!!!", "POPCORN_poster")
      expect(filename).toContain("POPCORN_poster_Spider-Man_Across_the_Spider-Verse")
      expect(filename.endsWith(".png")).toBe(true)
    })

    it("builds poster caption correctly", () => {
      const { buildPosterCaption } = useShareImage()
      const caption = buildPosterCaption("Inception", "2010", "movie")
      expect(caption).toContain("Inception (2010)")
      expect(caption).toContain("Shared from POPCORN")
    })

    it("builds review caption correctly with author attribution", () => {
      const { buildReviewCaption } = useShareImage()
      const caption = buildReviewCaption("Dune: Part Two", "JohnDoe", 4.5)
      expect(caption).toContain("4.5/5 — Dune: Part Two")
      expect(caption).toContain("JohnDoe")
      expect(caption).toContain("Shared from POPCORN")
    })

    it("builds share links accurately", () => {
      const { buildShareLink } = useShareImage()
      expect(buildShareLink("movie", 550)).toContain("/movies/550")
      expect(buildShareLink("tv", 1399)).toContain("/tv/1399")
    })
  })

  describe("PosterShareTemplate Component", () => {
    const mockMedia: ShareMediaContext = {
      id: 550,
      mediaType: "movie",
      title: "Fight Club",
      posterPath: "/poster.jpg",
      posterUrl: null,
      releaseYear: "1999",
      genres: [{ id: 18, name: "Drama" }],
      voteAverage: 8.4,
      popcornRating: 4.8,
    }

    it("renders movie title, year, genre chips and POPCORN rating", () => {
      const wrapper = mount(PosterShareTemplate, {
        props: {
          media: mockMedia,
          posterBlobUrl: "blob:http://localhost/fake-poster-blob",
        },
        global: {
          plugins: [i18n],
        },
      })

      expect(wrapper.find(".media-title").text()).toBe("Fight Club")
      expect(wrapper.find(".media-year").text()).toBe("1999")
      expect(wrapper.find(".genre-chip").text()).toBe("Drama")
      expect(wrapper.find(".rating-value").text()).toBe("4.8")
      expect(wrapper.find(".rating-source").text()).toBe("POPCORN")
    })

    it("renders poster fallback when posterBlobUrl is null", () => {
      const wrapper = mount(PosterShareTemplate, {
        props: {
          media: { ...mockMedia, posterPath: null },
          posterBlobUrl: null,
        },
        global: {
          plugins: [i18n],
        },
      })

      expect(wrapper.find(".poster-fallback").exists()).toBe(true)
    })
  })

  describe("ReviewShareTemplate Component", () => {
    const mockMedia: ShareMediaContext = {
      id: 101,
      mediaType: "movie",
      title: "Interstellar",
      posterPath: "/interstellar.jpg",
      posterUrl: null,
      releaseYear: "2014",
      genres: [{ id: 878, name: "Sci-Fi" }],
      voteAverage: 8.6,
      popcornRating: 4.9,
    }

    const mockReview: ShareReviewContext = {
      id: 1,
      authorDisplayName: "Nolan Fan",
      authorUsername: "nolanfan",
      authorAvatarUrl: "/avatar.jpg",
      rating: 5.0,
      body: "A masterpiece of modern cinema.",
      visibility: 'public',
      createdAt: "2024-01-16",
    }

    it("renders review author attribution, rating, and body text", () => {
      const wrapper = mount(ReviewShareTemplate, {
        props: {
          media: mockMedia,
          review: mockReview,
          posterBlobUrl: "blob:http://localhost/fake-poster",
          avatarBlobUrl: "blob:http://localhost/fake-avatar",
        },
        global: {
          plugins: [i18n],
        },
      })

      expect(wrapper.find(".author-name").text()).toBe("Nolan Fan")
      expect(wrapper.find(".author-username").text()).toBe("@nolanfan")
      expect(wrapper.find(".review-text").text()).toContain("masterpiece")
      expect(wrapper.find(".reviewed-label").text()).toBe("REVIEWED ON")
    })

    it("renders avatar fallback when avatarBlobUrl is null", () => {
      const wrapper = mount(ReviewShareTemplate, {
        props: {
          media: mockMedia,
          review: { ...mockReview, authorAvatarUrl: null },
          posterBlobUrl: null,
          avatarBlobUrl: null,
        },
        global: {
          plugins: [i18n],
        },
      })

      expect(wrapper.find(".avatar-fallback").exists()).toBe(true)
    })
  })

  describe("ShareImageModal Component", () => {
    const mockMedia: ShareMediaContext = {
      id: 200,
      mediaType: "movie",
      title: "Spider-Man",
      posterPath: "/spiderman.jpg",
      posterUrl: null,
      releaseYear: "2002",
      genres: [{ id: 28, name: "Action" }],
      voteAverage: 7.3,
      popcornRating: 4.5,
    }

    it("renders modal shell with download and copy buttons when open", () => {
      const wrapper = mount(ShareImageModal, {
        props: {
          modelValue: true,
          shareType: "poster",
          media: mockMedia,
        },
        global: {
          plugins: [i18n],
        },
        attachTo: document.body,
      })

      expect(document.body.querySelector(".btn-download")).not.toBeNull()
      expect(document.body.querySelectorAll(".btn-secondary").length).toBe(2)
      wrapper.unmount()
    })
  })
})
