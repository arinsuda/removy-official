import type { Movie } from "@/types"

export const MIN_PICKS = 5
export const TOTAL_MOVIES = 24
/** Time the shell plays its exit transition before navigating away. */
export const EXIT_MS = 320

/** TMDB genre ids that have an `onboarding.genres.<id>` label. */
export const KNOWN_GENRE_IDS = [
  28, 12, 16, 35, 80, 99, 18, 10751, 14, 36, 27, 10402, 9648, 10749, 878, 10770, 53, 10752, 37,
]

// Genres sampled first so the pool is varied, two movies each.
const TARGET_GENRES = [28, 35, 18, 878, 27, 14, 12, 80, 10749, 53]

const score = (m: Movie) => m.vote_average * Math.log(m.vote_count + 1)

export function genreLabelKey(id: number | undefined): string | null {
  return id !== undefined && KNOWN_GENRE_IDS.includes(id) ? `onboarding.genres.${id}` : null
}

/** Dedupe, then take the best two titles per target genre and fill up by score. */
export function pickDiverse(source: Movie[], total = TOTAL_MOVIES, random = Math.random): Movie[] {
  const seen = new Set<number>()
  const all: Movie[] = []
  for (const m of source) {
    if (!seen.has(m.id) && m.poster_path) {
      seen.add(m.id)
      all.push(m)
    }
  }

  const picked = new Set<number>()
  const result: Movie[] = []
  for (const genreId of TARGET_GENRES) {
    const best = all
      .filter((m) => !picked.has(m.id) && m.genre_ids?.includes(genreId))
      .sort((a, b) => score(b) - score(a))
      .slice(0, 2)
    for (const m of best) {
      picked.add(m.id)
      result.push(m)
    }
  }

  const rest = all.filter((m) => !picked.has(m.id)).sort((a, b) => score(b) - score(a))
  for (const m of rest) {
    if (result.length >= total) break
    result.push(m)
  }

  // Shuffle so the grid does not read as sorted by genre.
  for (let i = result.length - 1; i > 0; i--) {
    const j = Math.floor(random() * (i + 1))
    ;[result[i], result[j]] = [result[j]!, result[i]!]
  }
  return result.slice(0, total)
}

/** Genre ids ranked by how many selected movies carry them. */
export function rankGenres(movies: Movie[], limit?: number): number[] {
  const freq = new Map<number, number>()
  for (const m of movies) {
    for (const g of m.genre_ids ?? []) freq.set(g, (freq.get(g) ?? 0) + 1)
  }
  const ranked = [...freq.entries()].sort((a, b) => b[1] - a[1]).map(([id]) => id)
  return limit ? ranked.slice(0, limit) : ranked
}

/** The favorite-genres payload: the five most common genres among the picks. */
export const extractGenres = (movies: Movie[]) => rankGenres(movies, 5)
