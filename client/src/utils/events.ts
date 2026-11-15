import type {
  AdminEvent,
  AdminEventBody,
  AppEvent,
  AwardMediaType,
  EventType,
  LocalizedText,
  RoundKind,
} from "@/types/event"
import { getTmdbImageUrl, type TmdbImageSize } from "@/utils/image"

/** localStorage key prefix for "don't show again today"; the user id is appended. */
export const EVENT_POPUP_HIDE_KEY_PREFIX = "popcorn_event_popup_hidden_"
/** sessionStorage key prefix for "closed in this session"; the user id is appended. */
export const EVENT_POPUP_SESSION_KEY_PREFIX = "popcorn_event_popup_closed_"

export const EVENT_TITLE_MAX = 120
export const EVENT_BODY_MAX = 2000
export const EVENT_IMAGE_MAX_BYTES = 5 * 1024 * 1024
export const EVENT_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"] as const
export const EVENTS_PAGE_SIZE = 12
export const ADMIN_EVENTS_PAGE_SIZE = 20
export const HTTP_NOT_FOUND = 404
/** Byte cap on link_url, mirrors server maxLinkURLLength. */
export const EVENT_LINK_MAX = 500
/** Whitespace and backslashes are rejected in link_url, mirrors server validLinkURL. */
export const EVENT_LINK_FORBIDDEN_RE = /[\s\\]/
/** Monthly round period (YYYY-MM), mirrors server monthlyPeriodRe. */
export const MONTHLY_PERIOD_RE = /^\d{4}-(0[1-9]|1[0-2])$/
/** Yearly round period (YYYY), mirrors server yearlyPeriodRe. */
export const YEARLY_PERIOD_RE = /^\d{4}$/
/** Fewest nominees a voting round may be saved or published with; mirrors server minNomineesToPublish. */
export const EVENT_MIN_NOMINEES_TO_PUBLISH = 2
/** Account age (days) before a user may vote; mirrors server MinVoterAccountAge. */
export const VOTER_MIN_ACCOUNT_AGE_DAYS = 7
export const EVENT_POPUP_PRIORITY_MIN = -100
export const EVENT_POPUP_PRIORITY_MAX = 100
/** TMDB poster width for nominee cards on the event detail page. */
export const NOMINEE_POSTER_SIZE: TmdbImageSize = "w342"
/** TMDB backdrop width for an event cover when the event has no uploaded image. */
export const EVENT_COVER_SIZE: TmdbImageSize = "w780"
/** TMDB poster width for winner cards on the awards page and the award badges. */
export const AWARD_POSTER_SIZE: TmdbImageSize = "w342"
const MINUTE_MS = 60 * 1000
const HOUR_MS = 60 * MINUTE_MS
const DAY_MS = 24 * HOUR_MS
/** Late start opens this many minutes after now; mirrors server lateStartLead. */
export const LATE_START_LEAD_MINUTES = 60
/** A late start is offered only while the official window keeps this many hours; mirrors server lateStartMinWindow. */
export const LATE_START_MIN_WINDOW_HOURS = 72
export const LATE_START_LEAD_MS = LATE_START_LEAD_MINUTES * MINUTE_MS
export const LATE_START_MIN_WINDOW_MS = LATE_START_MIN_WINDOW_HOURS * HOUR_MS
/** Thailand has no daylight saving time, so Asia/Bangkok is a fixed UTC+7 offset. */
const VOTING_UTC_OFFSET_MS = 7 * HOUR_MS
/** Default window for a new event in the admin form: starts now, lasts two weeks. */
export const EVENT_DEFAULT_DURATION_MS = 14 * DAY_MS
/** Day of month (Thailand time) on which monthly voting rounds roll over to the next month. */
export const VOTING_CUTOFF_DAY = 15
export const VOTING_TIMEZONE = "Asia/Bangkok"
const MONTHS_PER_YEAR = 12
/** A yearly round for year Y opens on 16 January of Y+1 and closes on 15 February (Thailand time). */
const YEARLY_WINDOW_START_MONTH = 1
const YEARLY_WINDOW_START_DAY = 16
/** The yearly window closes at 00:00 on 15 February (exclusive), Thailand time. */
const YEARLY_WINDOW_END_MONTH = 2
const YEARLY_WINDOW_END_DAY = 15
const PAD_WIDTH = 2

/** Monthly period "YYYY-MM" for a calendar year and month. */
function formatMonthPeriod(year: number, month: number): string {
  return `${year}-${String(month).padStart(PAD_WIDTH, "0")}`
}

/** Instant of 00:00 on `day` of the given month in Thailand time. */
function bangkokMidnight(year: number, month: number, day: number): Date {
  return new Date(Date.UTC(year, month - 1, day) - VOTING_UTC_OFFSET_MS)
}

/** A voting round's period and the instant its official window closes. */
interface CurrentRound {
  period: string
  endsAt: Date
}

/** Shift a calendar month by `delta` months, crossing year boundaries. */
function shiftMonth(year: number, month: number, delta: number): { year: number; month: number } {
  const index = year * MONTHS_PER_YEAR + (month - 1) + delta
  return { year: Math.floor(index / MONTHS_PER_YEAR), month: (index % MONTHS_PER_YEAR) + 1 }
}

/** Calendar year, month and day of `now` in the voting timezone (Thailand time). */
function votingDateParts(now: Date): { year: number; month: number; day: number } {
  const parts = new Intl.DateTimeFormat("en-CA", { // hardcode-ok: machine YYYY-MM-DD parts
    timeZone: VOTING_TIMEZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(now)
  return {
    year: Number(parts.find((p) => p.type === "year")?.value ?? now.getUTCFullYear()),
    month: Number(parts.find((p) => p.type === "month")?.value ?? now.getUTCMonth() + 1),
    day: Number(parts.find((p) => p.type === "day")?.value ?? now.getUTCDate()),
  }
}

/**
 * Period (YYYY-MM) of the monthly round to create next, in Thailand time. A round is named by the
 * month its voting starts (the 15th) and runs until the 15th of the following month.
 */
export function nextMonthlyPeriod(now: Date = new Date()): string {
  const { year, month, day } = votingDateParts(now)
  // Before the cutoff, the active round started last month, so the next one starts this month.
  // From the cutoff on, the active round started this month, so the next one starts next month.
  const monthsAhead = day >= VOTING_CUTOFF_DAY ? 1 : 0
  const next = shiftMonth(year, month, monthsAhead)
  return formatMonthPeriod(next.year, next.month)
}

/**
 * Period (YYYY) of the yearly round to create next, in Thailand time. Until the 16th of January the
 * round for the previous year is still to open; from that day on, the next round is this year's.
 */
export function nextYearlyPeriod(now: Date = new Date()): string {
  const { year, month, day } = votingDateParts(now)
  const windowStarted =
    month > YEARLY_WINDOW_START_MONTH || (month === YEARLY_WINDOW_START_MONTH && day >= YEARLY_WINDOW_START_DAY)
  return String(windowStarted ? year : year - 1)
}

/** Monthly round whose window contains `now`: round P is open from the 15th of P until the 15th of P+1. */
function currentMonthlyRound(now: Date): CurrentRound {
  const { year, month, day } = votingDateParts(now)
  const start = shiftMonth(year, month, day >= VOTING_CUTOFF_DAY ? 0 : -1)
  const end = shiftMonth(start.year, start.month, 1)
  return {
    period: formatMonthPeriod(start.year, start.month),
    endsAt: bangkokMidnight(end.year, end.month, VOTING_CUTOFF_DAY),
  }
}

/** Yearly round whose window (16 January to 15 February, Thailand time) contains `now`, or null. */
function currentYearlyRound(now: Date): CurrentRound | null {
  const { year, month, day } = votingDateParts(now)
  const inWindow =
    (month === YEARLY_WINDOW_START_MONTH && day >= YEARLY_WINDOW_START_DAY) ||
    (month === YEARLY_WINDOW_END_MONTH && day < YEARLY_WINDOW_END_DAY)
  if (!inWindow) return null
  // The window opens in January of Y+1 for round Y, so the round is always the previous year's.
  return {
    period: String(year - 1),
    endsAt: bangkokMidnight(year, YEARLY_WINDOW_END_MONTH, YEARLY_WINDOW_END_DAY),
  }
}

/** Period (YYYY-MM) of the monthly round open at `now`; monthly windows are back to back, so always set. */
export function currentMonthlyPeriod(now: Date = new Date()): string {
  return currentMonthlyRound(now).period
}

/** Period (YYYY) of the yearly round open at `now`, or null outside its window. */
export function currentYearlyPeriod(now: Date = new Date()): string | null {
  return currentYearlyRound(now)?.period ?? null
}

/** Period of the round of `kind` whose official window contains `now`, or null when none is open. */
export function currentRoundPeriod(kind: RoundKind, now: Date = new Date()): string | null {
  return kind === "monthly" ? currentMonthlyPeriod(now) : currentYearlyPeriod(now)
}

/**
 * True when the current round of `kind` can start late. A late start opens LATE_START_LEAD_MS from now
 * and keeps the official end, so at least LATE_START_MIN_WINDOW_MS must remain. Mirrors server rules.
 */
export function canStartLate(kind: RoundKind, now: Date = new Date()): boolean {
  const round = kind === "monthly" ? currentMonthlyRound(now) : currentYearlyRound(now)
  if (!round) return false
  const startsAt = Math.floor(now.getTime() / MINUTE_MS) * MINUTE_MS + LATE_START_LEAD_MS
  return round.endsAt.getTime() - startsAt >= LATE_START_MIN_WINDOW_MS
}

/** Bangkok calendar year of `now`, the default awards year. It is the calendar year, not the monthly round in progress. */
export function currentVotingYear(now: Date = new Date()): number {
  return votingDateParts(now).year
}

/** Pick the text for the UI locale, falling back to the other language. */
export function localizedText(text: LocalizedText | null | undefined, locale: string): string {
  if (!text) return ""
  const primary = locale === "th" ? text.th : text.en
  const other = locale === "th" ? text.en : text.th
  return primary?.trim() ? primary : other ?? ""
}

/** Translator shape the round copy helpers need; the vue-i18n `t` satisfies it. */
export type EventTranslate = (key: string, named: Record<string, string>, options: { locale: string }) => string

/** Fields the round copy reads; AppEvent and AdminEvent satisfy it. */
export type RoundCopySource = Pick<AppEvent, "type" | "media_type" | "period" | "ends_at" | "title" | "body">

/** Voting rounds whose copy is generated from i18n when the admin leaves both languages blank. */
const GENERATED_ROUND_TYPES: ReadonlySet<EventType> = new Set<EventType>(["vote_monthly", "vote_yearly"])
type RoundField = "title" | "body"

export function isGeneratedRoundType(type: EventType | null | undefined): boolean {
  return !!type && GENERATED_ROUND_TYPES.has(type)
}

const isBlankText = (text: LocalizedText) => !text.en.trim() && !text.th.trim()

/**
 * Label for an award or round period in the current locale, e.g. "Nov 26" / "November 2026" for "2026-11"
 * and "2026" / "2026" for a yearly period. Thai uses the Buddhist era, as Intl does for the th locale.
 */
export function formatAwardPeriod(period: string, locale: string, style: "short" | "long"): string {
  const isMonthly = MONTHLY_PERIOD_RE.test(period)
  const [year, month = 1] = period.split("-").map(Number) as [number, number | undefined]
  const options: Intl.DateTimeFormatOptions = isMonthly
    ? style === "short"
      ? { month: "short", year: "2-digit" }
      : { month: "long", year: "numeric" }
    : { year: "numeric" }
  return new Intl.DateTimeFormat(locale, { ...options, timeZone: VOTING_TIMEZONE }).format(
    new Date(Date.UTC(year, month - 1, 1)),
  )
}

/** Date and time of an instant in the voting timezone, e.g. "Nov 26, 2026, 12:00 AM" (locale-aware). */
export function formatVotingDateTime(iso: string, locale: string): string {
  return new Intl.DateTimeFormat(locale, {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: VOTING_TIMEZONE,
  }).format(new Date(iso))
}

/** Named period params for generated copy: {year} for a yearly round, {month} for a monthly one. */
function periodParams(type: EventType, period: string, locale: string): Record<string, string> {
  const label = formatAwardPeriod(period, locale, "long")
  return type === "vote_yearly" ? { year: label } : { month: label }
}

/** Generated title of a round, e.g. "Popular Movie of the Month - November 2026". */
function roundTitleText(
  type: EventType,
  mediaType: AwardMediaType,
  period: string,
  t: EventTranslate,
  locale: string,
): string {
  return t(`events.generated.title.${type}.${mediaType}`, periodParams(type, period, locale), { locale })
}

/**
 * Generated copy for a voting round in `locale`, ignoring any admin text; null for other events.
 * Monthly copy names the month ({month}), yearly copy names the year ({year}).
 */
function roundCopy(ev: RoundCopySource, field: RoundField, t: EventTranslate, locale: string): string | null {
  if (!isGeneratedRoundType(ev.type) || !ev.media_type || !ev.period) return null
  if (field === "title") return roundTitleText(ev.type, ev.media_type, ev.period, t, locale)
  const key = `events.generated.body.${ev.type}.${ev.media_type}`
  const tz = t("events.generated.timezone", {}, { locale })
  const closes = formatVotingDateTime(ev.ends_at, locale)
  return t(key, { ...periodParams(ev.type, ev.period, locale), closes, tz }, { locale })
}

/**
 * Round title for a notification's params (`event_type`, `media_type`, `period`), in `locale`.
 * Null when any part is missing or malformed, so the caller can fall back to generic text.
 */
export function roundTitleFromParams(
  params: { event_type?: string; media_type?: string; period?: string },
  t: EventTranslate,
  locale: string,
): string | null {
  const { event_type: type, media_type: mediaType, period } = params
  if (type !== "vote_monthly" && type !== "vote_yearly") return null
  if (mediaType !== "movie" && mediaType !== "tv") return null
  if (!period) return null
  const valid = type === "vote_yearly" ? isValidYearlyPeriod(period) : isValidMonthlyPeriod(period)
  return valid ? roundTitleText(type, mediaType, period, t, locale) : null
}

/** Generated title for a voting round in `locale`; also the admin form placeholder. */
export function generatedRoundTitle(ev: RoundCopySource, t: EventTranslate, locale: string): string | null {
  return roundCopy(ev, "title", t, locale)
}

/** Title for display: generated copy when both admin titles are blank, else admin text. */
export function eventTitle(ev: RoundCopySource, t: EventTranslate, locale: string): string {
  if (isBlankText(ev.title)) return roundCopy(ev, "title", t, locale) ?? ""
  return localizedText(ev.title, locale)
}

/** Body for display: generated copy when both admin bodies are blank, else admin text. */
export function eventBody(ev: RoundCopySource, t: EventTranslate, locale: string): string {
  if (isBlankText(ev.body)) return roundCopy(ev, "body", t, locale) ?? ""
  return localizedText(ev.body, locale)
}

/** Local calendar date as YYYY-MM-DD. */
export function localDateKey(date: Date = new Date()): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, "0")
  const d = String(date.getDate()).padStart(2, "0")
  return `${y}-${m}-${d}`
}

const hideKey = (userId: number) => `${EVENT_POPUP_HIDE_KEY_PREFIX}${userId}`
const sessionKey = (userId: number) => `${EVENT_POPUP_SESSION_KEY_PREFIX}${userId}`

function safeGet(storage: Storage | undefined, key: string): string | null {
  try {
    return storage?.getItem(key) ?? null
  } catch {
    return null
  }
}

function safeSet(storage: Storage | undefined, key: string, value: string) {
  try {
    storage?.setItem(key, value)
  } catch {
    /* storage unavailable (private mode); fall back to in-memory only */
  }
}

/** True when the user ticked "don't show again today" today, or closed it this session. */
export function isPopupHidden(userId: number, now: Date = new Date()): boolean {
  if (safeGet(globalThis.localStorage, hideKey(userId)) === localDateKey(now)) return true
  return safeGet(globalThis.sessionStorage, sessionKey(userId)) === "1"
}

/** Remember a dismissal: for today (checkbox) or for this session only. */
export function dismissPopup(userId: number, hideToday: boolean, now: Date = new Date()) {
  if (hideToday) safeSet(globalThis.localStorage, hideKey(userId), localDateKey(now))
  safeSet(globalThis.sessionStorage, sessionKey(userId), "1")
}

/** link_url is internal when it is a path starting with a single "/". */
export function isInternalLink(url: string | null | undefined): boolean {
  return !!url && url.startsWith("/") && !url.startsWith("//")
}

/** Mirrors server validLinkURL: a bounded, whitespace- and backslash-free path or http(s) URL. */
export function isValidLink(url: string): boolean {
  if (!url) return true
  if (new TextEncoder().encode(url).length > EVENT_LINK_MAX || EVENT_LINK_FORBIDDEN_RE.test(url)) return false
  if (isInternalLink(url)) return true
  try {
    const u = new URL(url)
    return u.protocol === "http:" || u.protocol === "https:"
  } catch {
    return false
  }
}

/** Cover image for an event card or banner: the uploaded image, else the TMDB backdrop of the nominee used for voting. */
export function eventCoverUrl(ev: Pick<AppEvent, "image_url" | "cover_backdrop_path">): string | null {
  return ev.image_url || getTmdbImageUrl(ev.cover_backdrop_path, EVENT_COVER_SIZE)
}

/** Format an ISO timestamp with `format`; empty string when the timestamp is invalid. */
export function formatEventDay(iso: string, format: (date: Date) => string): string {
  const date = new Date(iso)
  return Number.isNaN(date.getTime()) ? "" : format(date)
}

export function isValidMonthlyPeriod(period: string): boolean {
  return MONTHLY_PERIOD_RE.test(period)
}

export function isValidYearlyPeriod(period: string): boolean {
  return YEARLY_PERIOD_RE.test(period)
}

/** ISO string -> value for <input type="datetime-local"> in local time. */
export function isoToLocalInput(iso: string | null | undefined): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ""
  const hh = String(d.getHours()).padStart(2, "0")
  const mm = String(d.getMinutes()).padStart(2, "0")
  return `${localDateKey(d)}T${hh}:${mm}`
}

/** <input type="datetime-local"> value -> ISO string ("" when empty/invalid). */
export function localInputToIso(value: string): string {
  if (!value) return ""
  const d = new Date(value)
  return isNaN(d.getTime()) ? "" : d.toISOString()
}

export type EventFormField = "title_en" | "title_th" | "body_en" | "body_th" | "link_url" | "starts_at" | "ends_at"
export type EventFormErrors = Partial<Record<EventFormField, string>>

const runes = (s: string) => [...s].length

/** Client-side mirror of the server validation; values are i18n keys. */
export function validateEventForm(body: AdminEventBody): EventFormErrors {
  const errors: EventFormErrors = {}
  // Generated voting rounds may leave titles blank; the generated copy fills them in.
  const titleRequired = !isGeneratedRoundType(body.type)
  for (const f of ["title_en", "title_th"] as const) {
    const v = body[f].trim()
    if (!v) {
      if (titleRequired) errors[f] = "admin.events.errors.titleRequired"
    } else if (runes(v) > EVENT_TITLE_MAX) errors[f] = "admin.events.errors.titleTooLong"
  }
  for (const f of ["body_en", "body_th"] as const) {
    if (runes(body[f]) > EVENT_BODY_MAX) errors[f] = "admin.events.errors.bodyTooLong"
  }
  if (!isValidLink(body.link_url.trim())) errors.link_url = "admin.events.errors.linkInvalid"
  if (!body.starts_at) errors.starts_at = "admin.events.errors.dateRequired"
  if (!body.ends_at) errors.ends_at = "admin.events.errors.dateRequired"
  if (body.starts_at && body.ends_at && new Date(body.starts_at) >= new Date(body.ends_at)) {
    errors.ends_at = "admin.events.errors.endBeforeStart"
  }
  return errors
}

type ErrShape = { response?: { status?: number; data?: { code?: string } } }

/** Map an admin events API failure to an i18n key; never surfaces raw server text. */
export function eventErrorKey(err: unknown): string {
  const res = (err as ErrShape)?.response
  const code = res?.data?.code
  if (res?.status === 413) return "admin.events.errors.imageTooLarge"
  if (code === "UNSUPPORTED_MEDIA_TYPE" || res?.status === 415) return "admin.events.errors.imageType"
  if (code === "EVENT_ALREADY_EXISTS") return "admin.events.errors.eventExists"
  if (code === "MONTHLY_ROUNDS_PENDING") return "admin.events.errors.monthlyRoundsPending"
  if (code === "VOTING_LOCKED") return "admin.events.errors.votingLocked"
  if (code === "EVENT_NOT_PUBLISHED") return "admin.events.errors.notPublished"
  if (code === "VOTING_NOT_ENDED") return "admin.events.errors.votingNotEnded"
  if (code === "NOT_ENOUGH_NOMINEES") return "admin.events.errors.notEnoughNominees"
  if (code === "INVALID_INPUT") return "admin.events.errors.invalidInput"
  if (code === "INVALID_EVENT_TYPE") return "admin.events.errors.invalidType"
  if (code === "VOTING_WINDOW_TOO_SHORT") return "admin.events.errors.votingWindowTooShort"
  if (code === "NOT_FOUND" || res?.status === 404) return "admin.events.errors.notFound"
  if (code === "INVALID_ID") return "admin.events.errors.notFound"
  if (code === "EVENT_NOT_DELETABLE") return "admin.events.errors.notDeletable"
  if (code === "TOO_FEW_YEARLY_NOMINEES") return "admin.events.errors.tooFewYearlyNominees"
  if (res?.status === 401 || res?.status === 403) return "admin.errors.forbidden"
  return "admin.errors.generic"
}

/**
 * Map a failed publish (event create/update or nominee save): NOT_ENOUGH_NOMINEES here means a
 * published round is short of titles, which reads differently from generating a round.
 */
export function publishErrorKey(err: unknown): string {
  if ((err as ErrShape)?.response?.data?.code === "NOT_ENOUGH_NOMINEES") {
    return "admin.events.errors.publishNeedsNominees"
  }
  return eventErrorKey(err)
}

/** A voting round can be deleted only while it has no votes and is not finalized (server rule). */
export function isDeleteBlocked(ev: Pick<AdminEvent, "type" | "is_finalized" | "has_votes">): boolean {
  return isGeneratedRoundType(ev.type) && (!!ev.is_finalized || !!ev.has_votes)
}

/** Map a failed auto publish toggle to an i18n key; other failures fall back to the generic event mapping. */
export function autoPublishErrorKey(err: unknown): string {
  const code = (err as ErrShape)?.response?.data?.code
  if (code === "EVENT_NOT_DRAFT") return "admin.events.errors.notDraft"
  if (code === "VOTING_CLOSED") return "admin.events.errors.votingClosed"
  return eventErrorKey(err)
}

/** Map a user voting API failure to an i18n key. */
export function voteErrorKey(err: unknown): string {
  const code = (err as ErrShape)?.response?.data?.code
  if (code === "VOTING_CLOSED") return "events.voting.errors.closed"
  if (code === "VOTING_NOT_STARTED") return "events.voting.errors.notStarted"
  if (code === "NOMINEE_NOT_FOUND") return "events.voting.errors.nomineeNotFound"
  if (code === "ACCOUNT_TOO_NEW") return "events.voting.errors.accountTooNew"
  return "common.error"
}
