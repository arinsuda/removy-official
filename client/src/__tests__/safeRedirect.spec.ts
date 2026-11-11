/**
 * @vitest-environment happy-dom
 */
import { describe, expect, it } from "vitest"
import { buildAuthQuery, sanitizeRedirect } from "@/utils/safeRedirect"

describe("safeRedirect", () => {
  it("keeps valid relative paths including query strings", () => {
    expect(sanitizeRedirect("/")).toBe("/")
    expect(sanitizeRedirect("/feed")).toBe("/feed")
    expect(sanitizeRedirect("  /movies/603?tab=reviews#top  ")).toBe("/movies/603?tab=reviews")
    expect(sanitizeRedirect(["/upcoming?genre=28", "/ignored"])).toBe("/upcoming?genre=28")
  })

  it("drops the hash so fragment secrets are never carried", () => {
    expect(sanitizeRedirect("/x#access_token=a")).toBe("/x")
  })

  it("rejects empty, non-string, missing-leading-slash, and external URLs", () => {
    expect(sanitizeRedirect(undefined)).toBeNull()
    expect(sanitizeRedirect(null)).toBeNull()
    expect(sanitizeRedirect("")).toBeNull()
    expect(sanitizeRedirect("   ")).toBeNull()
    expect(sanitizeRedirect(123)).toBeNull()
    expect(sanitizeRedirect("movies")).toBeNull()
    expect(sanitizeRedirect("https://evil.com/phish")).toBeNull()
    expect(sanitizeRedirect("http://localhost:3000/movies/603")).toBeNull()
    expect(sanitizeRedirect("javascript:alert(1)")).toBeNull()
    expect(sanitizeRedirect("data:text/html,hi")).toBeNull()
  })

  it("rejects protocol-relative, backslash, encoded-slash, and control-char payloads", () => {
    expect(sanitizeRedirect("//evil.com")).toBeNull()
    expect(sanitizeRedirect("///evil.com")).toBeNull()
    expect(sanitizeRedirect("/\\evil.com")).toBeNull()
    expect(sanitizeRedirect("/movies\\..\\admin")).toBeNull()
    expect(sanitizeRedirect("/%2fevil.com")).toBeNull()
    expect(sanitizeRedirect("/%2F%2Fevil.com")).toBeNull()
    expect(sanitizeRedirect("/%5cevil.com")).toBeNull()
    expect(sanitizeRedirect("/movies\n603")).toBeNull()
    expect(sanitizeRedirect("/movies\r603")).toBeNull()
    expect(sanitizeRedirect("/movies\u0000603")).toBeNull()
    expect(sanitizeRedirect("/movies\u007f603")).toBeNull()
  })

  it("rejects auth entry routes to prevent redirect loops", () => {
    expect(sanitizeRedirect("/login")).toBeNull()
    expect(sanitizeRedirect("/login?redirect=/feed")).toBeNull()
    expect(sanitizeRedirect("/login/callback")).toBeNull()
    expect(sanitizeRedirect("/register")).toBeNull()
    expect(sanitizeRedirect("/forgot-password")).toBeNull()
    expect(sanitizeRedirect("/reset-password")).toBeNull()
    expect(sanitizeRedirect("/check-email")).toBeNull()
  })

  it("buildAuthQuery omits root and unsafe targets, and keeps safe internal paths", () => {
    expect(buildAuthQuery("/")).toBeUndefined()
    expect(buildAuthQuery("//evil.com")).toBeUndefined()
    expect(buildAuthQuery("/login")).toBeUndefined()
    expect(buildAuthQuery("/movies/603?action=review#reviews")).toEqual({
      redirect: "/movies/603?action=review",
    })
  })
})
