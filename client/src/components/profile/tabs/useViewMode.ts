import { ref, watch } from "vue"

export type ViewMode = "grid" | "list"

/** View toggle remembered per tab; storage failures (private mode) fall back to the default. */
export function useViewMode(tab: string, fallback: ViewMode = "grid") {
  const key = `popcorn-profile-view:${tab}`
  let initial = fallback
  try {
    const saved = localStorage.getItem(key)
    if (saved === "grid" || saved === "list") initial = saved
  } catch {
    // Storage unavailable.
  }
  const mode = ref<ViewMode>(initial)
  watch(mode, (v) => {
    try {
      localStorage.setItem(key, v)
    } catch {
      // Storage unavailable.
    }
  })
  return mode
}
