export const AVATAR_OUTPUT_SIZE = 512
export const AVATAR_MAX_BYTES = 5 * 1024 * 1024
export const AVATAR_ACCEPTED_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"]
export const MAX_ZOOM = 3

export type AvatarFileError = "type" | "size"

/** Returns an error code, or null when the file may be used as an avatar. */
export function validateAvatarFile(file: Pick<File, "type" | "size">): AvatarFileError | null {
  if (!AVATAR_ACCEPTED_TYPES.includes(file.type)) return "type"
  if (file.size > AVATAR_MAX_BYTES) return "size"
  return null
}

export interface CropState {
  /** 1 = the shorter image side exactly fills the viewport. */
  zoom: number
  /** Image centre offset from the viewport centre, in viewport pixels. */
  x: number
  y: number
}

/** Pixels the image is drawn per source pixel for a square viewport. */
export function cropScale(imgW: number, imgH: number, viewport: number, zoom: number): number {
  return (viewport / Math.min(imgW, imgH)) * zoom
}

/** Keeps the image covering the whole viewport. */
export function clampCrop(state: CropState, imgW: number, imgH: number, viewport: number): CropState {
  const zoom = Math.min(MAX_ZOOM, Math.max(1, state.zoom))
  const scale = cropScale(imgW, imgH, viewport, zoom)
  const maxX = Math.max(0, (imgW * scale - viewport) / 2)
  const maxY = Math.max(0, (imgH * scale - viewport) / 2)
  return { zoom, x: Math.min(maxX, Math.max(-maxX, state.x)), y: Math.min(maxY, Math.max(-maxY, state.y)) }
}

/** Square source rectangle (in image pixels) visible in the viewport. */
export function cropRect(state: CropState, imgW: number, imgH: number, viewport: number) {
  const s = clampCrop(state, imgW, imgH, viewport)
  const scale = cropScale(imgW, imgH, viewport, s.zoom)
  const size = viewport / scale
  return { sx: imgW / 2 - s.x / scale - size / 2, sy: imgH / 2 - s.y / scale - size / 2, size }
}

/** Draws the visible square into a 512px canvas and encodes it as WebP. */
export function cropToBlob(img: CanvasImageSource & { naturalWidth: number; naturalHeight: number }, state: CropState, viewport: number): Promise<Blob> {
  const { sx, sy, size } = cropRect(state, img.naturalWidth, img.naturalHeight, viewport)
  const canvas = document.createElement("canvas")
  canvas.width = canvas.height = AVATAR_OUTPUT_SIZE
  const ctx = canvas.getContext("2d")
  if (!ctx) return Promise.reject(new Error("canvas unavailable"))
  ctx.drawImage(img, sx, sy, size, size, 0, 0, AVATAR_OUTPUT_SIZE, AVATAR_OUTPUT_SIZE)
  return new Promise((resolve, reject) =>
    canvas.toBlob(blob => (blob ? resolve(blob) : reject(new Error("encode failed"))), "image/webp", 0.9),
  )
}
