import api from "../index"
import type { BMOLItemResponse, MediaType } from "@/types"

export const bmolApi = {
  /**
   * เพิ่มไอเทมเข้า BMOL ranking
   * POST /api/bmol
   */
  addItem: (data: { media_id: number; media_type: MediaType; rank: number }) =>
    api.post<{ item: BMOLItemResponse }>("/bmol", data),

  /**
   * ดึงรายการ BMOL ranking ของยูสเซอร์
   * GET /api/bmol/user/:userId
   */
  getUserBMOL: (userId: number, params?: { media_type?: MediaType }) =>
    api.get<{ items: BMOLItemResponse[] }>(`/bmol/user/${userId}`, { params }),

  /**
   * อัปเดตอันดับและ/หรือโน้ตของไอเทมใน BMOL (ต้องส่งอย่างน้อยหนึ่งฟิลด์)
   * PUT /api/bmol/:itemId
   */
  updateItem: (itemId: number, data: { rank?: number; note?: string }) =>
    api.put<{ item: BMOLItemResponse }>(`/bmol/${itemId}`, data),

  /**
   * ลบไอเทมออกจาก BMOL ranking
   * DELETE /api/bmol/:itemId
   */
  removeItem: (itemId: number) =>
    api.delete(`/bmol/${itemId}`),
}
