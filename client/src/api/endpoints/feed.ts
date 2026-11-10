import api from "../index";
import type {
  FeedListResponse,
  FeedQueryParams,
  FeedScope,
  ActivitySettingsResponse,
  UpdateActivitySettingsRequest,
} from "@/types/feed";

export const feedApi = {
  // GET /feed — ฟีดของคนที่ user ติดตามอยู่ (accepted follow เท่านั้น)
  getFeed: (params: FeedQueryParams = {}) =>
    api.get<FeedListResponse>("/feed", {
      params: { page: 1, limit: 20, ...params },
    }),

  // GET /users/:userId/activities — activity ของ user คนเดียว ใช้กับหน้าโปรไฟล์
  getUserActivities: (userId: number, params: FeedQueryParams = {}) =>
    api.get<FeedListResponse>(`/users/${userId}/activities`, {
      params: { page: 1, limit: 20, ...params },
    }),

  // PATCH /activities/:activityId/visibility — อัปเดตความเป็นส่วนตัวของ activity
  updateVisibility: (
    activityId: number,
    visibility: "default" | "public" | "followers" | "private"
  ) =>
    api.patch(`/activities/${activityId}/visibility`, {
      visibility,
    }),

  // DELETE /activities/:activityId — ลบ activity ของตนเองออกจาก feed
  deleteActivity: (activityId: number) =>
    api.delete(`/activities/${activityId}`),

  // GET /me/activity-settings
  getSettings: () => api.get<ActivitySettingsResponse>("/me/activity-settings"),

  // PATCH /me/activity-settings
  updateSettings: (data: UpdateActivitySettingsRequest) =>
    api.patch<ActivitySettingsResponse>("/me/activity-settings", data),

  // GET /feed/new-count — นับกิจกรรมใหม่ตั้งแต่ after_id
  getNewCount: (afterId: number, category?: string, scope?: FeedScope) =>
    api.get<{ count: number }>("/feed/new-count", {
      params: { after_id: afterId, category, scope },
    }),
};

