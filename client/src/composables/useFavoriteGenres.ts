import { computed } from "vue"
import { useAuthStore } from "@/stores/auth"

/**
 * `favorite_genres` is stored as a JSON array string, with legacy values
 * "null" / "skip" / a plain comma list. Exposes whether the user picked any
 * and the comma-joined id list the recommended endpoint expects.
 */
export function useFavoriteGenres() {
  const authStore = useAuthStore()

  const raw = computed(() => {
    const g = authStore.user?.favorite_genres
    return !g || g === "null" || g === "skip" ? "" : g
  })

  const hasGenres = computed(() => {
    if (!raw.value) return false
    try {
      const parsed: unknown = JSON.parse(raw.value)
      return Array.isArray(parsed) && parsed.length > 0
    } catch {
      return raw.value.length > 0
    }
  })

  const genreQuery = computed(() => {
    if (!hasGenres.value) return ""
    try {
      const parsed: unknown = JSON.parse(raw.value)
      if (Array.isArray(parsed)) return parsed.join(",")
    } catch {
      // legacy plain-text value, used as is
    }
    return raw.value
  })

  return { hasGenres, genreQuery }
}
