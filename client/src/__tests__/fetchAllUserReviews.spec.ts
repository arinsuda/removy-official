import { beforeEach, describe, expect, it, vi } from "vitest"
import { fetchAllUserReviews, USER_REVIEWS_MAX_PAGES, USER_REVIEWS_MAX_PAGE_SIZE } from "@/utils/fetchAllUserReviews"

const m = vi.hoisted(() => ({ getUserReviews: vi.fn() }))
vi.mock("@/api/api", () => ({ reviewApi: { getUserReviews: m.getUserReviews } }))

const page = (ids: number[], hasMore: boolean) => ({ data: { reviews: ids.map((id) => ({ id })), pagination: { has_more: hasMore } } })

describe("fetchAllUserReviews", () => {
  beforeEach(() => m.getUserReviews.mockReset())

  it("pages with the max page size until has_more is false", async () => {
    m.getUserReviews.mockResolvedValueOnce(page([1, 2], true)).mockResolvedValueOnce(page([3], false))
    const all = await fetchAllUserReviews(7)
    expect(all.map((r) => r.id)).toEqual([1, 2, 3])
    expect(m.getUserReviews).toHaveBeenCalledTimes(2)
    expect(m.getUserReviews).toHaveBeenLastCalledWith(7, { page: 2, page_size: USER_REVIEWS_MAX_PAGE_SIZE })
  })

  it("stops at the page cap", async () => {
    m.getUserReviews.mockResolvedValue(page([1], true))
    await fetchAllUserReviews(7)
    expect(m.getUserReviews).toHaveBeenCalledTimes(USER_REVIEWS_MAX_PAGES)
  })
})
