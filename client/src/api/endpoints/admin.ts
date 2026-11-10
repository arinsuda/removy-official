import api from "../index"
import type {
  AdminOverviewStats,
  GrowthPoint,
  AdminUserPaginatedResponse,
  AdminReviewPaginatedResponse,
  AdminAuditLogPaginatedResponse,
  UserFilterQuery,
  EraseUserBody,
  ReviewFilterQuery,
  AuditLogFilterQuery,
} from "@/types"

export const adminApi = {
  getOverview: () =>
    api.get<{ overview: AdminOverviewStats }>("/admin/overview"),

  getGrowth: () =>
    api.get<{ growth: GrowthPoint[] }>("/admin/growth"),

  listUsers: (params: UserFilterQuery) =>
    api.get<AdminUserPaginatedResponse>("/admin/users", { params }),

  updateUserRole: (userId: number, role: string, reason?: string) =>
    api.patch(`/admin/users/${userId}/role`, { role, reason }),

  updateUserStatus: (userId: number, is_active: boolean, reason?: string) =>
    api.patch(`/admin/users/${userId}/status`, { is_active, reason }),

  deleteUser: (userId: number, reason?: string) =>
    api.delete(`/admin/users/${userId}`, { data: { reason } }),

  eraseUser: (userId: number, body: EraseUserBody) =>
    api.post(`/admin/users/${userId}/erasure`, body),

  listReviews: (params: ReviewFilterQuery) =>
    api.get<AdminReviewPaginatedResponse>("/admin/reviews", { params }),

  deleteReview: (reviewId: number, reason?: string) =>
    api.delete(`/admin/reviews/${reviewId}`, { data: { reason } }),

  listAuditLogs: (params: AuditLogFilterQuery) =>
    api.get<AdminAuditLogPaginatedResponse>("/admin/audit-logs", { params }),
}
