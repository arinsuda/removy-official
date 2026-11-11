import type { LocationQueryValue } from "vue-router"

const ENCODED_SLASH_PREFIX_RE = /^\/(?:%2f|%5c)/i
const DUMMY_ORIGIN = "https://local.invalid"
const AUTH_PREFIXES = [
  "/login",
  "/register",
  "/forgot-password",
  "/reset-password",
  "/check-email",
] as const

function firstQueryValue(
  raw: LocationQueryValue | LocationQueryValue[] | unknown,
): string | null {
  const value = Array.isArray(raw) ? raw[0] : raw
  return typeof value === "string" ? value : null
}

function hasControlChars(value: string): boolean {
  for (let i = 0; i < value.length; i++) {
    const code = value.charCodeAt(i)
    if (code <= 0x1f || code === 0x7f) return true
  }
  return false
}

function isAuthLoopPath(pathname: string): boolean {
  return AUTH_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  )
}

/**
 * Accepts only same-origin relative paths (starting with a single "/").
 * Rejects scheme URLs, protocol-relative URLs ("//"), backslashes, control chars,
 * and auth entry paths to prevent open redirects and redirect loops. The hash is dropped
 * because it can carry secrets (e.g. access_token).
 */
export function sanitizeRedirect(raw: unknown): string | null {
  const candidate = firstQueryValue(raw)
  if (!candidate || hasControlChars(candidate)) return null

  const value = candidate.trim()
  if (
    !value.startsWith("/") ||
    value.startsWith("//") ||
    value.startsWith("/\\") ||
    value.includes("\\")
  ) {
    return null
  }

  try {
    const parsed = new URL(value, DUMMY_ORIGIN)
    if (parsed.origin !== DUMMY_ORIGIN) return null
    if (
      !parsed.pathname.startsWith("/") ||
      parsed.pathname.startsWith("//") ||
      ENCODED_SLASH_PREFIX_RE.test(parsed.pathname)
    ) {
      return null
    }
    if (isAuthLoopPath(parsed.pathname)) return null
    return `${parsed.pathname}${parsed.search}`
  } catch {
    return null
  }
}

/** Returns `{ redirect }` for login/signup links when the target is a valid non-home internal path. */
export function buildAuthQuery(raw: unknown): { redirect: string } | undefined {
  const safe = sanitizeRedirect(raw)
  return safe && safe !== "/" ? { redirect: safe } : undefined
}
