import { isAxiosError } from "axios"
import {
  MANGA_ADULT_ACK_KEY_PREFIX,
  MANGA_ADULT_MIN_AGE,
  MANGA_AGE_RESTRICTED_CODE,
  MANGA_AGE_TIME_OFFSET_MS,
} from "./mangaConfig"

const ACK_VALUE = "1"

/**
 * True when a YYYY-MM-DD birth date makes the person at least MANGA_ADULT_MIN_AGE on the Asia/Bangkok
 * calendar day of `today`. Days are compared as Y/M/D, so the unlock falls on the same day as the server's.
 */
export function isAdultByBirth(dateOfBirth: string | null | undefined, today: Date): boolean {
  if (!dateOfBirth) return false
  const [year, month, day] = dateOfBirth.split("-").map(Number)
  if (!year || !month || !day) return false
  // Date.UTC rolls 29 February into 1 March in non-leap years, as the server's time.Date does.
  const eighteenth = Date.UTC(year + MANGA_ADULT_MIN_AGE, month - 1, day)
  const local = new Date(today.getTime() + MANGA_AGE_TIME_OFFSET_MS)
  const todayStart = Date.UTC(local.getUTCFullYear(), local.getUTCMonth(), local.getUTCDate())
  return eighteenth <= todayStart
}

/** Reads the per-user 18+ acknowledgement. Storage can be blocked, so failures mean "not acknowledged". */
export function readAdultAck(userId: number | null): boolean {
  if (userId === null) return false
  try {
    return localStorage.getItem(MANGA_ADULT_ACK_KEY_PREFIX + userId) === ACK_VALUE
  } catch {
    return false
  }
}

/** Remembers the acknowledgement for this user. A failed write only means the modal shows again. */
export function writeAdultAck(userId: number): void {
  try {
    localStorage.setItem(MANGA_ADULT_ACK_KEY_PREFIX + userId, ACK_VALUE)
  } catch {
    // Storage unavailable (private mode or quota): the acknowledgement lasts for this page only.
  }
}

/** True for the API's AGE_RESTRICTED 403 on an adult title. */
export function isAgeRestricted(error: unknown): boolean {
  if (!isAxiosError(error)) return false
  const data = error.response?.data as { code?: unknown } | undefined
  return data?.code === MANGA_AGE_RESTRICTED_CODE
}
