import type { LikesResponse, MediaStatsResponse } from "@/types"
import api from "../index"

export const mediaApi = {
  // Shared Media Stats
  getMediaStats: (mediaType: "movie" | "tv", mediaId: number) =>
    api.get<MediaStatsResponse>(`/stats/${mediaType}/${mediaId}`),

  recordMediaView: (mediaType: "movie" | "tv", mediaId: number) =>
    api.post(`/stats/${mediaType}/${mediaId}/view`),

  getLikeByUserId: (userId: number) =>
    api.get<LikesResponse>(`/users/${userId}/likes`),
  
  likeMedia: (mediaType: "movie" | "tv", mediaId: number) =>
    api.post(`/stats/${mediaType}/${mediaId}/like`),

  unlikeMedia: (mediaType: "movie" | "tv", mediaId: number) =>
    api.delete(`/stats/${mediaType}/${mediaId}/like`),
}
