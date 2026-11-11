import { afterEach, describe, expect, it, vi } from "vitest"
import type { TransactionEvent } from "@/monitoring/scrub"
import type { ErrorEvent } from "@sentry/vue"
import { scrubBreadcrumb, scrubEvent, scrubTransaction, scrubUrl } from "@/monitoring/scrub"

const initMock = vi.hoisted(() => vi.fn())
vi.mock("@sentry/vue", () => ({
  init: initMock,
  browserTracingIntegration: vi.fn(() => ({})),
}))

describe("sentry scrub", () => {
  it("strips the URL fragment and token query values", () => {
    expect(scrubUrl("https://a.test/auth/verify-email#token=abc")).toBe("https://a.test/auth/verify-email")
    expect(scrubUrl("https://a.test/x?token=abc&page=2")).toBe("https://a.test/x?token=[redacted]&page=2")
  })

  it("redacts emails and drops request cookies, headers and body", () => {
    const event = {
      type: undefined,
      message: "fail for bob@example.com",
      request: { url: "https://a.test/p?code=1#f", cookies: { a: "b" }, headers: { Authorization: "x" }, data: "body" },
      user: { id: "5", email: "bob@example.com" },
    } as unknown as ErrorEvent
    const out = scrubEvent(event)!
    expect(out.message).toBe("fail for [email]")
    expect(out.request).toEqual({ method: undefined, url: "https://a.test/p?code=[redacted]" })
    expect(out.user).toEqual({ id: "5" })
  })

  it("skips expected 4xx axios errors but keeps 5xx and network errors", () => {
    const ev = () => ({}) as ErrorEvent
    const axiosErr = (status?: number) => ({ isAxiosError: true, response: status ? { status } : undefined })
    expect(scrubEvent(ev(), { originalException: axiosErr(401) })).toBeNull()
    expect(scrubEvent(ev(), { originalException: axiosErr(429) })).toBeNull()
    expect(scrubEvent(ev(), { originalException: axiosErr(503) })).not.toBeNull()
    expect(scrubEvent(ev(), { originalException: axiosErr() })).not.toBeNull()
  })
})

describe("sentry scrub extras", () => {
  it("drops vue props context", () => {
    const out = scrubEvent({ contexts: { vue: { propsData: { password: "x" } } } } as unknown as ErrorEvent)!
    expect(out.contexts?.vue).toBeUndefined()
  })

  it("scrubs transactions (request url, span urls, query attributes)", () => {
    const ev = {
      transaction: "/auth/google#access_token=abc",
      request: { url: "https://a.test/cb?code=1#access_token=abc" },
      contexts: { trace: { data: { "url.full": "https://a.test/x?token=1#f", "url.query": "token=1", "query.token": "x", "query.q": "batman", "url.fragment": "f" } } },
      spans: [{ data: { "http.url": "https://a.test/y#h", "url.query": "a=b" }, description: "GET /z?code=9" }],
    } as unknown as TransactionEvent
    const out = scrubTransaction(ev)!
    expect(out.transaction).toBe("/auth/google")
    expect(out.request?.url).toBe("https://a.test/cb?code=[redacted]")
    expect(out.contexts?.trace?.data).toEqual({ "url.full": "https://a.test/x?token=[redacted]" })
    expect(out.spans?.[0]?.data).toEqual({ "http.url": "https://a.test/y" })
    expect(out.spans?.[0]?.description).toBe("GET /z?code=[redacted]")
  })

  it("does not throw on bad percent-encoding and redacts encoded emails", () => {
    expect(scrubUrl("https://a.test/x?%E0%A4=1")).toBe("https://a.test/x?%E0%A4=[redacted]")
    expect(scrubUrl("https://a.test/x?q=bob%40example.com")).toBe("https://a.test/x?q=[email]")
  })

  it("drops console breadcrumb arguments", () => {
    const out = scrubBreadcrumb({ category: "console", data: { arguments: ["pw"], logger: "console" } })!
    expect(out.data).toEqual({ logger: "console" })
  })
})

describe("initSentry", () => {
  afterEach(() => {
    vi.unstubAllEnvs()
    initMock.mockClear()
  })

  it("is skipped without a DSN", async () => {
    vi.stubEnv("VITE_SENTRY_DSN", "")
    const { initSentry } = await import("@/monitoring/sentry")
    expect(initSentry({} as never, {} as never)).toBe(false)
    expect(initMock).not.toHaveBeenCalled()
  })

  it("initializes when a DSN is set", async () => {
    vi.stubEnv("VITE_SENTRY_DSN", "https://k@o.ingest.sentry.io/1")
    const { initSentry } = await import("@/monitoring/sentry")
    expect(initSentry({} as never, {} as never)).toBe(true)
    expect(initMock.mock.calls[0]?.[0]).toMatchObject({
      attachProps: false,
      dataCollection: { userInfo: false, cookies: false, httpHeaders: false, urlQueryParams: false },
    })
  })
})
