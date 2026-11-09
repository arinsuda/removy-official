import api from "../index"
import type { CreateListRequest, ListDetailResponse, ListItem, ListSummary, MediaType, UpdateListRequest } from "@/types"

export const listApi = {
  /** Own lists; with media set, each carries `contains_media`. GET /api/lists */
  getOwn: (media?: { media_type: MediaType; media_id: number }) =>
    api.get<{ lists: ListSummary[] }>("/lists", { params: media }),

  /** Lists of another user that the viewer may see. GET /api/users/:userId/lists */
  getForUser: (userId: number) => api.get<{ lists: ListSummary[] }>(`/users/${userId}/lists`),

  create: (data: CreateListRequest) => api.post<{ list: ListSummary }>("/lists", data),

  get: (listId: number, page = 1) => api.get<ListDetailResponse>(`/lists/${listId}`, { params: { page } }),

  update: (listId: number, data: UpdateListRequest) => api.patch<{ list: ListSummary }>(`/lists/${listId}`, data),

  remove: (listId: number) => api.delete(`/lists/${listId}`),

  addItem: (listId: number, data: { media_id: number; media_type: MediaType }) =>
    api.post<{ item: ListItem }>(`/lists/${listId}/items`, data),

  removeItem: (listId: number, itemId: number) => api.delete(`/lists/${listId}/items/${itemId}`),

  /** Sends every item id of the list in the new order. */
  reorder: (listId: number, itemIds: number[]) => api.put(`/lists/${listId}/items/order`, { item_ids: itemIds }),
}
