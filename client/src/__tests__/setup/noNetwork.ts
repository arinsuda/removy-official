// Hermetic test guard: any request that reaches a real transport fails at once, naming method + URL.
// Tests must mock the api module/endpoint, set their own adapter, or vi.spyOn(globalThis, "fetch").
import { afterEach, beforeEach, vi } from "vitest"
import type { AxiosAdapter, InternalAxiosRequestConfig } from "axios"
import axios from "axios"

export const NO_NETWORK_PREFIX = "[no-network] Unmocked network request blocked:"

// Recorded so a test fails even when app code swallows the rejection (catch -> error state).
const blocked: string[] = []

const describeRequest = (method: string | undefined, url: string | undefined) => {
  const message = `${NO_NETWORK_PREFIX} ${(method ?? "GET").toUpperCase()} ${url ?? "<unknown url>"}. Mock it in the test.`
  blocked.push(message)
  return message
}

const blockedAdapter: AxiosAdapter = (config: InternalAxiosRequestConfig) =>
  Promise.reject(new Error(describeRequest(config.method, `${config.baseURL ?? ""}${config.url ?? ""}`)))

const blockedFetch = (input: RequestInfo | URL, init?: RequestInit) => {
  const url = input instanceof Request ? input.url : String(input)
  return Promise.reject(new Error(describeRequest(init?.method ?? (input instanceof Request ? input.method : undefined), url)))
}

class BlockedXMLHttpRequest {
  private method = "GET"
  private target = ""
  open(method: string, url: string | URL) {
    this.method = method
    this.target = String(url)
  }
  send() {
    throw new Error(describeRequest(this.method, this.target))
  }
  setRequestHeader() {}
  addEventListener() {}
  removeEventListener() {}
  abort() {}
}

// Global default: every axios.create() made after this (including after vi.resetModules) inherits it.
// Not importing @/api here keeps specs' own vi.mock() of its dependencies (socket.io-client...) effective.
axios.defaults.adapter = blockedAdapter

// Re-stub before every test so a previous vi.unstubAllGlobals() cannot re-open the network;
// a test's own vi.spyOn / vi.stubGlobal after this point still wins.
beforeEach(() => {
  vi.stubGlobal("fetch", blockedFetch)
  vi.stubGlobal("XMLHttpRequest", BlockedXMLHttpRequest)
})

afterEach(() => {
  const attempts = blocked.splice(0)
  if (attempts.length) throw new Error(attempts.join("\n"))
})
