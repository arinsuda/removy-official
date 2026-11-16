import type { AxiosResponse } from "axios"
import api from "../index"
import type { AnimeMediaType, AnimeQuery, Movie, PaginatedResult, TVSeries } from "@/types"

/** Same response shape as /tv/popular and /movies/popular. */
const ANIME_PATHS: Record<AnimeMediaType, string> = {
  tv: "/anime/tv",
  movie: "/anime/movies",
}

function list(type: "tv", query: AnimeQuery, page?: number): Promise<AxiosResponse<PaginatedResult<TVSeries>>>
function list(type: "movie", query: AnimeQuery, page?: number): Promise<AxiosResponse<PaginatedResult<Movie>>>
function list(type: AnimeMediaType, query: AnimeQuery, page = 1) {
  return api.get<PaginatedResult<Movie | TVSeries>>(ANIME_PATHS[type], { params: { ...query, page } })
}

export const animeApi = { list }
