import { computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import {
  applyAnimeChange,
  mergeAnimeQuery,
  parseAnimeQuery,
  type AnimeBrowseState,
} from "@/components/anime/animeQuery"

/** The URL is the single source of truth for the anime filters. */
export function useAnimeRouteState() {
  const route = useRoute()
  const router = useRouter()
  const state = computed<AnimeBrowseState>(() => parseAnimeQuery(route.query, new Date().getFullYear()))

  // Always carry route.hash: OAuth and handoff tokens can arrive in the fragment.
  function update(patch: Partial<AnimeBrowseState>) {
    const next = applyAnimeChange(state.value, patch)
    return router.replace({ query: mergeAnimeQuery(route.query, next), hash: route.hash })
  }

  return { state, update }
}
