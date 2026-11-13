import type { AdminEventBody, LocalizedText } from "@/types/event"

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
const DAY_MS = 24 * 60 * 60 * 1000
/** Default window for a new event in the admin form: starts now, lasts two weeks. */
export const EVENT_DEFAULT_DURATION_MS = 14 * DAY_MS

/** Pick the text for the UI locale, falling back to the other language. */
export function localizedText(text: LocalizedText | null | undefined, locale: string): string {
  if (!text) return ""
  const primary = locale === "th" ? text.th : text.en
  const other = locale === "th" ? text.en : text.th
  return primary?.trim() ? primary : other ?? ""
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
  for (const f of ["title_en", "title_th"] as const) {
    const v = body[f].trim()
    if (!v) errors[f] = "admin.events.errors.titleRequired"
    else if (runes(v) > EVENT_TITLE_MAX) errors[f] = "admin.events.errors.titleTooLong"
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
  if (code === "INVALID_INPUT") return "admin.events.errors.invalidInput"
  if (code === "NOT_FOUND" || res?.status === 404) return "admin.events.errors.notFound"
  if (code === "INVALID_ID") return "admin.events.errors.notFound"
  if (res?.status === 401 || res?.status === 403) return "admin.errors.forbidden"
  return "admin.errors.generic"
}
