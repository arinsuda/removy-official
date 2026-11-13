/** Client-side shape check only; the server is the authority on what a valid address is. */
export const EMAIL_FORMAT_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

/** How long the "email changed" confirmation stays before the card returns to idle. */
export const EMAIL_CHANGED_NOTICE_MS = 2500

/** Length of each emailed one-time code. */
export const EMAIL_OTP_LENGTH = 6

/** Digit placeholder for a code input (not translatable copy). */
export const OTP_PLACEHOLDER = "0".repeat(EMAIL_OTP_LENGTH)

export const EMAIL_SEND_FAILED_KEY = "settings.email.errors.sendFailed"

/** Server error codes of POST /users/:id/email -> i18n keys. */
export const EMAIL_ERROR_KEYS: Readonly<Record<string, string>> = {
  EMAIL_IN_USE: "settings.email.errors.inUse",
  EMAIL_INVALID: "settings.email.errors.invalid",
  EMAIL_SAME: "settings.email.errors.same",
}

/** The i18n key for a failed email-change request; unknown codes fall back to the generic message. */
export function emailErrorKey(err: unknown): string {
  const code = (err as { response?: { data?: { code?: string } } })?.response?.data?.code
  return (code && EMAIL_ERROR_KEYS[code]) || EMAIL_SEND_FAILED_KEY
}
