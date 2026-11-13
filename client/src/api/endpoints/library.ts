import api from "../index"
import type {
  AddItemRequest,
  LibraryItemResponse,
  MediaStatusResponse,
} from "@/types"

export const libraryApi = {
  /**
   * เพิ่มไอเทมเข้าคลังภาพยนตร์ (Watchlist / Watched / Favorite)
   * POST /api/library
   */
  addItem: (data: AddItemRequest) =>
    api.post<{ item: LibraryItemResponse }>("/library", data),

  /**
   * ดึงรายการหนังในคลังของยูสเซอร์คนอื่น/ตัวเองแบบผ่าน Visibility Policy
   * GET /api/library/user/:userId
   */
  getVisibleUserLibrary: (
    userId: number,
    params?: { list_type?: string; media_type?: string },
  ) =>
    api.get<{ items: LibraryItemResponse[]; has_more?: boolean }>(`/library/user/${userId}`, {
      params,
    }),

  /**
   * ตรวจสอบสถานะสื่อของตัวเอง
   * GET /api/library/media/:mediaType/:mediaId
   */
  getOwnMediaStatus: (
    mediaId: number,
    mediaType: "movie" | "tv",
  ) =>
    api.get<MediaStatusResponse>(`/library/media/${mediaType}/${mediaId}`),

  /**
   * อัปเดตไอเทมในคลัง (watched_at, tags, note)
   * PATCH /api/library/:itemId
   */
  updateItem: (
    itemId: number,
    data: { watched_at?: string | null; tags?: string[]; note?: string | null },
  ) =>
    api.patch<{ item: LibraryItemResponse }>(`/library/${itemId}`, data),

  /**
   * ลบไอเทมออกจากคลัง
   * DELETE /api/library/:itemId
   */
  removeItem: (itemId: number) =>
    api.delete(`/library/${itemId}`),
}
