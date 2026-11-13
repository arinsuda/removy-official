import api from "../index"
import type {
  AdminEvent,
  AdminEventBody,
  AdminEventListQuery,
  AppEvent,
  EventListQuery,
  EventListResponse,
} from "@/types/event"

export const eventApi = {
  list: (params: EventListQuery) => api.get<EventListResponse>("/events", { params }),

  get: (id: number) => api.get<{ event: AppEvent }>(`/events/${id}`),

  popup: () => api.get<{ events: AppEvent[] }>("/events/popup"),
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
}
