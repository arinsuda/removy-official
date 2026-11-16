export type MediaType = "movie" | "tv" | "manga"
export type ListType = "watchlist" | "likes" | "watched" | "email_change"
export type Visibility = "public" | "followers" | "private"

export interface PaginatedResult<T> {
  page: number
  results: T[]
  total_pages: number
  total_results: number
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
  total_pages: number
}
