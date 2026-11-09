import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import MovieReviews from "@/components/movie/MovieReviews.vue"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ getMediaReviews: vi.fn(), getMyReview: vi.fn(), getFollowStatus: vi.fn() }))
vi.mock("@/api/api", () => ({
  reviewApi: m,
  followApi: { getFollowStatus: m.getFollowStatus },
}))
vi.mock("html-to-image", () => ({ toPng: vi.fn() }))

const i18n = useRealI18n()
const review = (id: number, userId: number) => ({
  id, user: { id: userId, username: `u${userId}`, display_name: null, avatar_url: null },
  media: { id: 5, media_type: "movie", title: "T", genres: [] }, rating: 4, body: `b${id}`, visibility: "public",
  like_count: 0, comment_count: 0, is_liked: false, helpful_count: 0, is_helpful_voted: false, created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z",
})

async function mountReviews(userId: number | null) {
  const pinia = createPinia()
  setActivePinia(pinia)
  if (userId) useAuthStore().user = { id: userId } as never
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: "/login", name: "login", component: { render: () => null } }, { path: "/:p(.*)*", component: { render: () => null } }] })
  const w = mount(MovieReviews, { props: { movieId: 5, mediaType: "movie" }, global: { plugins: [i18n, pinia, router] } })
  await flushPromises()
  return w
}

describe("MovieReviews paging", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    m.getFollowStatus.mockResolvedValue({ data: {} })
    m.getMediaReviews.mockImplementation((_t: string, _id: number, p: { page: number }) =>
      Promise.resolve({
        data: p.page === 1
          ? { reviews: [review(1, 2), review(2, 3)], pagination: { page: 1, page_size: 2, total: 3, has_more: true } }
          : { reviews: [review(3, 4)], pagination: { page: 2, page_size: 2, total: 3, has_more: false } },
      }),
    )
  })

  it("loads more pages and shows the total", async () => {
    m.getMyReview.mockResolvedValue({ data: { review: null } })
    const w = await mountReviews(1)
    expect(w.findAll("li")).toHaveLength(2)
    expect(w.get('[data-test="reviews-total"]').text()).toContain("3")
    await w.get('[data-test="reviews-more"]').trigger("click")
    await flushPromises()
    expect(w.findAll("li")).toHaveLength(3)
    expect(w.find('[data-test="reviews-more"]').exists()).toBe(false)
  })

  it("finds the own review even when it is not on page 1, so no composer shows", async () => {
    m.getMyReview.mockResolvedValue({ data: { review: review(9, 1) } })
    const w = await mountReviews(1)
    expect(w.text()).toContain("b9")
    expect(w.findComponent({ name: "ReviewComposer" }).exists()).toBe(false)
  })

  it("shows the composer only when the user has no review for this title", async () => {
    m.getMyReview.mockResolvedValue({ data: { review: null } })
    const w = await mountReviews(1)
    expect(w.findComponent({ name: "ReviewComposer" }).exists()).toBe(true)
  })

  it("does not look up an own review for guests", async () => {
    const w = await mountReviews(null)
    expect(m.getMyReview).not.toHaveBeenCalled()
    expect(w.findComponent({ name: "ReviewComposer" }).exists()).toBe(false)
  })

  it("fetches follow status only for authors of the new page", async () => {
    m.getMyReview.mockResolvedValue({ data: { review: null } })
    const w = await mountReviews(1)
    expect(m.getFollowStatus.mock.calls.map((c) => c[0]).sort()).toEqual([2, 3])
    await w.get('[data-test="reviews-more"]').trigger("click")
    await flushPromises()
    expect(m.getFollowStatus.mock.calls.map((c) => c[0]).sort()).toEqual([2, 3, 4])
  })

  it("keeps the composer hidden and offers a retry when the own review fails to load", async () => {
    m.getMyReview.mockRejectedValueOnce({ isAxiosError: true, response: { status: 500 } })
    const w = await mountReviews(1)
    expect(w.findComponent({ name: "ReviewComposer" }).exists()).toBe(false)
    expect(w.find('[data-test="own-failed"]').exists()).toBe(true)
    m.getMyReview.mockResolvedValueOnce({ data: { review: null } })
    await w.get('[data-test="own-retry"]').trigger("click")
    await flushPromises()
    expect(w.find('[data-test="own-failed"]').exists()).toBe(false)
    expect(w.findComponent({ name: "ReviewComposer" }).exists()).toBe(true)
  })
})
