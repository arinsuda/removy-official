import type { AxiosResponse } from "axios"
import api from "../index"
import { MANGA_PROGRESS_PATH } from "@/components/manga/mangaConfig"
import type { ReadingProgressInput, ReadingProgressResponse, SavedReadingProgressResponse } from "@/types/readingProgress"

const mangaProgressUrl = (mediaId: number) => `${MANGA_PROGRESS_PATH}/${mediaId}`

function get(mediaId: number): Promise<AxiosResponse<ReadingProgressResponse>> {
  return api.get<ReadingProgressResponse>(mangaProgressUrl(mediaId))
}

function put(mediaId: number, input: ReadingProgressInput): Promise<AxiosResponse<SavedReadingProgressResponse>> {
  return api.put<SavedReadingProgressResponse>(mangaProgressUrl(mediaId), input)
}

function remove(mediaId: number): Promise<AxiosResponse<void>> {
  return api.delete<void>(mangaProgressUrl(mediaId))
}

export const readingProgressApi = { get, put, remove }
