import type { LikesResponse, MediaStatsResponse } from "@/types"
import api from "../index"
import type { MediaType } from "@/types/common"

export const mediaApi = {
  // Shared Media Stats
  getMediaStats: (mediaType: MediaType, mediaId: number) =>
    api.get<MediaStatsResponse>(`/stats/${mediaType}/${mediaId}`),

  recordMediaView: (mediaType: MediaType, mediaId: number) =>
    api.post(`/stats/${mediaType}/${mediaId}/view`),

  getLikeByUserId: (userId: number) =>
    api.get<LikesResponse>(`/users/${userId}/likes`),
  
  likeMedia: (mediaType: MediaType, mediaId: number) =>
    api.post(`/stats/${mediaType}/${mediaId}/like`),

  unlikeMedia: (mediaType: MediaType, mediaId: number) =>
    api.delete(`/stats/${mediaType}/${mediaId}/like`),
}
