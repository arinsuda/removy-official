import { computed, ref, watch, type Ref } from "vue"
import { useInfiniteQuery, useQuery } from "@tanstack/vue-query"
import { MAX_PAGES, MIN_SEARCH_CHARS, type BrowseConfig } from "@/components/browse/browseConfig"
import { filterTier } from "@/components/browse/browseSort"
import { sortMedia } from "@/components/browse/browseSort"
import type { BrowseState } from "@/components/browse/browseQuery"
import type { HomeMedia } from "@/components/home/homeMedia"
import { countByTier, dedupeMedia } from "@/components/home/popcornHome"

/** With a tier selected, keep loading pages until this many titles match. */
export const TIER_MIN_MATCHES = 12
export const TIER_MAX_AUTO_PAGES = 5

export function useBrowseData(config: BrowseConfig, state: Ref<BrowseState>) {
  const searching = computed(() => state.value.q.length >= MIN_SEARCH_CHARS)

  const list = useInfiniteQuery({
    // Sort and tier are applied on the loaded pages, so they are not part of the key.
    queryKey: computed(() =>
      searching.value
        ? ["browse", config.type, "search", state.value.q]
        : ["browse", config.type, state.value.tab, state.value.genre, ""],
    ),
    initialPageParam: 1,
    queryFn: ({ pageParam }) =>
      searching.value
        ? config.fetchSearch(state.value.q, pageParam)
        : config.fetchList(state.value.tab, pageParam, state.value.genre),
    getNextPageParam: (last, pages) =>
      pages.length < MAX_PAGES && last.page < Math.min(last.total_pages, 500) ? last.page + 1 : undefined,
    staleTime: 60_000,
  })

  const genres = useQuery({
    queryKey: ["browse", config.type, "genres"],
    queryFn: config.fetchGenres,
    staleTime: Infinity,
  })

  const loaded = computed<HomeMedia[]>(() => {
    const items = dedupeMedia((list.data.value?.pages ?? []).flatMap((p) => p.results))
    const genre = state.value.genre
    return !config.genreServerSide && genre && !searching.value
      ? items.filter((m) => (m.source.genre_ids ?? []).includes(genre))
      : items
  })
  const tierCounts = computed(() => countByTier(loaded.value))
  const items = computed(() => sortMedia(filterTier(loaded.value, state.value.tier), state.value.sort))

  /** True when the page cap stopped a list that TMDB could still continue. */
  const capped = computed(() => {
    const pages = list.data.value?.pages ?? []
    const last = pages[pages.length - 1]
    return !!last && pages.length >= MAX_PAGES && last.page < Math.min(last.total_pages, 500)
  })

  // A tier can be rare on one page, so pull a few more pages until enough titles match.
  const autoPages = ref(0)
  watch(
    () => [state.value.tier, searching.value, state.value.tab, state.value.genre, state.value.q],
    () => (autoPages.value = 0),
  )
  watch(
    [items, () => list.hasNextPage.value, () => list.isFetching.value],
    () => {
      if (state.value.tier === "all" || list.isFetching.value || !list.hasNextPage.value) return
      if (items.value.length >= TIER_MIN_MATCHES || autoPages.value >= TIER_MAX_AUTO_PAGES) return
      autoPages.value++
      void list.fetchNextPage()
    },
    { flush: "post" },
  )

  return {
    items,
    loaded,
    tierCounts,
    searching,
    capped,
    genres: computed(() => genres.data.value ?? []),
    isLoading: list.isLoading,
    isError: list.isError,
    isFetchingNextPage: list.isFetchingNextPage,
    hasNextPage: list.hasNextPage,
    fetchNextPage: () => list.fetchNextPage(),
    refetch: () => list.refetch(),
  }
}
