import { computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import { mergeMangaQuery, parseMangaQuery, type MangaBrowseState } from "@/components/manga/mangaQuery"

/** The URL is the single source of truth for the manga filters. */
export function useMangaRouteState() {
  const route = useRoute()
  const router = useRouter()
  const state = computed<MangaBrowseState>(() => parseMangaQuery(route.query))

  // Always carry route.hash: OAuth and handoff tokens can arrive in the fragment.
  function update(patch: Partial<MangaBrowseState>) {
    const next = { ...state.value, ...patch }
    return router.replace({ query: mergeMangaQuery(route.query, next), hash: route.hash })
  }

  return { state, update }
}
