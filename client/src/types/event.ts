export type EventType = "announcement" | "vote_monthly" | "vote_yearly"
export type EventState = "upcoming" | "active" | "past"
export type EventStatus = "draft" | "published" | "cancelled"

export interface LocalizedText {
  en: string
  th: string
}

export interface Nominee {
  id: number
  event_id: number
  media_id: number
  media_type: "movie" | "tv"
  position: number
  media_title: string
  poster_path?: string | null
  backdrop_path?: string | null
  overview?: string | null
  vote_count?: number
  percentage?: number
  is_winner: boolean
}

export interface AppEvent {
  id: number
  type: EventType
  media_type: "movie" | "tv" | null
  period: string | null
  title: LocalizedText
  body: LocalizedText
  image_url: string | null
  /** TMDB backdrop of the leading nominee; used as the cover when there is no image_url. */
  cover_backdrop_path?: string | null
  link_url: string | null
  starts_at: string
  ends_at: string
  show_popup: boolean
  popup_priority: number
  state: EventState
  nominees?: Nominee[]
  my_vote_nominee_id?: number | null
  total_votes?: number | null
  is_finalized?: boolean
}

export type AwardKind = "monthly" | "yearly"
export type AwardMediaType = "movie" | "tv"

/** A won award as shown on a movie or TV detail page. */
export interface MediaAward {
  kind: AwardKind
  /** "YYYY-MM" for monthly, "YYYY" for yearly. Co-winners share the same period. */
  period: string
}

/** One winner of a finalized round, as listed on the awards page. */
export interface EventAward extends MediaAward {
  event_id: number
  media_id: number
  media_type: AwardMediaType
  media_title: string
  poster_path: string | null
  backdrop_path: string | null
  votes: number
}

export interface AwardsQuery {
  media_type: AwardMediaType
  year: number
}

export interface AwardsResponse {
  awards: EventAward[]
  /** Years that have at least one award for the requested media type. */
  years: number[]
}

export interface AdminEvent extends AppEvent {
  status: EventStatus
  created_at: string
  updated_at: string
}

export interface NomineeInput {
  id?: number
  media_id: number
  media_type: "movie" | "tv"
  position?: number
  media_title: string
  poster_path?: string
  backdrop_path?: string
  overview?: string
}

export interface SuggestedNominee {
  media_id: number
  media_type: "movie" | "tv"
  media_title: string
  poster_path: string
  backdrop_path: string
  overview: string
  score: number
  source: string
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
  type?: EventType
  media_type?: "movie" | "tv" | null
  period?: string | null
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
  nominees?: NomineeInput[]
}
