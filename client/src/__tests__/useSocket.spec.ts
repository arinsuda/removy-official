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
  connect: ReturnType<typeof vi.fn>
  connected: boolean
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
      connect: vi.fn(),
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

  it("on 'unauthorized' (event or connect_error) disconnects only, keeping socket and listeners", () => {
    const { connect, isConnecting } = useSocket()
    connect()
    const first = created[0]
    first.handlers.error("something else")
    expect(first.disconnect).not.toHaveBeenCalled()

    first.handlers.error("unauthorized")
    first.handlers.connect_error(new Error("unauthorized"))
    expect(first.disconnect).toHaveBeenCalledTimes(2)
    expect(first.removeAllListeners).not.toHaveBeenCalled()
    expect(first.handlers.error).toBeTypeOf("function")
    expect(isConnecting.value).toBe(false)

    connect()
    expect(created).toHaveLength(1)
  })

  it("resume() reconnects the same socket with the newest token, and is a no-op when connected", () => {
    const { connect, resume } = useSocket()
    connect()
    const s = created[0]
    s.handlers.error("unauthorized")

    localStorage.setItem("access_token", "at-2")
    resume()
    expect(s.auth).toEqual({ token: "at-2" })
    expect(s.connect).toHaveBeenCalledTimes(1)

    s.connected = true
    resume()
    expect(s.connect).toHaveBeenCalledTimes(1)
  })

  it("resume() does nothing without a token or before the first connect", () => {
    const { connect, resume } = useSocket()
    resume()
    connect()
    localStorage.removeItem("access_token")
    resume()
    expect(created[0].connect).not.toHaveBeenCalled()
  })

  it("resumes when another tab writes a new access token", () => {
    useSocket().connect()
    localStorage.setItem("access_token", "at-3")
    window.dispatchEvent(new StorageEvent("storage", { key: "access_token" }))
    expect(created[0].auth).toEqual({ token: "at-3" })
    expect(created[0].connect).toHaveBeenCalledTimes(1)
  })
})
