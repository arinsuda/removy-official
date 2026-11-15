import api from "../index"
import type {
  AdminEvent,
  AdminEventBody,
  AdminEventListQuery,
  AppEvent,
  AwardsQuery,
  AwardsResponse,
  EventListQuery,
  EventListResponse,
  NomineeInput,
  SuggestedNominee,
} from "@/types/event"

export const eventApi = {
  list: (params: EventListQuery) => api.get<EventListResponse>("/events", { params }),

  get: (id: number) => api.get<{ event: AppEvent }>(`/events/${id}`),

  popup: () => api.get<{ events: AppEvent[] }>("/events/popup"),

  vote: (eventId: number, nomineeId: number) =>
    api.put<{ success: boolean; nominee_id: number }>(`/events/${eventId}/vote`, { nominee_id: nomineeId }),

  cancelVote: (eventId: number) =>
    api.delete<{ success: boolean }>(`/events/${eventId}/vote`),

  getAwards: (params: AwardsQuery) => api.get<AwardsResponse>("/events/awards", { params }),
}

export const adminEventApi = {
  list: (params: AdminEventListQuery) => api.get<EventListResponse<AdminEvent>>("/admin/events", { params }),

  create: (body: AdminEventBody) => api.post<{ event: AdminEvent }>("/admin/events", body),

  update: (id: number, body: AdminEventBody) => api.put<{ event: AdminEvent }>(`/admin/events/${id}`, body),

  remove: (id: number) => api.delete(`/admin/events/${id}`),

  uploadImage: (id: number, file: File) => {
    const form = new FormData()
    form.append("image", file)
    return api.post<{ event: AdminEvent }>(`/admin/events/${id}/image`, form)
  },

  removeImage: (id: number) => api.delete<{ event: AdminEvent }>(`/admin/events/${id}/image`),

  finalize: (id: number) => api.post<{ success: boolean }>(`/admin/events/${id}/finalize`),

  setNominees: (id: number, nominees: NomineeInput[]) =>
    api.put<{ success: boolean }>(`/admin/events/${id}/nominees`, { nominees }),

  suggestNominees: (mediaType = "movie", period = "") =>
    api.get<{ nominees: SuggestedNominee[] }>("/admin/events/suggest-nominees", {
      params: { media_type: mediaType, period },
    }),

  generateMonthly: (mediaType = "movie", period = "") =>
    api.post<{ event: AdminEvent }>("/admin/events/generate-monthly", {
      media_type: mediaType,
      period,
    }),
}
