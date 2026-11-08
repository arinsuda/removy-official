import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { i18n } from "../i18n"
import PopupCard from "../components/movie/PopupCard.vue"
import PopcornRatingIcon from "../components/movie/PopcornRatingIcon.vue"
import PopcornScoreBadge from "../components/movie/PopcornScoreBadge.vue"
import { mediaApi } from "../api/endpoints/media"
import { libraryApi } from "../api/endpoints/library"

// Mock dependencies
vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({
    user: { id: 1 }
  })
}))

vi.mock("../api/endpoints/media", () => ({
  mediaApi: {
    getMediaStats: vi.fn(),
    likeMedia: vi.fn(),
    unlikeMedia: vi.fn()
  }
}))

vi.mock("../api/endpoints/library", () => ({
  libraryApi: {
    getOwnMediaStatus: vi.fn().mockResolvedValue({ data: { in_lists: [] } }),
    addItem: vi.fn(),
    removeItem: vi.fn()
  }
}))

describe("Phase 1 Frontend Requirements", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    i18n.global.locale.value = "th"
  })

  describe("vue-i18n Translation Mapping", () => {
    it("maps DUPLICATE_REVIEW and INVALID_RATING translation keys correctly in Thai", () => {
      i18n.global.locale.value = "th"
      expect(i18n.global.t("reviews.errors.alreadyExists")).toBe("คุณรีวิวสื่อนี้ไปแล้ว")
      expect(i18n.global.t("reviews.errors.invalidRating")).toBe("คะแนนต้องอยู่ระหว่าง 0.5 ถึง 5.0")
      expect(i18n.global.t("reviews.errors.notRated")).toBe("ไม่มีคะแนน")
    })

    it("maps DUPLICATE_REVIEW and INVALID_RATING translation keys correctly in English", () => {
      i18n.global.locale.value = "en"
      expect(i18n.global.t("reviews.errors.alreadyExists")).toBe("You have already reviewed this media")
      expect(i18n.global.t("reviews.errors.invalidRating")).toBe("Rating must be between 0.5 and 5.0")
      expect(i18n.global.t("reviews.errors.notRated")).toBe("NR")
    })
  })

  describe("PopcornRatingIcon Collision Safety", () => {
    it("generates unique element ID prefixes to prevent DOM collisions", () => {
      const wrapper1 = mount(PopcornRatingIcon)
      const wrapper2 = mount(PopcornRatingIcon)
      const gradId1 = wrapper1.find("linearGradient").attributes("id")
      const gradId2 = wrapper2.find("linearGradient").attributes("id")

      expect(gradId1).toContain("popcorn-grad-")
      expect(gradId2).toContain("popcorn-grad-")
      expect(gradId1).not.toBe(gradId2)
    })
  })

  describe("PopupCard Rating & Stats Integration", () => {
    const mockMovieWithPopcorn = {
      id: 101,
      title: "Sample Movie",
      vote_average: 7.8,
      vote_count: 50,
      ratings: {
        popcorn: { average: 4.2, count: 5, available: true, scale: 5.0 },
        tmdb: { average: 7.8, count: 50, available: true, scale: 10.0 }
      }
    }

    const mockMovieTmdbOnly = {
      id: 102,
      title: "TMDB Only Movie",
      vote_average: 8.1,
      vote_count: 120,
      ratings: {
        popcorn: { average: 0.0, count: 0, available: false, scale: 5.0 },
        tmdb: { average: 8.1, count: 120, available: true, scale: 10.0 }
      }
    }

    const mockMovieUnrated = {
      id: 103,
      title: "Unrated Movie",
      vote_average: 0.0,
      vote_count: 0,
      ratings: {
        popcorn: { average: 0.0, count: 0, available: false, scale: 5.0 },
        tmdb: { average: 0.0, count: 0, available: false, scale: 10.0 }
      }
    }

    it("renders POPCORN ratings first when available", () => {
      const wrapper = mount(PopupCard, {
        global: { plugins: [i18n] },
        props: {
          movie: mockMovieWithPopcorn,
          currentTrailer: null,
          trailerUnavailable: true,
          isIframeMounted: false,
          isIframeLoaded: false,
          showSkeleton: false,
          showFallback: true,
          attachPlayer: () => {}
        }
      })

      expect(wrapper.findComponent(PopcornScoreBadge).attributes("data-tier")).toBe("hot")
      expect(wrapper.text()).toContain("4.2/5")
    })

    it("falls back to TMDB rating if POPCORN is not available", () => {
      const wrapper = mount(PopupCard, {
        global: { plugins: [i18n] },
        props: {
          movie: mockMovieTmdbOnly,
          currentTrailer: null,
          trailerUnavailable: true,
          isIframeMounted: false,
          isIframeLoaded: false,
          showSkeleton: false,
          showFallback: true,
          attachPlayer: () => {}
        }
      })

      expect(wrapper.findComponent(PopcornScoreBadge).attributes("data-tier")).toBe("none")
      expect(wrapper.text()).toContain("0.0")
      expect(wrapper.text()).toContain("8.1")
    })

    it("displays accessible not-rated translation state when no rating is available", () => {
      i18n.global.locale.value = "th"
      const wrapper = mount(PopupCard, {
        global: { plugins: [i18n] },
        props: {
          movie: mockMovieUnrated,
          currentTrailer: null,
          trailerUnavailable: true,
          isIframeMounted: false,
          isIframeLoaded: false,
          showSkeleton: false,
          showFallback: true,
          attachPlayer: () => {}
        }
      })

      expect(wrapper.text()).toContain("ไม่มีคะแนน")
    })

    it("makes no API calls to mediaApi.getMediaStats on mount", () => {
      mount(PopupCard, {
        global: { plugins: [i18n] },
        props: {
          movie: mockMovieWithPopcorn,
          currentTrailer: null,
          trailerUnavailable: true,
          isIframeMounted: false,
          isIframeLoaded: false,
          showSkeleton: false,
          showFallback: true,
          attachPlayer: () => {}
        }
      })

      expect(mediaApi.getMediaStats).not.toHaveBeenCalled()
    })

    it("preserves heart/like favorite icon structure and toggles", () => {
      const wrapper = mount(PopupCard, {
        global: { plugins: [i18n] },
        props: {
          movie: mockMovieWithPopcorn,
          currentTrailer: null,
          trailerUnavailable: true,
          isIframeMounted: false,
          isIframeLoaded: false,
          showSkeleton: false,
          showFallback: true,
          attachPlayer: () => {}
        }
      })

      const favButton = wrapper.find(".action-btn--favorite")
      expect(favButton.exists()).toBe(true)
      expect(favButton.find("svg").exists()).toBe(true)
    })
  })
})
