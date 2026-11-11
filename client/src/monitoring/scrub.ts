import type { BrowserOptions, Breadcrumb, ErrorEvent, EventHint } from "@sentry/vue"
// @sentry/vue does not re-export TransactionEvent; derive it from the public options type.
export type TransactionEvent = Parameters<NonNullable<BrowserOptions["beforeSendTransaction"]>>[0]

import {
  SENTRY_MIN_REPORTED_STATUS,
  SENTRY_REDACTED,
  SENTRY_REDACTED_EMAIL,
  SENTRY_SENSITIVE_PARAM,
} from "./config"

const EMAIL_PATTERN = /[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}/g

// Router query values (query.<key>), raw query strings and URL fragments can carry search text or tokens.
const DROPPED_ATTRIBUTE_PREFIXES = ["query.", "url.query", "http.query", "url.fragment"] as const

function safeDecode(text: string): string | null {
  try {
    return decodeURIComponent(text)
  } catch {
    return null
  }
}

export function scrubText(text: string): string {
  const decoded = safeDecode(text)
  const source = decoded !== null && EMAIL_PATTERN.test(decoded) ? decoded : text
  EMAIL_PATTERN.lastIndex = 0
  return source.replace(EMAIL_PATTERN, SENTRY_REDACTED_EMAIL)
}

/** Drops the fragment (it carries handoff/verify/reset tokens) and sensitive query values. */
export function scrubUrl(url: string): string {
  const noFragment = url.split("#")[0] ?? ""
  const [base, query] = noFragment.split("?")
  if (!query) return scrubText(base ?? "")
  const params = query
    .split("&")
    .map((pair) => {
      const [key = "", value] = pair.split("=")
      const decodedKey = safeDecode(key)
      // An undecodable key is treated as sensitive.
      return value !== undefined && (decodedKey === null || SENTRY_SENSITIVE_PARAM.test(decodedKey))
        ? `${key}=${SENTRY_REDACTED}`
        : scrubText(pair)
    })
    .join("&")
  return `${base}?${params}`
}

function isExpectedHttpError(hint?: EventHint): boolean {
  const err = hint?.originalException as { isAxiosError?: boolean; response?: { status?: number } } | undefined
  const status = err?.isAxiosError ? err.response?.status : undefined
  return status !== undefined && status < SENTRY_MIN_REPORTED_STATUS
}

export function scrubBreadcrumb(crumb: Breadcrumb): Breadcrumb | null {
  const out: Breadcrumb = { ...crumb }
  if (out.message) out.message = scrubText(out.message)
  if (out.data) {
    const data = { ...out.data }
    for (const key of ["url", "from", "to"]) {
      if (typeof data[key] === "string") data[key] = scrubUrl(data[key])
    }
    delete data.request_body_size
    if (out.category === "console") delete data.arguments
    out.data = data
  }
  return out
}

/** beforeSend: skips expected 4xx axios errors and strips PII, request data and tokens. */
export function scrubEvent(event: ErrorEvent, hint?: EventHint): ErrorEvent | null {
  if (isExpectedHttpError(hint)) return null
  if (event.message) event.message = scrubText(event.message)
  for (const ex of event.exception?.values ?? []) {
    if (ex.value) ex.value = scrubText(ex.value)
  }
  if (event.request) {
    const { method, url } = event.request
    event.request = { method, url: url ? scrubUrl(url) : undefined }
  }
  if (event.contexts) delete event.contexts.vue
  if (event.user) event.user = event.user.id ? { id: event.user.id } : undefined
  if (event.breadcrumbs) {
    event.breadcrumbs = event.breadcrumbs.map(scrubBreadcrumb).filter((b): b is Breadcrumb => b !== null)
  }
  return event
}

const URL_ATTRIBUTES = ["url.full", "http.url", "http.target", "url"]

function scrubAttributes(data: Record<string, unknown> | undefined): void {
  if (!data) return
  for (const key of Object.keys(data)) {
    if (DROPPED_ATTRIBUTE_PREFIXES.some((prefix) => key.startsWith(prefix))) delete data[key]
    else if (URL_ATTRIBUTES.includes(key) && typeof data[key] === "string") data[key] = scrubUrl(data[key])
  }
}

/** beforeSendTransaction: transactions skip beforeSend, so scrub URLs, query attributes and fragments here. */
export function scrubTransaction(event: TransactionEvent): TransactionEvent | null {
  if (event.transaction) event.transaction = scrubUrl(event.transaction)
  if (event.request) {
    const { method, url } = event.request
    event.request = { method, url: url ? scrubUrl(url) : undefined }
  }
  if (event.contexts) delete event.contexts.vue
  if (event.user) event.user = event.user.id ? { id: event.user.id } : undefined
  scrubAttributes(event.contexts?.trace?.data)
  for (const span of event.spans ?? []) {
    scrubAttributes(span.data)
    if (span.description) span.description = scrubUrl(span.description)
  }
  return event
}
