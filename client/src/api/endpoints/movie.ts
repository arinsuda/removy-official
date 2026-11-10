import api from "../index"
import type {
  PaginatedResult,
  Movie,
  MovieDetail,
  Credits,
  Video,
  Genre,
  TVSeries,
  TVSeriesDetail,
  PersonDetail,
  PersonSearchResult,
  PersonMovieCredits,
  PersonTVCredits,
  WatchProviders,
} from "@/types"

/** Optional server-side filters of GET /movies/search and /tv/search (null or 0 = off). */
export interface SearchFilters {
  year?: number | null
  genre?: number | null
}

const searchFilterParams = ({ year, genre }: SearchFilters) => ({ year: year || undefined, genre: genre || undefined })

export const movieApi = {
  // Movies
  getPopular: (page = 1, with_genres?: string | number) =>
    api.get<PaginatedResult<Movie>>("/movies/popular", {
      params: { page, with_genres },
    }),

  getNowPlaying: (page = 1, with_genres?: string | number) =>
    api.get<PaginatedResult<Movie>>("/movies/now-playing", {
      params: { page, with_genres },
    }),

  getTopRated: (page = 1, with_genres?: string | number) =>
    api.get<PaginatedResult<Movie>>("/movies/top-rated", {
      params: { page, with_genres },
    }),

  getUpcoming: (page = 1, with_genres?: string | number) =>
    api.get<PaginatedResult<Movie>>("/movies/upcoming", {
      params: { page, with_genres },
    }),

  search: (q: string, page = 1, filters: SearchFilters = {}) =>
    api.get<PaginatedResult<Movie>>("/movies/search", { params: { q, page, ...searchFilterParams(filters) } }),

  getGenres: () => api.get<{ genres: Genre[] }>("/movies/genres"),

  getById: (id: number) =>
    api.get<{ movie: MovieDetail; credits: Credits; videos: Video[] }>(
      `/movies/${id}`,
    ),

  getVideos: (id: number) =>
    api
      .get<{
        movie: MovieDetail
        credits: Credits
        videos: Video[]
      }>(`/movies/${id}`)
      .then(res => ({ data: { results: res.data.videos } })),

  getWatchProviders: (type: "movie" | "tv", id: number, region: string) =>
    api.get<{ providers: WatchProviders }>(`/${type === "tv" ? "tv" : "movies"}/${id}/watch-providers`, {
      params: { region },
    }),

  getSimilar: (id: number, page = 1) =>
    api.get<PaginatedResult<Movie>>(`/movies/${id}/similar`, {
      params: { page },
    }),
  getRecommended: (genres: string) =>
    api.get<PaginatedResult<Movie>>("/movies/recommended", {
      params: {
        with_genres: genres,
      },
    }),

  // TV Series
  getPopularSeries: (page = 1, with_genres?: string | number) =>
    api.get<PaginatedResult<TVSeries>>("/tv/popular", {
      params: { page, with_genres },
    }),

  getNowAiringSeries: (page = 1, with_genres?: string | number) =>
    api.get<PaginatedResult<TVSeries>>("/tv/now-airing", {
      params: { page, with_genres },
    }),

  getTopRatedSeries: (page = 1, with_genres?: string | number) =>
    api.get<PaginatedResult<TVSeries>>("/tv/top-rated", {
      params: { page, with_genres },
    }),

  searchSeries: (q: string, page = 1, filters: SearchFilters = {}) =>
    api.get<PaginatedResult<TVSeries>>("/tv/search", { params: { q, page, ...searchFilterParams(filters) } }),

  getSeriesGenres: () => api.get<{ genres: Genre[] }>("/tv/genres"),

  getSeriesById: (id: number) =>
    api.get<{ series: TVSeriesDetail; credits: Credits; videos: Video[] }>(
      `/tv/${id}`,
    ),

  getSimilarSeries: (id: number, page = 1) =>
    api.get<PaginatedResult<TVSeries>>(`/tv/${id}/similar`, {
      params: { page },
    }),

  getTVVideos: (id: number) =>
    api
      .get<{
        tv: TVSeriesDetail
        credits: Credits
        videos: Video[]
      }>(`/tv/${id}`)
      .then(res => ({ data: { results: res.data.videos } })),
}

export const actorApi = {
  search: (q: string, page = 1) =>
    api.get<PaginatedResult<PersonSearchResult>>("/actors/search", { params: { q, page } }),
  getById: (id: number) => api.get<PersonDetail>(`/actors/${id}`),
  getMovies: (id: number) => api.get<PersonMovieCredits>(`/actors/${id}/movies`),
  getTVSeries: (id: number) => api.get<PersonTVCredits>(`/actors/${id}/tv`),
}
