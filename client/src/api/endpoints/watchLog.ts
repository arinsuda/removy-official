import api from "../index";
import type { MediaType } from "@/types/common";
import type {
  WatchLogResponse,
  CreateWatchLogRequest,
  UpdateWatchLogRequest,
  WatchLogListResponse,
  WatchSummaryResponse,
} from "@/types";

export const watchLogApi = {
  createWatchLog: (
    mediaType: MediaType,
    mediaId: number,
    data: CreateWatchLogRequest
  ) =>
    api.post<{ log: WatchLogResponse }>(`/media/${mediaType}/${mediaId}/watch-logs`, data),

  getMyWatchLogs: (mediaType: MediaType, mediaId: number) =>
    api.get<{ logs: WatchLogResponse[]; summary: WatchSummaryResponse }>(
      `/media/${mediaType}/${mediaId}/watch-logs/me`
    ),

  updateWatchLog: (watchLogId: number, data: UpdateWatchLogRequest) =>
    api.patch<{ log: WatchLogResponse }>(`/watch-logs/${watchLogId}`, data),

  deleteWatchLog: (watchLogId: number) =>
    api.delete(`/watch-logs/${watchLogId}`),

  getUserWatchHistory: (
    userId: number,
    params?: { page?: number; limit?: number }
  ) =>
    api.get<WatchLogListResponse>(`/users/${userId}/watch-history`, { params }),
};
