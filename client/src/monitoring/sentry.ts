import type { App } from "vue"
import type { Router } from "vue-router"
import * as Sentry from "@sentry/vue"
import {
  SENTRY_DEFAULT_ENVIRONMENT,
  SENTRY_MAX_BREADCRUMBS,
  SENTRY_TRACES_SAMPLE_RATE,
} from "./config"
import { scrubBreadcrumb, scrubEvent, scrubTransaction } from "./scrub"

/** Starts Sentry only when VITE_SENTRY_DSN is set. Returns whether it is enabled. */
export function initSentry(app: App, router: Router): boolean {
  const dsn = import.meta.env.VITE_SENTRY_DSN
  if (!dsn) return false
  Sentry.init({
    app,
    dsn,
    environment: import.meta.env.VITE_SENTRY_ENVIRONMENT || SENTRY_DEFAULT_ENVIRONMENT,
    release: import.meta.env.VITE_SENTRY_RELEASE || undefined,
    // v11 replaced sendDefaultPii with dataCollection; collect nothing request/user related.
    dataCollection: { userInfo: false, cookies: false, httpHeaders: false, httpBodies: [], urlQueryParams: false },
    attachProps: false, // component props can hold secrets (e.g. a password field)
    maxBreadcrumbs: SENTRY_MAX_BREADCRUMBS,
    tracesSampleRate: SENTRY_TRACES_SAMPLE_RATE,
    integrations: [Sentry.browserTracingIntegration({ router })],
    beforeSend: scrubEvent,
    beforeSendTransaction: scrubTransaction,
    beforeBreadcrumb: scrubBreadcrumb,
  })
  return true
}
