import { computed, type Ref } from "vue"
import { useInfiniteQuery, useQuery } from "@tanstack/vue-query"
import { mangaApi } from "@/api/api"
import { MANGA_LIST_STALE_MS, MANGA_PAGE_CAP } from "@/components/manga/mangaConfig"
import { toMangaApiQuery, type MangaBrowseState } from "@/components/manga/mangaQuery"
import type { MangaItem } from "@/types/manga"

/** Drops repeats across pages (a title can shift between pages while the list changes). */
export function dedupeManga(items: MangaItem[]): MangaItem[] {
  const seen = new Set<number>()
  return items.filter((item) => (seen.has(item.id) ? false : (seen.add(item.id), true)))
}

/**
 * Manga list for the browse page. Every filter is server side, so the query key carries them
 * and each change starts a fresh list.
 */
export function useMangaBrowse(state: Ref<MangaBrowseState>, includeAdult: Ref<boolean>) {
  const list = useInfiniteQuery({
    queryKey: computed(() => ["manga", "list", toMangaApiQuery(state.value, includeAdult.value)]),
    initialPageParam: 1,
    queryFn: async ({ pageParam }) => (await mangaApi.list(toMangaApiQuery(state.value, includeAdult.value), pageParam)).data,
    getNextPageParam: (last, pages) =>
      pages.length < MANGA_PAGE_CAP && last.page < last.total_pages ? last.page + 1 : undefined,
    staleTime: MANGA_LIST_STALE_MS,
  })

  const genres = useQuery({
    queryKey: ["manga", "genres"],
    queryFn: async () => (await mangaApi.genres()).data.genres,
    staleTime: Infinity,
  })

  const loaded = computed<MangaItem[]>(() => dedupeManga((list.data.value?.pages ?? []).flatMap((p) => p.results)))

  /** True when the page cap stopped a list that the API could still continue. */
  const capped = computed(() => {
    const pages = list.data.value?.pages ?? []
    const last = pages[pages.length - 1]
    return !!last && pages.length >= MANGA_PAGE_CAP && last.page < last.total_pages
  })

  return {
    loaded,
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
