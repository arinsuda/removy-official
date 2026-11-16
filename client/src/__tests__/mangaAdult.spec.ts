import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { AxiosError, AxiosHeaders } from "axios"
import { MANGA_ADULT_ACK_KEY_PREFIX } from "@/components/manga/mangaConfig"
import { isAdultByBirth, isAgeRestricted, readAdultAck, writeAdultAck } from "@/components/manga/mangaAdult"

// Explicit Bangkok instants keep the spec independent of the runner's time zone.
const TODAY = new Date("2026-11-16T12:00:00+07:00")

describe("isAdultByBirth", () => {
  it("is true from the 18th birthday, not the day before", () => {
    expect(isAdultByBirth("2008-11-16", TODAY)).toBe(true)
    expect(isAdultByBirth("2008-11-17", TODAY)).toBe(false)
    expect(isAdultByBirth("1990-01-01", TODAY)).toBe(true)
  })

  it("unlocks on the Bangkok calendar day, the same day as the server", () => {
    const birthday = "2008-11-16"
    expect(isAdultByBirth(birthday, new Date("2026-11-15T23:59:00+07:00"))).toBe(false)
    expect(isAdultByBirth(birthday, new Date("2026-11-16T00:00:00+07:00"))).toBe(true)
    // 18:00 UTC on 15 November is already 16 November in Bangkok.
    expect(isAdultByBirth(birthday, new Date("2026-11-15T18:00:00Z"))).toBe(true)
  })

  it("moves a 29 February birthday to 1 March in non-leap years", () => {
    expect(isAdultByBirth("2008-02-29", new Date("2026-02-28T23:59:00+07:00"))).toBe(false)
    expect(isAdultByBirth("2008-02-29", new Date("2026-03-01T00:00:00+07:00"))).toBe(true)
    expect(isAdultByBirth("2008-02-29", new Date("2028-02-29T12:00:00+07:00"))).toBe(true)
  })

  it("is false without or with a malformed date of birth", () => {
    expect(isAdultByBirth(null, TODAY)).toBe(false)
    expect(isAdultByBirth(undefined, TODAY)).toBe(false)
    expect(isAdultByBirth("", TODAY)).toBe(false)
    expect(isAdultByBirth("not-a-date", TODAY)).toBe(false)
  })
})

describe("adult acknowledgement storage", () => {
  beforeEach(() => localStorage.clear())
  afterEach(() => vi.unstubAllGlobals())

  it("is remembered per user", () => {
    expect(readAdultAck(7)).toBe(false)
    writeAdultAck(7)
    expect(readAdultAck(7)).toBe(true)
    expect(readAdultAck(8)).toBe(false)
    expect(localStorage.getItem(`${MANGA_ADULT_ACK_KEY_PREFIX}7`)).toBe("1")
  })

  it("reads as not acknowledged for guests", () => {
    expect(readAdultAck(null)).toBe(false)
  })

  it("does not throw when storage is blocked", () => {
    const blocked = {
      getItem: () => {
        throw new Error("blocked")
      },
      setItem: () => {
        throw new Error("blocked")
      },
    }
    vi.stubGlobal("localStorage", blocked)
    expect(() => writeAdultAck(7)).not.toThrow()
    expect(readAdultAck(7)).toBe(false)
    vi.unstubAllGlobals()
  })
})

describe("isAgeRestricted", () => {
  const httpError = (status: number, data: unknown) =>
    new AxiosError("denied", "ERR_BAD_REQUEST", undefined, undefined, {
      status,
      statusText: "",
      headers: {},
      config: { headers: new AxiosHeaders() },
      data,
    })

  it("matches the AGE_RESTRICTED code only", () => {
    expect(isAgeRestricted(httpError(403, { code: "AGE_RESTRICTED", error: "x" }))).toBe(true)
    expect(isAgeRestricted(httpError(403, { code: "XHR_HEADER_REQUIRED" }))).toBe(false)
    expect(isAgeRestricted(new Error("boom"))).toBe(false)
    expect(isAgeRestricted(null)).toBe(false)
  })
})
