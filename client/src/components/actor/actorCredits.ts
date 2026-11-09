import { fromMovie, fromSeries, type HomeMedia } from "@/components/home/homeMedia"
import type { ActorMovieCredit, ActorTVCredit, Movie, TVSeries } from "@/types"
import type { CreditFilter, CreditSort } from "./actorConfig"

export interface ActorCredit {
  key: string
  media: HomeMedia
  /** Release or first-air year; null when the date is unknown. */
  year: number | null
  /** Character name(s), only for acting credits. */
  character: string
  popularity: number
}

export const yearOf = (date: string | null | undefined): number | null => {
  const y = Number.parseInt((date ?? "").slice(0, 4), 10)
  return Number.isFinite(y) ? y : null
}

function build(media: HomeMedia, character: string | undefined, popularity: number): ActorCredit {
  return { key: media.key, media, year: yearOf(media.date), character: character?.trim() ?? "", popularity }
}

/** Joins the characters of a person who played several roles in one title. */
function mergeCharacter(existing: ActorCredit, character: string | undefined) {
  const next = character?.trim()
  if (next && existing.character && !existing.character.includes(next)) existing.character = `${existing.character} / ${next}`
  else if (next && !existing.character) existing.character = next
}

/** One credit per title; TMDB lists a title once per role. */
export function dedupeCredits(credits: ActorCredit[]): ActorCredit[] {
  const map = new Map<string, ActorCredit>()
  for (const credit of credits) {
    const existing = map.get(credit.key)
    if (existing) mergeCharacter(existing, credit.character)
    else map.set(credit.key, { ...credit })
  }
  return [...map.values()]
}

export function actingCredits(movies: ActorMovieCredit[], series: ActorTVCredit[]): ActorCredit[] {
  return dedupeCredits([
    ...movies.map((m) => build(fromMovie(m), m.character, m.popularity)),
    ...series.map((s) => build(fromSeries(s), s.character, s.popularity)),
  ])
}

export function crewCredits(movies: Movie[], series: TVSeries[]): ActorCredit[] {
  return dedupeCredits([
    ...movies.map((m) => build(fromMovie(m), undefined, m.popularity)),
    ...series.map((s) => build(fromSeries(s), undefined, s.popularity)),
  ])
}

export function filterCredits(credits: ActorCredit[], filter: CreditFilter): ActorCredit[] {
  return filter === "all" ? credits : credits.filter((c) => c.media.type === filter)
}

/** Credits without a date sort last in both date orders. */
export function sortCredits(credits: ActorCredit[], sort: CreditSort): ActorCredit[] {
  const list = [...credits]
  const byPopularity = (a: ActorCredit, b: ActorCredit) => b.popularity - a.popularity
  if (sort === "popular") return list.sort(byPopularity)
  const dir = sort === "newest" ? -1 : 1
  return list.sort((a, b) => {
    if (a.year === null || b.year === null) return a.year === b.year ? byPopularity(a, b) : a.year === null ? 1 : -1
    return a.year === b.year ? byPopularity(a, b) : (a.year - b.year) * dir
  })
}

export interface YearGroup {
  /** null for credits without a date. */
  year: number | null
  credits: ActorCredit[]
}

/** Groups credits by year, keeping the order of the (already sorted) input; unknown year goes last. */
export function groupByYear(credits: ActorCredit[], ascending: boolean): YearGroup[] {
  const map = new Map<number | null, ActorCredit[]>()
  for (const credit of credits) map.set(credit.year, [...(map.get(credit.year) ?? []), credit])
  const years = [...map.keys()].filter((y): y is number => y !== null).sort((a, b) => (ascending ? a - b : b - a))
  const ordered: (number | null)[] = map.has(null) ? [...years, null] : years
  return ordered.map((year) => ({ year, credits: map.get(year) ?? [] }))
}

export interface CreditStats {
  total: number
  movies: number
  tv: number
  firstYear: number | null
  lastYear: number | null
}

export function creditStats(credits: ActorCredit[]): CreditStats {
  const years = credits.map((c) => c.year).filter((y): y is number => y !== null)
  return {
    total: credits.length,
    movies: credits.filter((c) => c.media.type === "movie").length,
    tv: credits.filter((c) => c.media.type === "tv").length,
    firstYear: years.length ? Math.min(...years) : null,
    lastYear: years.length ? Math.max(...years) : null,
  }
}

/** Most popular titles; those without a poster are skipped because the rail is visual. */
export function knownFor(credits: ActorCredit[], limit: number): ActorCredit[] {
  return sortCredits(
    credits.filter((c) => c.media.poster_path),
    "popular",
  ).slice(0, limit)
}

export interface AgeInfo {
  age: number
  deceased: boolean
}

/** Whole years between birth and death (or `now`), or null when the birthday is missing or invalid. */
export function ageOf(birthday: string | null, deathday: string | null, now = new Date()): AgeInfo | null {
  const born = parseDay(birthday)
  if (!born) return null
  const died = parseDay(deathday)
  const end = died ?? now
  let age = end.getFullYear() - born.getFullYear()
  const beforeBirthday = end.getMonth() < born.getMonth() || (end.getMonth() === born.getMonth() && end.getDate() < born.getDate())
  if (beforeBirthday) age--
  return { age: Math.max(0, age), deceased: !!died }
}

/** Parses TMDB "YYYY-MM-DD" as a local date so no timezone shifts the day. */
export function parseDay(value: string | null | undefined): Date | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(value ?? "")
  if (!m) return null
  return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
}
