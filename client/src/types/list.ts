import type { MediaType, Visibility } from "./common"
import type { MediaSummary } from "./movie"

/** A user-made list without its items. */
export interface ListSummary {
  id: number
  user_id: number
  name: string
  description: string
  visibility: Visibility
  item_count: number
  cover_item_id: number | null
  /** Poster of the picked cover item; empty when none is picked. */
  cover_poster?: string
  /** Up to four posters, the fallback when no cover is picked. */
  collage_posters: string[]
  /** Only present when the list was requested with media_type and media_id. */
  contains_media?: boolean
  created_at: string
  updated_at: string
}

export interface ListItem {
  id: number
  position: number
  media: MediaSummary
  media_type: MediaType
  added_at: string
}

export interface ListPagination {
  page: number
  page_size: number
  total: number
  has_more: boolean
}

export interface ListDetailResponse {
  list: ListSummary
  items: ListItem[]
  pagination: ListPagination
}

export interface CreateListRequest {
  name: string
  description?: string
  visibility?: Visibility
}

export interface UpdateListRequest {
  name?: string
  description?: string
  visibility?: Visibility
  /** 0 clears the cover. */
  cover_item_id?: number
}
