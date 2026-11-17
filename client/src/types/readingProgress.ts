/** The signed-in viewer's position in one manga. Chapter and volume are whole numbers. */
export interface ReadingProgress {
  media_id: number
  media_type: "manga"
  chapter: number
  volume: number
  updated_at: string
}

/** Chapter and volume as the viewer sets them; the server refuses both at 0. */
export interface ReadingProgressInput {
  chapter: number
  volume: number
}

/** GET /api/reading-progress/manga/:id. `progress` is null until something is saved. */
export interface ReadingProgressResponse {
  progress: ReadingProgress | null
}

/** PUT /api/reading-progress/manga/:id. */
export interface SavedReadingProgressResponse {
  progress: ReadingProgress
}
