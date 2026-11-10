import type {
  NotificationListResponse,
} from "@/types/notification"
import api from "../index"

export interface ListNotificationParams {
  page?: number
  limit?: number
  unread?: boolean
  category?: string
  type?: string
}

export const notiApi = {
  list: (params: ListNotificationParams = {}) =>
    api.get<NotificationListResponse>("/notifications", {
      params: { page: 1, limit: 20, ...params },
    }),

  getUnreadCount: () =>
    api.get<{ unread_count: number }>("/notifications/unread-count"),

  markAsRead: (notificationId: number) =>
    api.patch<void>("/notifications/read", { ids: [notificationId] }),

  markAllAsRead: () => api.patch<void>("/notifications/read", { ids: [] }),

  deleteNotification: (notificationId: number) =>
    api.delete<void>("/notifications", { data: { ids: [notificationId] } }),

  clearAll: () => api.delete<void>("/notifications", { params: { all: true } }),
}
