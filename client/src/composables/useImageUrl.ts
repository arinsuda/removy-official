import { mediaImage, type TmdbImageSize } from '@/utils/image'

export function useImageUrl() {
  /** TMDB paths get the TMDB base; absolute URLs (manga covers) pass through. */
  function tmdbImage(path: string | null | undefined, size: TmdbImageSize = 'w500'): string {
    return mediaImage(path, size) ?? '/placeholder-poster.png'
  }

  function backdropImage(path: string | null | undefined): string {
    return tmdbImage(path, 'w780')
  }

  function posterImage(path: string | null | undefined): string {
    return tmdbImage(path, 'w342')
  }

  return { tmdbImage, backdropImage, posterImage }
}
