import { computed, type Ref } from "vue"
import { useInfiniteQuery, useQuery } from "@tanstack/vue-query"
import { movieApi } from "@/api/api"
import { MAX_PAGES } from "@/components/browse/browseConfig"
import { fromMovie, type HomeMedia } from "@/components/home/homeMedia"
import { dedupeMedia } from "@/components/home/popcornHome"
import { useLocale } from "@/i18n"

/** Thai releases from today on, soonest first; the server owns date filtering. */
export function useUpcomingData(genre: Ref<number | null>) {
  const { locale } = useLocale()

  const list = useInfiniteQuery({
    queryKey: computed(() => ["upcoming", genre.value, locale.value]),
    initialPageParam: 1,
    queryFn: async ({ pageParam }) => {
      const { data } = await movieApi.getUpcoming(pageParam, genre.value ?? undefined)
      return { ...data, results: (data.results ?? []).map(fromMovie) }
    },
    getNextPageParam: (last, pages) =>
      pages.length < MAX_PAGES && last.page < Math.min(last.total_pages, 500) ? last.page + 1 : undefined,
    staleTime: 60_000,
  })

  const genres = useQuery({
    queryKey: ["browse", "movie", "genres"],
    queryFn: async () => (await movieApi.getGenres()).data.genres,
    staleTime: Infinity,
  })

  const items = computed<HomeMedia[]>(() => dedupeMedia((list.data.value?.pages ?? []).flatMap((p) => p.results)))

  /** True when the page cap stopped a list TMDB could still continue. */
  const capped = computed(() => {
    const pages = list.data.value?.pages ?? []
    const last = pages[pages.length - 1]
    return !!last && pages.length >= MAX_PAGES && last.page < Math.min(last.total_pages, 500)
  })

  return {
    items,
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
