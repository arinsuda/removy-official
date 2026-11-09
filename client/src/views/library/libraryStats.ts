import { dedupeWatched } from "@/components/profile/tabs/watched"
import { summaryToHomeMedia, type ProfileEntry } from "@/components/profile/tabs/profileMedia"
import type { LibraryItemResponse, MediaSummary, ReviewResponse } from "@/types"
import { RATING_MAX, RATING_STEP } from "./config"

/** A saved title with its genres kept for the dashboard breakdown. */
export interface LibraryEntry extends ProfileEntry {
  summary: Pick<MediaSummary, "genres">
}

export function toEntries(items: LibraryItemResponse[], dateOf: (i: LibraryItemResponse) => string = (i) => i.created_at): LibraryEntry[] {
  return items.filter((i) => i.media).map((i) => ({ id: i.id, summary: i.media, media: summaryToHomeMedia(i.media), at: dateOf(i) }))
}

export const toWatchedEntries = (items: LibraryItemResponse[]): LibraryEntry[] => dedupeWatched(toEntries(items, (i) => i.watched_at ?? i.created_at))

const time = (iso: string) => new Date(iso).getTime() || 0

export const newestFirst = <T extends { at: string }>(entries: T[]): T[] => [...entries].sort((a, b) => time(b.at) - time(a.at))

/** ISO timestamp of the most recent entry, or null for an empty list. */
export function latestAt(entries: { at: string }[]): string | null {
  return entries.length ? newestFirst(entries)[0]!.at : null
}

/** Poster paths for a list cover, newest first, skipping titles without one. */
export function collagePosters(entries: { media: { poster_path: string } }[], size: number): string[] {
  return entries.map((e) => e.media.poster_path).filter(Boolean).slice(0, size)
}

export function averageRating(reviews: ReviewResponse[]): number {
  if (!reviews.length) return 0
  return reviews.reduce((sum, r) => sum + r.rating, 0) / reviews.length
}

export function topGenres(entries: LibraryEntry[], limit: number): { name: string; count: number }[] {
  const counts = new Map<string, number>()
  for (const e of entries) for (const g of e.summary.genres ?? []) counts.set(g.name, (counts.get(g.name) ?? 0) + 1)
  return [...counts].map(([name, count]) => ({ name, count })).sort((a, b) => b.count - a.count || a.name.localeCompare(b.name)).slice(0, limit)
}

/** Reviews written this calendar month minus last month. */
export function reviewMonthDiff(reviews: ReviewResponse[], now = new Date()): number {
  const index = (d: Date) => d.getFullYear() * 12 + d.getMonth()
  const current = index(now)
  let diff = 0
  for (const r of reviews) {
    const gap = current - index(new Date(r.created_at))
    if (gap === 0) diff++
    else if (gap === 1) diff--
  }
  return diff
}

export interface RatingBucket {
  rating: number
  count: number
  percent: number
}

/** One bucket per half star, highest rating first. */
export function ratingBuckets(reviews: ReviewResponse[]): RatingBucket[] {
  const size = RATING_MAX / RATING_STEP
  const counts = Array<number>(size).fill(0)
  for (const r of reviews) {
    const i = Math.round(r.rating / RATING_STEP) - 1
    if (i >= 0 && i < size) counts[i]!++
  }
  const max = Math.max(...counts, 1)
  return counts.map((count, i) => ({ rating: (i + 1) * RATING_STEP, count, percent: (count / max) * 100 })).reverse()
}
