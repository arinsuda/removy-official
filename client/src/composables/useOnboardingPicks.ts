import { computed, ref } from "vue"
import { movieApi, userApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { MIN_PICKS, extractGenres, pickDiverse, rankGenres } from "@/components/onboarding/onboardingConfig"
import type { Movie } from "@/types"

export type PicksStatus = "loading" | "ready" | "error"

/** State and API calls behind the onboarding page; navigation stays in the view. */
export function useOnboardingPicks() {
  const authStore = useAuthStore()
  const movies = ref<Movie[]>([])
  const selected = ref<Movie[]>([])
  const status = ref<PicksStatus>("loading")
  const saving = ref(false)
  const saveFailed = ref(false)

  const selectedIds = computed(() => new Set(selected.value.map((m) => m.id)))
  const count = computed(() => selected.value.length)
  const remaining = computed(() => Math.max(MIN_PICKS - count.value, 0))
  const canContinue = computed(() => remaining.value === 0 && !saving.value)
  const selectedGenreIds = computed(() => rankGenres(selected.value, 6))

  async function load() {
    status.value = "loading"
    try {
      const [pop1, pop2, top] = await Promise.all([
        movieApi.getPopular(1).then((r) => r.data.results as Movie[]),
        movieApi.getPopular(2).then((r) => r.data.results as Movie[]),
        movieApi.getTopRated(1).then((r) => r.data.results as Movie[]),
      ])
      movies.value = pickDiverse([...pop1, ...pop2, ...top])
      status.value = movies.value.length ? "ready" : "error"
    } catch {
      status.value = "error"
    }
  }

  function toggle(movie: Movie) {
    if (saving.value) return
    selected.value = selectedIds.value.has(movie.id)
      ? selected.value.filter((m) => m.id !== movie.id)
      : [...selected.value, movie]
  }

  /** Saves the picks; resolves true when the user may leave the page. */
  async function confirm(): Promise<boolean> {
    if (!canContinue.value || !authStore.user) return false
    saving.value = true
    saveFailed.value = false
    try {
      const genres = extractGenres(selected.value)
      await userApi.updateFavoriteGenres(authStore.user.id, genres)
      authStore.setUser({ ...authStore.user, favorite_genres: JSON.stringify(genres) })
      return true
    } catch {
      saveFailed.value = true
      return false
    } finally {
      saving.value = false
    }
  }

  /** Skipping never blocks the user, even if the save fails. */
  async function skip() {
    if (saving.value || !authStore.user) return
    saving.value = true
    try {
      await userApi.updateFavoriteGenres(authStore.user.id, [])
    } catch {
      // ignored on purpose
    }
    authStore.setUser({ ...authStore.user, onboarded_at: new Date().toISOString() })
    saving.value = false
  }

  return { movies, selected, selectedIds, selectedGenreIds, status, saving, saveFailed, count, remaining, canContinue, load, toggle, confirm, skip }
}
