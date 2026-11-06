import { ref } from "vue"
import { toPng } from "html-to-image"
import { useToast } from "@/composables/useToast"
import { useI18n } from "vue-i18n"
import type { MediaType } from "@/types/common"

/** Build API base URL matching the same logic as @/api/index.ts */
function getApiBaseUrl(): string {
  const envUrl = import.meta.env.VITE_API_BASE_URL
  if (!envUrl) return "/api"
  const cleanUrl = (envUrl as string).replace(/\/+$/, "")
  return cleanUrl.endsWith("/api") ? cleanUrl : `${cleanUrl}/api`
}

export function useShareImage() {
  let t: (key: string) => string
  try {
    const i18n = useI18n()
    t = i18n.t
  } catch {
    t = (key: string) => key
  }

  const toast = useToast()

  const isExporting = ref(false)
  const blobUrlsToRevoke: string[] = []
  let hasFallback = false

  /** Fetch a remote image as a Data URI to bypass CORS canvas taint and avoid console errors */
  async function fetchImageAsBlob(url: string): Promise<string | null> {
    try {
      if (!url) return null

      // If it's already a blob URL or data URI, return as-is
      if (url.startsWith("blob:") || url.startsWith("data:")) {
        return url
      }

      // 1. Try direct CORS fetch first (TMDB image.tmdb.org supports CORS out-of-the-box!)
      let response = await fetch(url, { mode: "cors" }).catch(() => null)

      // 2. If direct CORS fetch failed and it's an external HTTP/HTTPS image, fall back to backend proxy
      if ((!response || !response.ok) && (url.startsWith("http://") || url.startsWith("https://"))) {
        const currentOrigin = window.location.origin
        if (!url.startsWith(currentOrigin)) {
          const apiBase = getApiBaseUrl()
          const proxyUrl = `${apiBase}/share/proxy-image?url=${encodeURIComponent(url)}`
          response = await fetch(proxyUrl).catch(() => null)
        }
      }

      if (!response || !response.ok) return null

      const blob = await response.blob()
      return new Promise<string | null>((resolve) => {
        const reader = new FileReader()
        reader.onloadend = () => resolve(reader.result as string)
        reader.onerror = () => resolve(null)
        reader.readAsDataURL(blob)
      })
    } catch {
      return null
    }
  }

  /** Wait for all images inside a node to finish loading */
  async function waitForImages(node: HTMLElement): Promise<void> {
    const images = node.querySelectorAll("img")
    const promises: Promise<void>[] = []
    images.forEach((img) => {
      if (img.complete && img.naturalWidth > 0) return
      promises.push(
        img.decode().catch(() => {
          /* fallback image or broken — ignore */
        }),
      )
    })
    await Promise.all(promises)
  }

  /** Generate a PNG blob from a DOM node (1080×1920) */
  async function generatePng(node: HTMLElement): Promise<Blob | null> {
    if (isExporting.value) return null
    isExporting.value = true
    try {
      await document.fonts.ready
      await waitForImages(node)

      const dataUrl = await toPng(node, {
        pixelRatio: 1,
        width: 1080,
        height: 1920,
        cacheBust: false,
      })

      const response = await fetch(dataUrl)
      return await response.blob()
    } catch {
      toast.error(t("share.failedToGenerate"))
      return null
    } finally {
      isExporting.value = false
    }
  }

  /** Download blob or trigger native mobile share sheet for direct 'Save to Gallery/Photos' */
  async function downloadBlob(blob: Blob, filename: string): Promise<void> {
    try {
      const file = new File([blob], filename, { type: blob.type || "image/png" })

      // Check if Web Share API with files support is available (Mobile iOS / Android)
      if (
        navigator.canShare &&
        navigator.canShare({ files: [file] }) &&
        /Android|iPhone|iPad|iPod/i.test(navigator.userAgent)
      ) {
        await navigator.share({
          files: [file],
          title: filename,
        })
        toast.success(t("share.downloadSuccess"))
        return
      }
    } catch (err: any) {
      // User dismissed native share sheet — handle gracefully
      if (err?.name === "AbortError") return
    }

    // Fallback standard browser download (Desktop PC / fallback)
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = filename
    a.style.display = "none"
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    setTimeout(() => URL.revokeObjectURL(url), 1000)
    toast.success(t("share.downloadSuccess"))
  }

  /** Copy text to clipboard with toast feedback */
  async function copyText(text: string, successKey: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(text)
      toast.success(t(successKey))
    } catch {
      toast.error(t("share.clipboardFailed"))
    }
  }

  /** Create a safe filename from a title */
  function sanitizeFilename(title: string, prefix: string): string {
    const sanitized = title
      .replace(/[^a-zA-Z0-9\u0E00-\u0E7F_-]/g, "_")
      .replace(/_+/g, "_")
      .replace(/^_|_$/g, "")
      .slice(0, 50)
    return `${prefix}_${sanitized || "share"}.png`
  }

  /** Deterministic Unicode-safe review truncation */
  function truncateReview(text: string, maxChars: number = 200): string {
    if (text.length <= maxChars) return text

    // Use Array.from to handle surrogate pairs correctly
    const chars = Array.from(text)
    if (chars.length <= maxChars) return text

    const truncated = chars.slice(0, maxChars).join("")

    // Try to find last space for word boundary (works for English)
    const lastSpace = truncated.lastIndexOf(" ")
    if (lastSpace > maxChars * 0.6) {
      return truncated.slice(0, lastSpace) + "\u2026"
    }

    // For Thai or no-space text, truncate at char limit directly
    return truncated + "\u2026"
  }

  /** Build a caption for poster share */
  function buildPosterCaption(
    title: string,
    year: string | null,
    mediaType: MediaType,
  ): string {
    const typeLabel = mediaType === "movie" ? "Movie" : "TV Series"
    const yearStr = year ? ` (${year})` : ""
    return `🎬 ${title}${yearStr}\n${typeLabel}\n\nShared from POPCORN ✨`
  }

  /** Build a caption for review share */
  function buildReviewCaption(
    title: string,
    authorName: string,
    rating: number,
  ): string {
    return `⭐ ${rating.toFixed(1)}/5 — ${title}\n📝 Reviewed on POPCORN • ${authorName}\n\nShared from POPCORN ✨`
  }

  /** Build a share link for the media detail page */
  function buildShareLink(mediaType: MediaType, mediaId: number): string {
    const base = window.location.origin
    const path = mediaType === "movie" ? "movies" : "tv"
    return `${base}/${path}/${mediaId}`
  }

  /** Mark that a fallback was used for an image */
  function markFallbackUsed(): void {
    hasFallback = true
  }

  /** Show fallback warning if any images failed */
  function showFallbackWarningIfNeeded(): void {
    if (hasFallback) {
      toast.warning(t("share.imageFallbackUsed"))
    }
  }

  /** Reset fallback tracking */
  function resetFallbackTracking(): void {
    hasFallback = false
  }

  /** Revoke all blob URLs and reset state */
  function cleanup(): void {
    blobUrlsToRevoke.forEach((url) => URL.revokeObjectURL(url))
    blobUrlsToRevoke.length = 0
    isExporting.value = false
    hasFallback = false
  }

  return {
    isExporting,
    fetchImageAsBlob,
    generatePng,
    downloadBlob,
    copyText,
    sanitizeFilename,
    truncateReview,
    buildPosterCaption,
    buildReviewCaption,
    buildShareLink,
    markFallbackUsed,
    showFallbackWarningIfNeeded,
    resetFallbackTracking,
    cleanup,
  }
}
