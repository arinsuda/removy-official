import type { BmolCandidate } from "@/components/bmol/useBmol"
import { anilistScore, type MangaItem } from "@/types/manga"

/** AniList scores run 0-100; the candidate list shows the 0-5 scale used elsewhere. */
export const ANILIST_TO_FIVE = 20

/** A manga as a Best of Life candidate (AniList cover and score, 0-5 scale). */
export function mangaToBmolCandidate(m: MangaItem): BmolCandidate {
  return {
    id: m.id,
    media_type: "manga",
    title: m.display_title,
    poster_url: m.cover_url,
    vote_average: (anilistScore(m.ratings) ?? 0) / ANILIST_TO_FIVE,
  }
}
