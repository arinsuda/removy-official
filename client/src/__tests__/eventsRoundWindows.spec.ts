import { describe, expect, it } from "vitest"
import { canStartLate, currentMonthlyPeriod, currentRoundPeriod, currentYearlyPeriod } from "@/utils/events"

/** Instants below are Thailand (UTC+7) wall-clock times written in UTC. */
const at = (iso: string) => new Date(iso)

describe("current voting round windows (Thailand time)", () => {
  it("keeps the previous monthly round until the 15th, then rolls over", () => {
    expect(currentMonthlyPeriod(at("2026-11-14T16:59:00Z"))).toBe("2026-10") // Nov 14 23:59 BKK
    expect(currentMonthlyPeriod(at("2026-11-14T17:00:00Z"))).toBe("2026-11") // Nov 15 00:00 BKK
    expect(currentMonthlyPeriod(at("2027-01-10T05:00:00Z"))).toBe("2026-12")
  })

  it("is open for the yearly round only from 16 January to 15 February", () => {
    expect(currentYearlyPeriod(at("2027-01-15T12:00:00Z"))).toBeNull()
    expect(currentYearlyPeriod(at("2027-01-15T17:00:00Z"))).toBe("2026") // Jan 16 00:00 BKK
    expect(currentYearlyPeriod(at("2027-02-14T12:00:00Z"))).toBe("2026")
    expect(currentYearlyPeriod(at("2027-02-14T17:00:00Z"))).toBeNull() // Feb 15 00:00 BKK, closed
  })

  it("maps a round kind to its current period", () => {
    expect(currentRoundPeriod("monthly", at("2026-11-15T12:00:00Z"))).toBe("2026-11")
    expect(currentRoundPeriod("yearly", at("2026-11-15T12:00:00Z"))).toBeNull()
  })
})

describe("canStartLate", () => {
  it("allows a late start while at least 72 hours of the official window remain after the lead", () => {
    // Dec 11 13:00 BKK: start at 14:00, round closes Dec 15 00:00 (82 hours later).
    expect(canStartLate("monthly", at("2026-12-11T06:00:00Z"))).toBe(true)
  })

  it("refuses a late start with less than 72 hours left", () => {
    // Dec 12 13:00 BKK: start at 14:00, round closes Dec 15 00:00 (58 hours later).
    expect(canStartLate("monthly", at("2026-12-12T06:00:00Z"))).toBe(false)
  })

  it("never offers a yearly late start outside the yearly window", () => {
    expect(canStartLate("yearly", at("2027-01-15T12:00:00Z"))).toBe(false)
    expect(canStartLate("yearly", at("2027-02-14T12:00:00Z"))).toBe(false) // one day left
  })

  it("offers a yearly late start early in the yearly window", () => {
    expect(canStartLate("yearly", at("2027-01-17T05:00:00Z"))).toBe(true)
  })
})
