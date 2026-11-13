export type EventType = "announcement" | "vote_monthly" | "vote_yearly"
export type EventState = "upcoming" | "active" | "past"
export type EventStatus = "draft" | "published" | "cancelled"

export interface LocalizedText {
  en: string
  th: string
}

export interface AppEvent {
  id: number
  type: EventType
  media_type: "movie" | "tv" | null
  period: string | null
  title: LocalizedText
  body: LocalizedText
  image_url: string | null
  link_url: string | null
  starts_at: string
  ends_at: string
  show_popup: boolean
  popup_priority: number
  state: EventState
}

export interface AdminEvent extends AppEvent {
  status: EventStatus
  created_at: string
  updated_at: string
}

export interface EventPagination {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface EventListResponse<T = AppEvent> {
  events: T[]
  pagination: EventPagination
}

export interface EventListQuery {
  state?: EventState
  page?: number
  limit?: number
}

export interface AdminEventListQuery {
  status?: EventStatus
  page?: number
  limit?: number
}

export interface AdminEventBody {
  title_en: string
  title_th: string
  body_en: string
  body_th: string
  link_url: string
  starts_at: string
  ends_at: string
  show_popup: boolean
  popup_priority: number
  status: EventStatus
}
