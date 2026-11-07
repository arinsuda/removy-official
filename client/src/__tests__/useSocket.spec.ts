/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect, vi, beforeEach } from "vitest"

type Handler = (...args: unknown[]) => void
const created = vi.hoisted(() => [] as Array<{
  opts: Record<string, unknown>
  auth: unknown
  handlers: Record<string, Handler>
  disconnect: ReturnType<typeof vi.fn>
  io: { on: ReturnType<typeof vi.fn>; opts: Record<string, unknown> }
}>)

vi.mock("socket.io-client", () => ({
  io: (_url: string, opts: Record<string, unknown>) => {
    const handlers: Record<string, Handler> = {}
    const sock = {
      opts,
      auth: opts.auth,
      handlers,
      connected: false,
      on: (ev: string, h: Handler) => void (handlers[ev] = h),
      off: vi.fn(),
      removeAllListeners: vi.fn(() => Object.keys(handlers).forEach((k) => delete handlers[k])),
      disconnect: vi.fn(),
      io: { on: vi.fn(), opts: {} },
    }
    created.push(sock)
    return sock
  },
}))

import { useSocket } from "@/composables/useSocket"

describe("useSocket", () => {
  beforeEach(() => {
    created.length = 0
    localStorage.setItem("access_token", "at-1")
    useSocket().disconnect()
  })

  it("sends the token in the handshake auth only, never in the URL query", () => {
    useSocket().connect()
    expect(created[0].opts.auth).toEqual({ token: "at-1" })
    expect(created[0].opts).not.toHaveProperty("query")
  })

  it("stops for good on 'unauthorized' instead of looping, and a later connect() starts fresh", () => {
    const { connect, isConnecting } = useSocket()
    connect()
    const first = created[0]
    const onError = first.handlers.error
    first.handlers.error("something else")
    expect(first.disconnect).not.toHaveBeenCalled()

    onError("unauthorized")
    expect(first.disconnect).toHaveBeenCalledTimes(1)
    expect(isConnecting.value).toBe(false)

    connect()
    expect(created).toHaveLength(2)
  })
})
