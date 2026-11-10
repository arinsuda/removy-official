import { computed, ref, watch, type Ref } from "vue"
import { useInfiniteQuery, useQuery } from "@tanstack/vue-query"
import { actorApi } from "@/api/api"
import type { SearchFilters } from "@/api/endpoints/movie"
import { MAX_PAGES, SEARCH_MAX_EMPTY_PAGES, movieBrowse, tvBrowse } from "@/components/browse/browseConfig"
import { sortMedia } from "@/components/browse/browseSort"
import type { HomeMedia } from "@/components/home/homeMedia"
import { dedupeMedia, interleave } from "@/components/home/popcornHome"
import type { SearchState } from "@/components/search/searchQuery"
import type { Genre, PaginatedResult, PersonSearchResult } from "@/types"

export function useSearchResults(state: Ref<SearchState>) {
  const q = computed(() => state.value.q)
  const enabled = computed(() => q.value.length > 0)

  /** Year and genre are applied by the API, so changing them restarts the lists at page 1. */
  const filters = computed<SearchFilters>(() => ({ year: state.value.year, genre: state.value.genre }))

  /** Pages at the end of the list that came back empty (genre filtering can empty a whole page). */
  const trailingEmptyPages = (pages: { results?: unknown[] }[] = []) => {
    let n = 0
    for (let i = pages.length - 1; i >= 0 && !(pages[i]!.results ?? []).length; i--) n++
    return n
  }

  function paged<T>(
    key: string,
    fetch: (q: string, page: number, filters: SearchFilters) => Promise<PaginatedResult<T>>,
    { filtered = true, active = enabled }: { filtered?: boolean; active?: Ref<boolean> } = {},
  ) {
    const query = useInfiniteQuery({
      queryKey: computed(() => ["search", key, q.value, ...(filtered ? [filters.value.year, filters.value.genre] : [])]),
      initialPageParam: 1,
      queryFn: ({ pageParam }) => fetch(q.value, pageParam, filtered ? filters.value : {}),
      getNextPageParam: (last, pages) =>
        pages.length < MAX_PAGES && last.page < Math.min(last.total_pages, 500) ? last.page + 1 : undefined,
      enabled: active,
      staleTime: 60_000,
    })
    // Skip over empty filtered pages instead of stalling, up to a bound.
    if (filtered) {
      watch(
        () => [query.data.value?.pages.length, query.isFetching.value] as const,
        () => {
          const empties = trailingEmptyPages(query.data.value?.pages)
          if (!filters.value.genre || query.isFetching.value || !query.hasNextPage.value) return
          if (empties > 0 && empties < SEARCH_MAX_EMPTY_PAGES) void query.fetchNextPage()
        },
      )
    }
    return query
  }

  const genreQuery = (type: "movie" | "tv", fn: () => Promise<Genre[]>) =>
    useQuery({ queryKey: ["browse", type, "genres"], queryFn: fn, staleTime: Infinity })
  const movieGenres = genreQuery("movie", movieBrowse.fetchGenres)
  const tvGenres = genreQuery("tv", tvBrowse.fetchGenres)

  /** A list is skipped when its genre catalog lacks the selected id (ids differ between movies and TV). */
  const genreFits = (catalog: typeof movieGenres) =>
    computed(() => {
      const genre = state.value.genre
      if (!genre) return true
      const list = catalog.data.value
      return list ? list.some((g) => g.id === genre) : !catalog.isLoading.value
    })
  const movieFits = genreFits(movieGenres)
  const tvFits = genreFits(tvGenres)

  const movies = paged("movie", movieBrowse.fetchSearch, { active: computed(() => enabled.value && movieFits.value) })
  const series = paged("tv", tvBrowse.fetchSearch, { active: computed(() => enabled.value && tvFits.value) })
  const people = paged<PersonSearchResult>("people", async (query, page) => (await actorApi.search(query, page)).data, { filtered: false })

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

  // Year options are the years seen so far for this query. The lists are already filtered by the
  // API, so they must not shrink to the selected year; the selected year is always kept.
  const seenYears = ref<number[]>([])
  watch(q, () => (seenYears.value = []))
  watch(
    pool,
    (items) => {
      const next = new Set(seenYears.value)
      for (const m of items) {
        const y = Number(m.date.slice(0, 4))
        if (y > 0) next.add(y)
      }
      if (next.size !== seenYears.value.length) seenYears.value = [...next]
    },
    { immediate: true },
  )
  const years = computed(() => [...new Set([...seenYears.value, ...(state.value.year ? [state.value.year] : [])])].sort((a, b) => b - a))

  /** Every TMDB genre of the active tab (the API filters by id, so options no longer depend on loaded titles). */
  const genres = computed(() => {
    const { tab } = state.value
    const source = tab === "movie" ? movieGenres.data.value : tab === "tv" ? tvGenres.data.value : [...(movieGenres.data.value ?? []), ...(tvGenres.data.value ?? [])]
    const names = new Map<number, string>()
    for (const g of source ?? []) names.set(g.id, g.name)
    return [...names].map(([id, name]) => ({ id, name })).sort((a, b) => a.name.localeCompare(b.name))
  })

  const items = computed(() => sortMedia(pool.value, state.value.sort))

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

  /** Counts of filtered lists are TMDB's unfiltered totals when a genre is set, so the UI hides them. */
  const countsApproximate = computed(() => !!state.value.genre)

  return { items, peopleItems, years, genres, counts, countsApproximate, busy, failed, hasNext, fetchingNext, fetchNext, retry }
}
