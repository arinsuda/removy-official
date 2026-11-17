import { MANGA_PROGRESS_MAX_CHAPTER, MANGA_PROGRESS_MAX_VOLUME } from "./mangaConfig"
import { knownNumber } from "./mangaFormat"
import type { ReadingProgress, ReadingProgressInput } from "@/types/readingProgress"

export type ProgressField = keyof ReadingProgressInput

/** Counts AniList knows for a title; null when the title does not say. */
export interface ProgressTotals {
  chapters: number | null
  volumes: number | null
}

const FIELD_CAP: Readonly<Record<ProgressField, number>> = {
  chapter: MANGA_PROGRESS_MAX_CHAPTER,
  volume: MANGA_PROGRESS_MAX_VOLUME,
}

const WHOLE_NUMBER = /^\d+$/

/** Server error codes of the progress endpoints -> i18n keys. Unknown codes fall back to the generic message. */
const PROGRESS_ERROR_KEYS: Readonly<Record<string, string>> = {
  PROGRESS_OUT_OF_RANGE: "manga.progress.errors.outOfRange",
  INVALID_INPUT: "manga.progress.errors.invalid",
  AGE_RESTRICTED: "manga.progress.errors.ageRestricted",
  UPSTREAM_UNAVAILABLE: "manga.progress.errors.upstream",
}
export const PROGRESS_SAVE_FAILED_KEY = "manga.progress.errors.generic"

/** Highest value a field may take: AniList's total when known, never above the server cap. */
export function progressMax(field: ProgressField, totals: ProgressTotals): number {
  const total = knownNumber(field === "chapter" ? totals.chapters : totals.volumes)
  return total === null ? FIELD_CAP[field] : Math.min(total, FIELD_CAP[field])
}

/** Highest value the server accepts for a field; typed values may exceed AniList's total up to this. */
export function progressCap(field: ProgressField): number {
  return FIELD_CAP[field]
}

export function isEmptyProgress(values: ReadingProgressInput): boolean {
  return values.chapter === 0 && values.volume === 0
}

/**
 * The pair after one stepper click, or null when the click changes nothing. The server refuses
 * (0, 0), so a step that would land there is dropped as well.
 */
export function stepProgress(
  values: ReadingProgressInput,
  field: ProgressField,
  delta: number,
  totals: ProgressTotals,
): ReadingProgressInput | null {
  const next = Math.min(Math.max(values[field] + delta, 0), progressMax(field, totals))
  // A stored value above AniList's total must not be pulled down by an increase.
  if (next === values[field] || (delta > 0 && next < values[field])) return null
  const stepped = { ...values, [field]: next }
  return isEmptyProgress(stepped) ? null : stepped
}

/** The saved position as steppers see it; no progress yet reads as 0 and 0. */
export function progressValues(progress: ReadingProgress | null): ReadingProgressInput {
  return { chapter: progress?.chapter ?? 0, volume: progress?.volume ?? 0 }
}

/**
 * A typed value: a whole number from 0 to `max`, or null when it is not one. A type=number input
 * hands its value over as a number through v-model, so the text is coerced before it is checked.
 */
export function parseProgressField(text: string, max: number): number | null {
  const trimmed = String(text).trim()
  if (!WHOLE_NUMBER.test(trimmed)) return null
  const value = Number(trimmed)
  return value <= max ? value : null
}

/** The i18n key for a failed progress request, read from the `code` the server sends. */
export function progressErrorKey(err: unknown): string {
  const code = (err as { response?: { data?: { code?: string } } })?.response?.data?.code
  return (code && PROGRESS_ERROR_KEYS[code]) || PROGRESS_SAVE_FAILED_KEY
}
