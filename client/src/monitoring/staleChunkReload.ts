import type { Router } from "vue-router"

// After a deploy, tabs opened before it still reference the old hashed chunks, which no longer
// exist on the CDN. Loading one fails ("Unable to preload CSS", "Failed to fetch dynamically
// imported module"). A full reload picks up the new index.html. The cooldown stops a reload loop
// when the asset is genuinely missing.
export const STALE_CHUNK_RELOAD_KEY = "popcorn:stale-chunk-reload-at"
export const STALE_CHUNK_RELOAD_COOLDOWN_MS = 10_000

const STALE_CHUNK_MESSAGES = [
  "Unable to preload CSS",
  "Failed to fetch dynamically imported module",
  "Importing a module script failed",
  "error loading dynamically imported module",
] as const

export function isStaleChunkError(error: unknown): boolean {
  const message = error instanceof Error ? error.message : String(error ?? "")
  return STALE_CHUNK_MESSAGES.some((needle) => message.includes(needle))
}

function readLastReload(): number {
  try {
    return Number(sessionStorage.getItem(STALE_CHUNK_RELOAD_KEY)) || 0
  } catch {
    return 0
  }
}

function markReload(now: number): void {
  try {
    sessionStorage.setItem(STALE_CHUNK_RELOAD_KEY, String(now))
  } catch {
    // Storage blocked: the cooldown is skipped, the reload still helps.
  }
}

/** Reloads to `path` once per cooldown window. Returns true when a reload was started. */
export function reloadForStaleChunk(path: string, now = Date.now()): boolean {
  if (now - readLastReload() < STALE_CHUNK_RELOAD_COOLDOWN_MS) return false
  markReload(now)
  window.location.assign(path)
  return true
}

export function installStaleChunkReload(router: Router): void {
  window.addEventListener("vite:preloadError", (event) => {
    if (reloadForStaleChunk(window.location.href)) event.preventDefault()
  })
  router.onError((error, to) => {
    if (isStaleChunkError(error)) reloadForStaleChunk(to.fullPath)
  })
}
