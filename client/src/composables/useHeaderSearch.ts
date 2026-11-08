import { computed, onBeforeUnmount, ref, watch } from "vue"
import { useRouter } from "vue-router"
import { movieApi } from "@/api/endpoints/movie"
import { getTmdbImageUrl } from "@/utils/image"
import type { Movie, TVSeries } from "@/types"
import type { SearchStatus, SearchSuggestion } from "@/components/search/types"

export const SEARCH_MIN_CHARS = 2
export const SEARCH_DEBOUNCE_MS = 300
const MAX_SUGGESTIONS = 6

const year = (d?: string | null) => (d ? d.slice(0, 4) : "")

function mapMovie(m: Movie): SearchSuggestion {
  return {
    id: m.id,
    type: "movie",
    title: m.title,
    poster_path: m.poster_path ? getTmdbImageUrl(m.poster_path, "w154") : null,
    year: year(m.release_date),
    rating: m.vote_average,
    ratings: m.ratings,
  }
}

function mapSeries(s: TVSeries): SearchSuggestion {
  return {
    id: s.id,
    type: "tv",
    title: s.name,
    poster_path: s.poster_path ? getTmdbImageUrl(s.poster_path, "w154") : null,
    year: year(s.first_air_date),
    rating: s.vote_average,
    ratings: s.ratings,
  }
}

/** Alternate movies and series so neither type crowds out the other. */
function interleave(movies: SearchSuggestion[], series: SearchSuggestion[]): SearchSuggestion[] {
  const out: SearchSuggestion[] = []
  for (let i = 0; i < Math.max(movies.length, series.length); i++) {
    if (movies[i]) out.push(movies[i]!)
    if (series[i]) out.push(series[i]!)
  }
  return out
}

/** State and keyboard logic for the header search box. */
export function useHeaderSearch() {
  const router = useRouter()
  const open = ref(false)
  const query = ref("")
  const status = ref<SearchStatus>("idle")
  const results = ref<SearchSuggestion[]>([])
  /** -1 = none; results.length = the "view all" row. */
  const activeIndex = ref(-1)
  let timer: ReturnType<typeof setTimeout> | null = null
  let requestId = 0

  const trimmed = computed(() => query.value.trim())
  const canSearch = computed(() => trimmed.value.length >= SEARCH_MIN_CHARS)
  const showPanel = computed(() => open.value && canSearch.value && status.value !== "idle")
  /** Selectable rows: every result plus the trailing "view all" row. */
  const optionCount = computed(() => (status.value === "success" && results.value.length ? results.value.length + 1 : 0))

  async function fetchSuggestions(q: string) {
    const id = ++requestId
    status.value = "loading"
    try {
      const [movieRes, seriesRes] = await Promise.all([movieApi.search(q, 1), movieApi.searchSeries(q, 1)])
      if (id !== requestId) return
      const merged = interleave(movieRes.data.results.map(mapMovie), seriesRes.data.results.map(mapSeries))
      results.value = merged.slice(0, MAX_SUGGESTIONS)
      status.value = "success"
    } catch {
      if (id !== requestId) return
      results.value = []
      status.value = "error"
    }
  }

  function reset() {
    if (timer) clearTimeout(timer)
    requestId++
    results.value = []
    status.value = "idle"
    activeIndex.value = -1
  }

  watch(query, () => {
    if (timer) clearTimeout(timer)
    activeIndex.value = -1
    if (!canSearch.value) {
      reset()
      return
    }
    status.value = "loading"
    timer = setTimeout(() => fetchSuggestions(trimmed.value), SEARCH_DEBOUNCE_MS)
  })

  function retry() {
    if (canSearch.value) fetchSuggestions(trimmed.value)
  }

  function close() {
    open.value = false
    query.value = ""
    reset()
  }

  function goToItem(item: SearchSuggestion) {
    router.push(item.type === "movie" ? `/movies/${item.id}` : `/tv/${item.id}`)
    close()
  }

  function viewAll() {
    if (!trimmed.value) return
    router.push({ name: "search-results", query: { q: trimmed.value } })
    close()
  }

  function move(delta: 1 | -1) {
    const n = optionCount.value
    if (!n) return
    activeIndex.value = activeIndex.value < 0 ? (delta === 1 ? 0 : n - 1) : (activeIndex.value + delta + n) % n
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault()
      move(e.key === "ArrowDown" ? 1 : -1)
    } else if (e.key === "Enter") {
      e.preventDefault()
      const item = results.value[activeIndex.value]
      if (item) goToItem(item)
      else viewAll()
    } else if (e.key === "Escape") {
      close()
    }
  }

  onBeforeUnmount(() => timer && clearTimeout(timer))

  return { open, query, status, results, activeIndex, trimmed, canSearch, showPanel, optionCount, retry, close, goToItem, viewAll, onKeydown }
}
