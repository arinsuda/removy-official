import api from "../index"
import type {
  ChangePassword,
  UserProfile,
  RequestEmailChangeRequest,
  VerifyEmailChangeRequest,
} from "@/types"

export interface UserStats {
  user_id: number
  level: number
  current_exp: number
  review_count: number
  like_count: number
  watchlist_count: number
  watched_count: number
  follower_count: number
  following_count: number
}

export const userApi = {
  getStats: (userId: number) => api.get<{ stats: UserStats }>(`/users/${userId}/stats`),

  me: (userId: number) => api.get<{ user: UserProfile }>(`/users/${userId}`),

  getProfile: (userId: number) =>
    api.get<{ user: UserProfile }>(`/users/${userId}`),

  updateProfile: (userId: number, data: FormData) =>
    api.patch<{ user: UserProfile }>(`/users/${userId}`, data, {
      headers: { "Content-Type": "multipart/form-data" },
    }),

  deleteUser: (userId: number) => api.delete(`/users/${userId}`),

  updateFavoriteGenres: (userId: number, genres: number[]) =>
    api.patch<{ user: UserProfile }>(`/users/${userId}/genres`, {
      favorite_genres: genres,
    }),

  requestEmailChange: (userId: number, data: RequestEmailChangeRequest) =>
    api.post<{ message: string }>(`/users/${userId}/email`, data),

  verifyEmailChange: (userId: number, data: VerifyEmailChangeRequest) =>
    api.put<{ user: UserProfile }>(`/users/${userId}/email`, data),

  changePassword: (userId: number, data: ChangePassword) =>
    api.patch<{ message: string; access_token?: string; refresh_token?: string }>(
      `/users/${userId}/password`,
      data,
    ),
}
