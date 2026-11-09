import { popcornTier, roundScore, type PopcornTier } from "@/components/movie/popcornTier"
import type {
  ShareMediaContext,
  ShareReviewContext,
  ShareStatsContext,
  ShareTemplateKey,
  ShareTheme,
} from "@/types/share"

export interface ShareTemplateDef {
  key: ShareTemplateKey
  width: number
  height: number
  /** Exported without a background so it can sit on top of a story. */
  transparent: boolean
  labelKey: string
  /** Format hint shown under the preview, e.g. "9:16". */
  ratio: string
}

export const SHARE_THEMES: ShareTheme[] = ["dark", "light", "brand"]

const DEFS: Record<ShareTemplateKey, ShareTemplateDef> = {
  poster: { key: "poster", width: 1080, height: 1920, transparent: false, labelKey: "share.templates.poster", ratio: "9:16" },
  take: { key: "take", width: 1080, height: 1920, transparent: false, labelKey: "share.templates.take", ratio: "9:16" },
  sticker: { key: "sticker", width: 1080, height: 1080, transparent: true, labelKey: "share.templates.sticker", ratio: "PNG" },
  square: { key: "square", width: 1080, height: 1080, transparent: false, labelKey: "share.templates.square", ratio: "1:1" },
  review: { key: "review", width: 1080, height: 1920, transparent: false, labelKey: "share.templates.review", ratio: "9:16" },
}

export function templateDef(key: ShareTemplateKey): ShareTemplateDef {
  return DEFS[key]
}

/** Review text may only appear on a card when the review is public. */
export function publicExcerpt(visibility: string | null | undefined, body: string | null | undefined): string | null {
  if (visibility !== "public") return null
  const text = body?.trim()
  return text ? text : null
}

export function hasOwnStats(stats: ShareStatsContext | null | undefined): stats is ShareStatsContext {
  return Boolean(stats && ((stats.rating ?? 0) > 0 || stats.watchCount > 0))
}

/**
 * Which templates can be offered. "My take" needs the owner's own stats; the
 * review card needs a public review. Stats are ignored when sharing someone's review.
 */
export function availableTemplates(
  shareType: "poster" | "review",
  stats: ShareStatsContext | null | undefined,
  review: ShareReviewContext | null | undefined,
): ShareTemplateDef[] {
  if (shareType === "review") {
    const keys: ShareTemplateKey[] = publicExcerpt(review?.visibility, review?.body) ? ["review", "poster", "square"] : ["poster", "square"]
    return keys.map(templateDef)
  }
  const keys: ShareTemplateKey[] = ["poster"]
  if (hasOwnStats(stats)) keys.push("take")
  keys.push("sticker", "square")
  return keys.map(templateDef)
}

export function shareTier(media: ShareMediaContext): PopcornTier {
  // Without a review count, a positive POPCORN score still counts as reviewed.
  const count = media.popcornCount ?? (media.popcornRating ? 1 : 0)
  return popcornTier(media.popcornRating, count)
}

export interface ShareScore {
  value: string
  scale: string
  source: "POPCORN" | "TMDB"
}

export function shareScore(media: ShareMediaContext): ShareScore | null {
  if (media.popcornRating && media.popcornRating > 0) {
    return { value: roundScore(media.popcornRating).toFixed(1), scale: "/5", source: "POPCORN" }
  }
  if (media.voteAverage > 0) return { value: media.voteAverage.toFixed(1), scale: "/10", source: "TMDB" }
  return null
}

const GENRES_EN: Record<number, string> = {
  28: "Action", 12: "Adventure", 16: "Animation", 35: "Comedy", 80: "Crime", 99: "Documentary", 18: "Drama",
  10751: "Family", 14: "Fantasy", 36: "History", 27: "Horror", 10402: "Music", 9648: "Mystery", 10749: "Romance",
  878: "Sci-Fi", 10770: "TV Movie", 53: "Thriller", 10752: "War", 37: "Western", 10759: "Action & Adventure",
  10762: "Kids", 10763: "News", 10764: "Reality", 10765: "Sci-Fi & Fantasy", 10766: "Soap", 10767: "Talk",
  10768: "War & Politics",
}

export function shareGenres(media: ShareMediaContext, max = 3): string[] {
  return media.genres.slice(0, max).map((g) => GENRES_EN[g.id] || g.name)
}

/** "example.com/movies/1" without protocol, for the card footer. */
export function shortShareUrl(media: ShareMediaContext): string {
  const path = media.mediaType === "movie" ? "movies" : "tv"
  return `${window.location.host}/${path}/${media.id}`
}

/** Title font size steps so long titles never overflow the card. */
export function titleSizeClass(title: string): string {
  const len = Array.from(title).length
  if (len > 50) return "sc-title--xs"
  if (len > 30) return "sc-title--sm"
  if (len > 18) return "sc-title--md"
  return "sc-title--lg"
}

/**
 * Cut text to at most `max` characters on a word boundary and append "…".
 * Text without spaces (e.g. Thai) is cut at `max`. Text that fits is returned as is.
 */
export function truncateExcerpt(text: string, max: number): string {
  const chars = Array.from(text)
  if (chars.length <= max) return text
  const head = chars.slice(0, max).join("")
  const space = head.lastIndexOf(" ")
  const cut = space >= max * 0.6 ? head.slice(0, space) : head
  return `${cut.trimEnd().replace(/[,;:.\-–—]+$/, "")}…`
}

// --- Best of Life (Hall of Fame) cards ---

export type BolTemplateKey = "story" | "square"

export interface BolTemplateDef {
  key: BolTemplateKey
  width: number
  height: number
  labelKey: string
  ratio: string
  /** How many titles the card shows, in rank order. */
  capacity: number
}

const BOL_DEFS: Record<BolTemplateKey, BolTemplateDef> = {
  story: { key: "story", width: 1080, height: 1920, labelKey: "library.bmol.share.templates.story", ratio: "9:16", capacity: 9 },
  square: { key: "square", width: 1080, height: 1080, labelKey: "library.bmol.share.templates.square", ratio: "1:1", capacity: 3 },
}

export const BOL_TEMPLATES: BolTemplateDef[] = [BOL_DEFS.story, BOL_DEFS.square]

export function bolTemplateDef(key: BolTemplateKey): BolTemplateDef {
  return BOL_DEFS[key]
}

/** Stable key for a poster lookup. */
export function bolPosterKey(item: { media_type: string; media: { id: number } }): string {
  return `${item.media_type}:${item.media.id}`
}

/** The first `max` titles in rank order; ties keep their insertion order. */
export function bolTopItems<T extends { rank: number }>(groups: { items: T[] }[], max: number): T[] {
  return groups.flatMap((g) => g.items).slice(0, max)
}

/**
 * Note on the card: only when the owner opted in, and only the note of the
 * first rank-1 title. Everything else on the card is already public.
 */
export function bolShareNote(items: { rank: number; note: string }[], include: boolean, max = 140): string | null {
  if (!include) return null
  const top = items.find((i) => i.rank === 1)
  const text = top?.note?.trim()
  return text ? truncateExcerpt(text, max) : null
}
