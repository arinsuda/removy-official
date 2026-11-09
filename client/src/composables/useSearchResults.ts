import { computed, type Ref } from "vue"
import { useInfiniteQuery, useQuery } from "@tanstack/vue-query"
import { actorApi } from "@/api/api"
import { MAX_PAGES, movieBrowse, tvBrowse } from "@/components/browse/browseConfig"
import { sortMedia } from "@/components/browse/browseSort"
import type { HomeMedia } from "@/components/home/homeMedia"
import { dedupeMedia, interleave } from "@/components/home/popcornHome"
import type { SearchState } from "@/components/search/searchQuery"
import type { Genre, PaginatedResult, PersonSearchResult } from "@/types"

export function useSearchResults(state: Ref<SearchState>) {
  const q = computed(() => state.value.q)
  const enabled = computed(() => q.value.length > 0)

  function paged<T>(key: string, fetch: (q: string, page: number) => Promise<PaginatedResult<T>>) {
    return useInfiniteQuery({
      queryKey: computed(() => ["search", key, q.value]),
      initialPageParam: 1,
      queryFn: ({ pageParam }) => fetch(q.value, pageParam),
      getNextPageParam: (last, pages) =>
        pages.length < MAX_PAGES && last.page < Math.min(last.total_pages, 500) ? last.page + 1 : undefined,
      enabled,
      staleTime: 60_000,
    })
  }

  const movies = paged("movie", movieBrowse.fetchSearch)
  const series = paged("tv", tvBrowse.fetchSearch)
  const people = paged<PersonSearchResult>("people", async (query, page) => (await actorApi.search(query, page)).data)

  const genreQuery = (type: "movie" | "tv", fn: () => Promise<Genre[]>) =>
    useQuery({ queryKey: ["browse", type, "genres"], queryFn: fn, staleTime: Infinity })
  const movieGenres = genreQuery("movie", movieBrowse.fetchGenres)
  const tvGenres = genreQuery("tv", tvBrowse.fetchGenres)

  const flat = <T>(pages: PaginatedResult<T>[] | undefined) => (pages ?? []).flatMap((p) => p.results ?? [])
  const movieItems = computed(() => dedupeMedia(flat<HomeMedia>(movies.data.value?.pages)))
  const seriesItems = computed(() => dedupeMedia(flat<HomeMedia>(series.data.value?.pages)))
  const peopleItems = computed(() => flat<PersonSearchResult>(people.data.value?.pages))

  /** Titles of the active tab before year/genre filters; "all" alternates movies and series. */
  const pool = computed(() => {
    const { tab } = state.value
    if (tab === "movie") return movieItems.value
    if (tab === "tv") return seriesItems.value
    return interleave(movieItems.value, seriesItems.value)
  })

  const years = computed(() =>
    [...new Set(pool.value.map((m) => Number(m.date.slice(0, 4))).filter((y) => y > 0))].sort((a, b) => b - a),
  )

  /** Genres present in the loaded titles, with TMDB names. */
  const genres = computed(() => {
    const names = new Map<number, string>()
    for (const g of [...(movieGenres.data.value ?? []), ...(tvGenres.data.value ?? [])]) names.set(g.id, g.name)
    const present = new Set(pool.value.flatMap((m) => m.source.genre_ids ?? []))
    return [...present].filter((id) => names.has(id)).map((id) => ({ id, name: names.get(id)! })).sort((a, b) => a.name.localeCompare(b.name))
  })

  const items = computed(() => {
    const { year, genre, sort } = state.value
    const filtered = pool.value.filter(
      (m) => (!year || m.date.startsWith(String(year))) && (!genre || (m.source.genre_ids ?? []).includes(genre)),
    )
    return sortMedia(filtered, sort)
  })

  const total = (r: { data: Ref<{ pages: { total_results: number }[] } | undefined> }) => r.data.value?.pages[0]?.total_results ?? 0
  const counts = computed(() => ({
    movie: total(movies),
    tv: total(series),
    people: total(people),
    all: total(movies) + total(series),
  }))

  const busy = computed(() =>
    state.value.tab === "people"
      ? people.isLoading.value
      : state.value.tab === "movie"
        ? movies.isLoading.value
        : state.value.tab === "tv"
          ? series.isLoading.value
          : movies.isLoading.value || series.isLoading.value,
  )
  // "all" only fails when both lists fail; one working list is still useful.
  const failed = computed(() =>
    state.value.tab === "people"
      ? people.isError.value
      : state.value.tab === "movie"
        ? movies.isError.value
        : state.value.tab === "tv"
          ? series.isError.value
          : movies.isError.value && series.isError.value,
  )
  const lists = computed(() => (state.value.tab === "people" ? [people] : state.value.tab === "movie" ? [movies] : state.value.tab === "tv" ? [series] : [movies, series]))
  const hasNext = computed(() => lists.value.some((l) => l.hasNextPage.value))
  const fetchingNext = computed(() => lists.value.some((l) => l.isFetchingNextPage.value))

  function fetchNext() {
    for (const l of lists.value) if (l.hasNextPage.value && !l.isFetchingNextPage.value) void l.fetchNextPage()
  }
  function retry() {
    for (const l of lists.value) void l.refetch()
  }

  return { items, peopleItems, years, genres, counts, busy, failed, hasNext, fetchingNext, fetchNext, retry }
}
