import { computed, type Ref } from "vue"
import { useInfiniteQuery, useQuery } from "@tanstack/vue-query"
import { animeApi, movieApi } from "@/api/api"
import { ANIME_EXCLUDED_GENRE, ANIME_LIST_STALE_MS, ANIME_PAGE_CAP } from "@/components/anime/animeConfig"
import { toAnimeApiQuery, type AnimeBrowseState } from "@/components/anime/animeQuery"
import { fromMovie, fromSeries, type HomeMedia, type HomeMediaType } from "@/components/home/homeMedia"
import { countByTier, dedupeMedia } from "@/components/home/popcornHome"
import type { Genre, PaginatedResult } from "@/types"

/** TMDB stops paging after this page. */
const TMDB_MAX_PAGE = 500

async function fetchAnimePage(
  type: HomeMediaType,
  state: AnimeBrowseState,
  page: number,
): Promise<PaginatedResult<HomeMedia>> {
  const query = toAnimeApiQuery(state)
  if (type === "tv") {
    const res = await animeApi.list("tv", query, page)
    return { ...res.data, results: res.data.results.map(fromSeries) }
  }
  const res = await animeApi.list("movie", query, page)
  return { ...res.data, results: res.data.results.map(fromMovie) }
}

async function fetchAnimeGenres(type: HomeMediaType): Promise<Genre[]> {
  const res = type === "tv" ? await movieApi.getSeriesGenres() : await movieApi.getGenres()
  return res.data.genres.filter((g) => g.id !== ANIME_EXCLUDED_GENRE)
}

/**
 * Anime list for the browse page. Sort, genre, keyword, country and release filters are all
 * server side, so the query key carries them and each change starts a fresh list.
 */
export function useAnimeBrowse(state: Ref<AnimeBrowseState>) {
  const list = useInfiniteQuery({
    queryKey: computed(() => ["anime", "list", state.value.type, toAnimeApiQuery(state.value)]),
    initialPageParam: 1,
    queryFn: ({ pageParam }) => fetchAnimePage(state.value.type, state.value, pageParam),
    getNextPageParam: (last, pages) =>
      pages.length < ANIME_PAGE_CAP && last.page < Math.min(last.total_pages, TMDB_MAX_PAGE) ? last.page + 1 : undefined,
    staleTime: ANIME_LIST_STALE_MS,
  })

  const genres = useQuery({
    queryKey: computed(() => ["anime", "genres", state.value.type]),
    queryFn: () => fetchAnimeGenres(state.value.type),
    staleTime: Infinity,
  })

  const loaded = computed<HomeMedia[]>(() => dedupeMedia((list.data.value?.pages ?? []).flatMap((p) => p.results)))

  /** True when the page cap stopped a list that TMDB could still continue. */
  const capped = computed(() => {
    const pages = list.data.value?.pages ?? []
    const last = pages[pages.length - 1]
    return !!last && pages.length >= ANIME_PAGE_CAP && last.page < Math.min(last.total_pages, TMDB_MAX_PAGE)
  })

  return {
    loaded,
    tierCounts: computed(() => countByTier(loaded.value)),
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
