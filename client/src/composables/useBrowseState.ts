import { computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import {
  mergeBrowseQuery,
  parseBrowseQuery,
  type BrowseState,
} from "@/components/browse/browseQuery"
import type { BrowseConfig } from "@/components/browse/browseConfig"
import { browseConfigFor } from "@/components/browse/browseConfig"
import type { HomeMediaType } from "@/components/home/homeMedia"

/** The URL is the single source of truth for tab, genre, sort, tier and search. */
export function useBrowseState() {
  const route = useRoute()
  const router = useRouter()
  const state = computed<BrowseState>(() => parseBrowseQuery(route.query))

  // Always carry route.hash: OAuth and handoff tokens arrive in the fragment on /movies.
  function update(patch: Partial<BrowseState>) {
    return router.replace({ query: mergeBrowseQuery(route.query, { ...state.value, ...patch }), hash: route.hash })
  }

  /** Movie and TV genre ids differ, so the genre is dropped; everything else carries over. */
  function switchType(config: BrowseConfig, type: HomeMediaType) {
    if (config.type === type) return
    const target = browseConfigFor(type)
    return router.push({
      name: target.routeName,
      query: mergeBrowseQuery(route.query, { ...state.value, genre: null }),
      hash: route.hash,
    })
  }

  return { state, update, switchType }
}
