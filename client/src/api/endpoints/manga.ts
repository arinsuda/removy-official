import type { AxiosResponse } from "axios"
import api from "../index"
import type { MangaDetail, MangaItem, MangaQuery, PaginatedResult } from "@/types"

function list(query: MangaQuery, page = 1): Promise<AxiosResponse<PaginatedResult<MangaItem>>> {
  return api.get<PaginatedResult<MangaItem>>("/manga", { params: { ...query, page } })
}

function search(q: string, page = 1, includeAdult = false): Promise<AxiosResponse<PaginatedResult<MangaItem>>> {
  return api.get<PaginatedResult<MangaItem>>("/manga/search", { params: { q, page, include_adult: includeAdult } })
}

function genres(): Promise<AxiosResponse<{ genres: string[] }>> {
  return api.get<{ genres: string[] }>("/manga/genres")
}

function detail(id: number): Promise<AxiosResponse<{ manga: MangaDetail }>> {
  return api.get<{ manga: MangaDetail }>(`/manga/${id}`)
}

function similar(id: number, includeAdult = false): Promise<AxiosResponse<{ results: MangaItem[] }>> {
  return api.get<{ results: MangaItem[] }>(`/manga/${id}/similar`, { params: { include_adult: includeAdult } })
}

export const mangaApi = { list, search, genres, detail, similar }
