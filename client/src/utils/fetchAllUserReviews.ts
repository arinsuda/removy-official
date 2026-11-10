import { reviewApi } from "@/api/api"
import type { GetUserReviewsParams } from "@/api/endpoints/social"
import type { ReviewResponse } from "@/types"

/** Mirrors MaxMediaReviewsPageSize in server review_module/dto.go. */
export const USER_REVIEWS_MAX_PAGE_SIZE = 50
/** Safety cap so a misbehaving server cannot make us page forever. */
export const USER_REVIEWS_MAX_PAGES = 20

/** GET /reviews/user/:id is paginated; this walks the pages and returns every review (up to the cap). */
export async function fetchAllUserReviews(userId: number, params: GetUserReviewsParams = {}): Promise<ReviewResponse[]> {
  const all: ReviewResponse[] = []
  for (let page = 1; page <= USER_REVIEWS_MAX_PAGES; page++) {
    const res = await reviewApi.getUserReviews(userId, { ...params, page, page_size: USER_REVIEWS_MAX_PAGE_SIZE })
    all.push(...(res.data.reviews ?? []))
    if (!res.data.pagination?.has_more) break
  }
  return all
}
