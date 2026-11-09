import { describe, expect, it } from "vitest"
import { AVATAR_MAX_BYTES, MAX_ZOOM, clampCrop, cropRect, validateAvatarFile } from "@/components/settings/avatarCrop"

describe("validateAvatarFile", () => {
  it("accepts supported images up to 5 MB", () => {
    expect(validateAvatarFile({ type: "image/webp", size: 1000 })).toBeNull()
    expect(validateAvatarFile({ type: "image/png", size: AVATAR_MAX_BYTES })).toBeNull()
  })
  it("rejects other types and oversized files", () => {
    expect(validateAvatarFile({ type: "application/pdf", size: 10 })).toBe("type")
    expect(validateAvatarFile({ type: "image/svg+xml", size: 10 })).toBe("type")
    expect(validateAvatarFile({ type: "image/jpeg", size: AVATAR_MAX_BYTES + 1 })).toBe("size")
  })
})

describe("crop geometry", () => {
  const V = 280
  it("covers the shorter side at zoom 1 and centres the square", () => {
    const r = cropRect({ zoom: 1, x: 0, y: 0 }, 800, 400, V)
    expect(r.size).toBeCloseTo(400)
    expect(r.sx).toBeCloseTo(200)
    expect(r.sy).toBeCloseTo(0)
  })
  it("halves the source square at zoom 2", () => {
    expect(cropRect({ zoom: 2, x: 0, y: 0 }, 400, 400, V).size).toBeCloseTo(200)
  })
  it("clamps the offset so the image always covers the viewport", () => {
    // 800x400 at zoom 1 is drawn 560x280: only horizontal slack of 140px each way.
    const c = clampCrop({ zoom: 1, x: 999, y: 999 }, 800, 400, V)
    expect(c.x).toBeCloseTo(140)
    expect(c.y).toBe(0)
    const r = cropRect({ zoom: 1, x: 999, y: 0 }, 800, 400, V)
    expect(r.sx).toBeGreaterThanOrEqual(0)
    expect(r.sx + r.size).toBeLessThanOrEqual(800 + 1e-6)
  })
  it("limits zoom to 1..MAX_ZOOM", () => {
    expect(clampCrop({ zoom: 0.2, x: 0, y: 0 }, 400, 400, V).zoom).toBe(1)
    expect(clampCrop({ zoom: 99, x: 0, y: 0 }, 400, 400, V).zoom).toBe(MAX_ZOOM)
  })
})
