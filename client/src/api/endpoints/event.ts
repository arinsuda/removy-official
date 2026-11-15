import api from "../index"
import type {
  AdminEvent,
  AdminEventBody,
  AdminEventListQuery,
  AppEvent,
  AutoPublishInput,
  AwardsQuery,
  AwardsResponse,
  EventListQuery,
  EventListResponse,
  GenerateRoundInput,
  NomineeInput,
  SuggestedNominee,
} from "@/types/event"

const ADMIN_EVENTS_PATH = "/admin/events"
const GENERATE_MONTHLY_PATH = `${ADMIN_EVENTS_PATH}/generate-monthly`
const GENERATE_YEARLY_PATH = `${ADMIN_EVENTS_PATH}/generate-yearly`

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
  list: (params: AdminEventListQuery) => api.get<EventListResponse<AdminEvent>>(ADMIN_EVENTS_PATH, { params }),

  create: (body: AdminEventBody) => api.post<{ event: AdminEvent }>(ADMIN_EVENTS_PATH, body),

  update: (id: number, body: AdminEventBody) => api.put<{ event: AdminEvent }>(`${ADMIN_EVENTS_PATH}/${id}`, body),

  remove: (id: number) => api.delete(`${ADMIN_EVENTS_PATH}/${id}`),

  setAutoPublish: (id: number, enabled: boolean) =>
    api.patch<{ event: AdminEvent }>(`${ADMIN_EVENTS_PATH}/${id}`, { auto_publish: enabled } satisfies AutoPublishInput),

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

  generateMonthly: (body: GenerateRoundInput) => api.post<{ event: AdminEvent }>(GENERATE_MONTHLY_PATH, body),

  generateYearlyRound: (body: GenerateRoundInput) => api.post<{ event: AdminEvent }>(GENERATE_YEARLY_PATH, body),
}
