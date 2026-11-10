import { movieApi } from "@/api/api"
import type { SearchFilters } from "@/api/endpoints/movie"
import { fromMovie, fromSeries, type HomeMedia, type HomeMediaType } from "@/components/home/homeMedia"
import type { Genre, PaginatedResult, Movie, TVSeries } from "@/types"

export type BrowseTab = "popular" | "now" | "top"

export const BROWSE_TABS: BrowseTab[] = ["popular", "now", "top"]
/** Search kicks in at this many characters. */
export const MIN_SEARCH_CHARS = 2
/**
 * With a genre filter the API filters each TMDB page, so a page can be empty while more exist.
 * Keep fetching through at most this many consecutive empty pages before showing the empty state.
 */
export const SEARCH_MAX_EMPTY_PAGES = 3
/** Hard cap on pages pulled into one list (10 x 20 titles). */
export const MAX_PAGES = 10

export interface BrowseConfig {
  type: HomeMediaType
  routeName: "movies" | "tv"
  /** Whether the API filters by genre; when false the loaded pages are filtered here. */
  genreServerSide: boolean
  fetchList(tab: BrowseTab, page: number, genre: number | null): Promise<PaginatedResult<HomeMedia>>
  fetchSearch(q: string, page: number, filters?: SearchFilters): Promise<PaginatedResult<HomeMedia>>
  fetchGenres(): Promise<Genre[]>
}

function mapPage<T>(res: { data: PaginatedResult<T> }, map: (item: T) => HomeMedia): PaginatedResult<HomeMedia> {
  return { ...res.data, results: (res.data.results ?? []).map(map) }
}

export const movieBrowse: BrowseConfig = {
  type: "movie",
  routeName: "movies",
  genreServerSide: true,
  fetchList: async (tab, page, genre) => {
    const g = genre ?? undefined
    const call = { popular: movieApi.getPopular, now: movieApi.getNowPlaying, top: movieApi.getTopRated }[tab]
    return mapPage<Movie>(await call(page, g), fromMovie)
  },
  fetchSearch: async (q, page, filters) => mapPage<Movie>(await movieApi.search(q, page, filters), fromMovie),
  fetchGenres: async () => (await movieApi.getGenres()).data.genres,
}

export const tvBrowse: BrowseConfig = {
  type: "tv",
  routeName: "tv",
  genreServerSide: true,
  fetchList: async (tab, page, genre) => {
    const g = genre ?? undefined
    const call = { popular: movieApi.getPopularSeries, now: movieApi.getNowAiringSeries, top: movieApi.getTopRatedSeries }[tab]
    return mapPage<TVSeries>(await call(page, g), fromSeries)
  },
  fetchSearch: async (q, page, filters) => mapPage<TVSeries>(await movieApi.searchSeries(q, page, filters), fromSeries),
  fetchGenres: async () => (await movieApi.getSeriesGenres()).data.genres,
}

export const browseConfigFor = (type: HomeMediaType): BrowseConfig => (type === "tv" ? tvBrowse : movieBrowse)
