import { ref } from "vue"
import { toPng } from "html-to-image"
import { useToast } from "@/composables/useToast"
import { useI18n } from "vue-i18n"
import type { MediaType } from "@/types/common"

/** Build API base URL matching the same logic as @/api/index.ts */
export function getApiBaseUrl(): string {
  const envUrl = import.meta.env.VITE_API_BASE_URL
  if (!envUrl) return "/api"
  const cleanUrl = (envUrl as string).replace(/\/+$/, "")
  return cleanUrl.endsWith("/api") ? cleanUrl : `${cleanUrl}/api`
}

/**
 * Turn an avatar/asset path stored by the API (e.g. "/uploads/a.png") into an absolute URL
 * on the API origin. Absolute http(s) and data/blob URLs are returned unchanged.
 */
export function resolveApiAssetUrl(raw: string | null | undefined): string {
  const value = raw?.trim() ?? ""
  if (!value || /^(https?:|data:|blob:)/.test(value)) return value
  const apiBase = getApiBaseUrl()
  const origin = apiBase.startsWith("http") ? new URL(apiBase).origin : window.location.origin
  return `${origin}${value.startsWith("/") ? "" : "/"}${value}`
}

export function isTmdbImageUrl(url: string): boolean {
  try {
    return new URL(url).hostname === "image.tmdb.org"
  } catch {
    return false
  }
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

  /**
   * Fetch an image as a data URL so the export canvas is never tainted.
   * TMDB images always go through the API proxy (never a direct CORS request to
   * image.tmdb.org); same-origin and API-hosted images are fetched directly.
   * Returns null on any failure.
   */
  async function fetchImageAsBlob(url: string): Promise<string | null> {
    try {
      if (!url) return null
      if (url.startsWith("blob:") || url.startsWith("data:")) return url

      const fetchUrl = isTmdbImageUrl(url)
        ? `${getApiBaseUrl()}/share/proxy-image?url=${encodeURIComponent(url)}`
        : url
      const response = await fetch(fetchUrl).catch(() => null)
      if (!response || !response.ok) return null

      const blob = await response.blob()
      return await new Promise<string | null>((resolve) => {
        const reader = new FileReader()
        reader.onloadend = () => resolve(typeof reader.result === "string" ? reader.result : null)
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

  /**
   * Rasterise a card at its export size. Fonts and images are settled first so
   * Thai text and posters are present. Cards with no background export transparent.
   */
  async function renderPng(
    node: HTMLElement,
    opts: { width: number; height: number; pixelRatio?: number },
  ): Promise<Blob | null> {
    if (isExporting.value) return null
    isExporting.value = true
    try {
      await document.fonts.ready
      // Make sure the Thai face is requested even if nothing on screen used it yet.
      await document.fonts.load('700 48px "Noto Sans Thai"', node.textContent ?? "").catch(() => [])
      await waitForImages(node)

      const dataUrl = await toPng(node, {
        pixelRatio: opts.pixelRatio ?? 1,
        width: opts.width,
        height: opts.height,
        cacheBust: false,
        // The preview scales the card with a CSS transform; export at full size.
        style: { transform: "none", margin: "0" },
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

  /** Generate a PNG blob from a DOM node (1080×1920) */
  function generatePng(node: HTMLElement): Promise<Blob | null> {
    return renderPng(node, { width: 1080, height: 1920 })
  }

  /** True when the browser can hand an image file to the native share sheet. */
  function canShareFiles(): boolean {
    try {
      if (typeof navigator.share !== "function" || typeof navigator.canShare !== "function") return false
      return navigator.canShare({ files: [new File([""], "share.png", { type: "image/png" })] })
    } catch {
      return false
    }
  }

  /** Open the native share sheet with the image. Returns false when dismissed or unavailable. */
  async function shareImageFile(blob: Blob, filename: string, text: string): Promise<boolean> {
    if (!canShareFiles()) return false
    try {
      await navigator.share({ files: [new File([blob], filename, { type: "image/png" })], title: "POPCORN", text })
      return true
    } catch (err) {
      // Dismissing the sheet is not an error.
      if ((err as { name?: string })?.name !== "AbortError") toast.error(t("share.nativeShareFailed"))
      return false
    }
  }

  /** True when the async clipboard can take an image. */
  function canCopyImage(): boolean {
    return typeof ClipboardItem !== "undefined" && typeof navigator.clipboard?.write === "function"
  }

  async function copyImage(blob: Blob): Promise<boolean> {
    if (!canCopyImage()) {
      toast.error(t("share.copyImageUnsupported"))
      return false
    }
    try {
      await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })])
      toast.success(t("share.imageCopied"))
      return true
    } catch {
      toast.error(t("share.clipboardFailed"))
      return false
    }
  }

  /** Plain browser download; never opens the share sheet. */
  function saveBlob(blob: Blob, filename: string): void {
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
    renderPng,
    canShareFiles,
    shareImageFile,
    canCopyImage,
    copyImage,
    saveBlob,
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
