import type { AdminEventBody, AppEvent, EventType, LocalizedText } from "@/types/event"
import type { TmdbImageSize } from "@/utils/image"

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
export const EVENT_POPUP_PRIORITY_MIN = -100
export const EVENT_POPUP_PRIORITY_MAX = 100
/** TMDB poster width for nominee cards on the event detail page. */
export const NOMINEE_POSTER_SIZE: TmdbImageSize = "w342"
const DAY_MS = 24 * 60 * 60 * 1000
/** Default window for a new event in the admin form: starts now, lasts two weeks. */
export const EVENT_DEFAULT_DURATION_MS = 14 * DAY_MS
/** Day of month (Thailand time) on which monthly voting rounds roll over to the next month. */
export const VOTING_CUTOFF_DAY = 15
export const VOTING_TIMEZONE = "Asia/Bangkok"
const MONTHS_PER_YEAR = 12
const PAD_WIDTH = 2

/** Shift a calendar month by `delta` months, crossing year boundaries. */
function shiftMonth(year: number, month: number, delta: number): { year: number; month: number } {
  const index = year * MONTHS_PER_YEAR + (month - 1) + delta
  return { year: Math.floor(index / MONTHS_PER_YEAR), month: (index % MONTHS_PER_YEAR) + 1 }
}

/** Default YYYY-MM period for the monthly voting round after the one active at `now` (Thailand time). */
export function nextMonthlyPeriod(now: Date = new Date()): string {
  const parts = new Intl.DateTimeFormat("en-CA", { // hardcode-ok: machine YYYY-MM parts
    timeZone: VOTING_TIMEZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(now)
  const year = Number(parts.find((p) => p.type === "year")?.value ?? now.getUTCFullYear())
  const month = Number(parts.find((p) => p.type === "month")?.value ?? now.getUTCMonth() + 1)
  const day = Number(parts.find((p) => p.type === "day")?.value ?? now.getUTCDate())
  // From the cutoff day on, the active round is already the following month, so the next one is two months ahead.
  const monthsAhead = day >= VOTING_CUTOFF_DAY ? 2 : 1
  const next = shiftMonth(year, month, monthsAhead)
  return `${next.year}-${String(next.month).padStart(PAD_WIDTH, "0")}`
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

/** Only monthly rounds are generated; yearly rounds always use admin text. */
const GENERATED_ROUND_TYPE: EventType = "vote_monthly"
type RoundField = "title" | "body"

const isBlankText = (text: LocalizedText) => !text.en.trim() && !text.th.trim()

/** YYYY-MM period as a month name with year, e.g. "November 2026" or "พฤศจิกายน 2569". */
function formatRoundMonth(period: string, locale: string): string {
  const [year, month] = period.split("-").map(Number) as [number, number]
  return new Intl.DateTimeFormat(locale, { month: "long", year: "numeric", timeZone: VOTING_TIMEZONE }).format(
    new Date(Date.UTC(year, month - 1, 1)),
  )
}

function formatClosing(iso: string, locale: string): string {
  return new Intl.DateTimeFormat(locale, {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: VOTING_TIMEZONE,
  }).format(new Date(iso))
}

/** Generated copy for a monthly round in `locale`, ignoring any admin text; null for other events. */
function roundCopy(ev: RoundCopySource, field: RoundField, t: EventTranslate, locale: string): string | null {
  if (ev.type !== GENERATED_ROUND_TYPE || !ev.media_type || !ev.period) return null
  const key = `events.generated.${field}.${ev.type}.${ev.media_type}`
  const month = formatRoundMonth(ev.period, locale)
  if (field === "title") return t(key, { month }, { locale })
  const tz = t("events.generated.timezone", {}, { locale })
  return t(key, { month, closes: formatClosing(ev.ends_at, locale), tz }, { locale })
}

/** Generated title for a monthly round in `locale`; also the admin form placeholder. */
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

export function isValidLink(url: string): boolean {
  if (!url) return true
  if (isInternalLink(url)) return true
  try {
    const u = new URL(url)
    return u.protocol === "http:" || u.protocol === "https:"
  } catch {
    return false
  }
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
  // Generated monthly rounds may leave titles blank; the generated copy fills them in.
  const titleRequired = body.type !== GENERATED_ROUND_TYPE
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
  if (code === "VOTING_LOCKED") return "admin.events.errors.votingLocked"
  if (code === "EVENT_NOT_PUBLISHED") return "admin.events.errors.notPublished"
  if (code === "VOTING_NOT_ENDED") return "admin.events.errors.votingNotEnded"
  if (code === "NOT_ENOUGH_NOMINEES") return "admin.events.errors.notEnoughNominees"
  if (code === "INVALID_INPUT") return "admin.events.errors.invalidInput"
  if (code === "INVALID_EVENT_TYPE") return "admin.events.errors.invalidType"
  if (code === "NOT_FOUND" || res?.status === 404) return "admin.events.errors.notFound"
  if (code === "INVALID_ID") return "admin.events.errors.notFound"
  if (res?.status === 401 || res?.status === 403) return "admin.errors.forbidden"
  return "admin.errors.generic"
}

/** Map a user voting API failure to an i18n key. */
export function voteErrorKey(err: unknown): string {
  const code = (err as ErrShape)?.response?.data?.code
  if (code === "VOTING_CLOSED") return "events.voting.errors.closed"
  if (code === "VOTING_NOT_STARTED") return "events.voting.errors.notStarted"
  if (code === "NOMINEE_NOT_FOUND") return "events.voting.errors.nomineeNotFound"
  return "common.error"
}
