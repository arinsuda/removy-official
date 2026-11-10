// Mailed / redirected secrets (verify, reset, handoff tokens) travel in the URL fragment so they
// never reach servers, proxies or Referer headers. Read them once and strip them at once. Never log them.

export const URL_SECRET_KEYS = {
  token: "token",
  uid: "uid",
  handoff: "t",
} as const

// vue-router keeps the full location (with `#...` / `?...`) in history.state.current; drop it there too.
function stateWithoutSecrets(state: unknown): unknown {
  if (!state || typeof state !== "object") return state
  const s = state as Record<string, unknown>
  return typeof s.current === "string" ? { ...s, current: s.current.split(/[?#]/)[0] } : state
}

export function consumeUrlSecrets(keys: readonly string[]): Record<string, string> {
  // Secrets are read from the fragment only; a stray `?token=` is never read, just stripped below.
  const params = new URLSearchParams(window.location.hash.replace(/^#/, ""))
  const out: Record<string, string> = {}
  for (const k of keys) {
    const v = params.get(k)
    if (v) out[k] = v
  }
  // Always clean the address bar and history entry, even when nothing was found.
  window.history.replaceState(stateWithoutSecrets(window.history.state), "", window.location.pathname)
  return out
}
