import type { PaginatedResponse } from "./common"

export type GrowthStatus = "growth" | "decline" | "no_change" | "no_previous_baseline"

export interface AdminOverviewStats {
  total_registered_users: number
  active_users_count: number
  inactive_users_count: number
  current_month_registrations: number
  previous_month_registrations: number
  absolute_growth: number
  growth_percentage: number | null
  growth_status: GrowthStatus
  unique_online_users: number
  total_activity_events: number
  activity_events_today: number
  dau_today: number
  wau_7d: number
  mau_30d: number
  total_reviews: number
  total_media_likes: number
}

export interface GrowthPoint {
  month: string
  user_count: number
}

export interface AdminUserRow {
  id: number
  username: string
  email: string
  display_name: string | null
  avatar_url: string | null
  role: string
  is_active: boolean
  created_at: string
  deleted_at: string | null
  review_count: number
}

export interface AdminReviewRow {
  id: number
  user_id: number
  username: string
  media_id: number
  media_type: string
  media_title?: string
  poster_url?: string
  rating: number
  body: string
  is_public: boolean
  like_count: number
  created_at: string
}

export interface AdminAuditLogRow {
  id: number
  admin_id: number | null
  admin_username: string | null
  action: string
  target_type: string
  target_id: number
  reason: string | null
  meta_data: Record<string, unknown> | null
  created_at: string
}

export type UserStatusFilter = "all" | "active" | "inactive" | "deleted"

export interface EraseUserBody {
  confirm_username: string
  reason?: string
}

export interface UserFilterQuery {
  page?: number
  limit?: number
  search?: string
  role?: string
  status?: UserStatusFilter
  sort_by?: string
  sort_order?: "asc" | "desc"
}

export interface ReviewFilterQuery {
  page?: number
  limit?: number
  search?: string
  media_type?: string
  visibility?: string
  min_rating?: number
  max_rating?: number
  sort_by?: string
  sort_order?: "asc" | "desc"
}

export interface AuditLogFilterQuery {
  page?: number
  limit?: number
  action?: string
  target_type?: string
}

export type AdminUserPaginatedResponse = PaginatedResponse<AdminUserRow>
export type AdminReviewPaginatedResponse = PaginatedResponse<AdminReviewRow>
export type AdminAuditLogPaginatedResponse = PaginatedResponse<AdminAuditLogRow>
