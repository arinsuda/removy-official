// client/src/utils/image.ts

const TMDB_IMAGE_BASE = "https://image.tmdb.org/t/p";
const ABSOLUTE_URL = /^https?:\/\//i;
/** Hosts an absolute image URL may come from: AniList covers and the TMDB CDN. Any other host gets no image. */
const ALLOWED_IMAGE_HOSTS: readonly string[] = ["s4.anilist.co", "image.tmdb.org"];

export type TmdbImageSize =
  | "w92"
  | "w154"
  | "w185"
  | "w342"
  | "w500"
  | "w780"
  | "w1280"
  | "original";

/** True for an https URL on one of ALLOWED_IMAGE_HOSTS. */
function isAllowedImageUrl(src: string): boolean {
  try {
    const url = new URL(src);
    return url.protocol === "https:" && ALLOWED_IMAGE_HOSTS.includes(url.hostname);
  } catch {
    return false;
  }
}

/**
 * Image URL for a media summary. An absolute URL passes through only when it is
 * https on an allowed host (AniList covers, TMDB CDN); any other absolute URL is
 * treated as no image. Other values are TMDB paths such as "/abc123.jpg" and get
 * the TMDB base and size.
 * @returns the URL, or null when there is no usable image
 */
export function mediaImage(
  src: string | null | undefined,
  size: TmdbImageSize = "w342",
): string | null {
  if (!src) return null;
  if (ABSOLUTE_URL.test(src)) return isAllowedImageUrl(src) ? src : null;
  return `${TMDB_IMAGE_BASE}/${size}${src}`;
}

/**
 * สร้าง URL เต็มสำหรับรูปจาก TMDB
 * @param path poster_path หรือ backdrop_path ที่ได้จาก API (เช่น "/abc123.jpg")
 * @param size ขนาดรูปที่ต้องการ (default: w342)
 * @returns URL เต็ม หรือ null ถ้าไม่มี path
 */
export function getTmdbImageUrl(
  path: string | null | undefined,
  size: TmdbImageSize = "w342",
): string | null {
  return mediaImage(path, size);
}

/**
 * เหมือน getTmdbImageUrl แต่ return placeholder แทน null
 * ใช้ตรงๆ กับ :src ได้เลยโดยไม่ต้องเช็ค null เอง
 */
export function getTmdbImageUrlOrPlaceholder(
  path: string | null | undefined,
  size: TmdbImageSize = "w342",
): string {
  return mediaImage(path, size) ?? "/placeholder.jpg";
}
