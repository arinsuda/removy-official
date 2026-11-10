// Mailed / redirected secrets (verify, reset, handoff tokens) travel in the URL fragment so they
// never reach servers, proxies or Referer headers. Read them once and strip them at once. Never log them.

export const URL_SECRET_KEYS = {
  token: "token",
  uid: "uid",
  handoff: "t",
} as const

// The legacy `?token=` query fallback stops working on this date; TODO: delete the fallback code in a follow-up.
export const LEGACY_QUERY_TOKEN_UNTIL = "2026-11-17"

interface ConsumeOptions {
  /** Also accept the old `?key=` form when the fragment lacks the keys. */
  legacyQuery?: boolean
}

// vue-router keeps the full location (with `#...` / `?...`) in history.state.current; drop it there too.
function stateWithoutSecrets(state: unknown): unknown {
  if (!state || typeof state !== "object") return state
  const s = state as Record<string, unknown>
  return typeof s.current === "string" ? { ...s, current: s.current.split(/[?#]/)[0] } : state
}

export function consumeUrlSecrets(keys: readonly string[], opts: ConsumeOptions = {}): Record<string, string> {
  let params = new URLSearchParams(window.location.hash.replace(/^#/, ""))
  const legacyOpen = Date.now() < Date.parse(LEGACY_QUERY_TOKEN_UNTIL)
  if (!keys.some((k) => params.has(k)) && opts.legacyQuery && legacyOpen) {
    params = new URLSearchParams(window.location.search)
  }
  const out: Record<string, string> = {}
  for (const k of keys) {
    const v = params.get(k)
    if (v) out[k] = v
  }
  // Always clean the address bar and history entry, even when nothing was found.
  window.history.replaceState(stateWithoutSecrets(window.history.state), "", window.location.pathname)
  return out
}
