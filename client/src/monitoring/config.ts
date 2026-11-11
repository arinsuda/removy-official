// Sentry settings. Reporting is off unless VITE_SENTRY_DSN is set.
export const SENTRY_TRACES_SAMPLE_RATE = 0.05
export const SENTRY_DEFAULT_ENVIRONMENT = "production"
export const SENTRY_MAX_BREADCRUMBS = 50
/** HTTP statuses below this are expected client errors and are not reported. */
export const SENTRY_MIN_REPORTED_STATUS = 500
export const SENTRY_REDACTED = "[redacted]"
export const SENTRY_REDACTED_EMAIL = "[email]"
/** Query params whose value is dropped. */
export const SENTRY_SENSITIVE_PARAM = /token|code|secret|password|key|auth|otp|session|state/i
