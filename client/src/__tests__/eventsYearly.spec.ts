import { describe, expect, it } from "vitest"
import type { AdminEvent, AdminEventBody } from "@/types/event"
import {
  type EventTranslate,
  eventErrorKey,
  eventTitle,
  isValidYearlyPeriod,
  nextYearlyPeriod,
  roundTitleFromParams,
  validateEventForm,
} from "@/utils/events"
import { expectKeysInBothLocales } from "./helpers/realI18n"

// Fake translator: looks the key up per locale and fills {named} placeholders.
const messages: Record<string, Record<string, string>> = {
  en: { "events.generated.title.vote_yearly.movie": "Popular Movie of the Year - {year}" },
  th: { "events.generated.title.vote_yearly.movie": "ภาพยนตร์ยอดนิยมประจำปี - {year}" },
}
const t: EventTranslate = (key, named, { locale }) =>
  (messages[locale]?.[key] ?? key).replace(/\{(\w+)\}/g, (_, name: string) => named[name] ?? "")

const yearly = (overrides: Partial<AdminEvent> = {}): AdminEvent =>
  ({
    id: 2,
    type: "vote_yearly",
    media_type: "movie",
    period: "2026",
    title: { en: "", th: "" },
    body: { en: "", th: "" },
    image_url: null,
    link_url: null,
    starts_at: "2027-01-15T17:00:00.000Z",
    ends_at: "2027-02-14T17:00:00.000Z",
    show_popup: false,
    popup_priority: 0,
    state: "upcoming",
    status: "draft",
    auto_generated: true,
    created_at: "2026-12-01T00:00:00.000Z",
    updated_at: "2026-12-01T00:00:00.000Z",
    ...overrides,
  }) as AdminEvent

describe("yearly rounds", () => {
  describe("nextYearlyPeriod", () => {
    it("keeps the previous year until the window of that round opens on 16 January (Thailand time)", () => {
      // 15 Jan 2026 19:00 BKK: the 2025 round has not opened yet
      expect(nextYearlyPeriod(new Date("2026-01-15T12:00:00Z"))).toBe("2025")
      // 16 Jan 2026 00:00 BKK = 15 Jan 17:00 UTC: the 2025 round opens, so 2026 is next
      expect(nextYearlyPeriod(new Date("2026-01-15T17:00:00Z"))).toBe("2026")
    })

    it("uses the Bangkok calendar, not UTC, at the year boundary", () => {
      // 31 Dec 2026 23:30 UTC is already 1 Jan 2027 in Bangkok, which is still before 16 Jan
      expect(nextYearlyPeriod(new Date("2026-12-31T23:30:00Z"))).toBe("2026")
      // 31 Dec 2026 12:00 UTC: 2026 is the year, so 2026 round is next
      expect(nextYearlyPeriod(new Date("2026-12-31T12:00:00Z"))).toBe("2026")
    })

    it("names the current year once the round has opened", () => {
      expect(nextYearlyPeriod(new Date("2026-11-15T12:00:00Z"))).toBe("2026")
    })
  })

  describe("isValidYearlyPeriod", () => {
    it("accepts four digits only", () => {
      expect(isValidYearlyPeriod("2026")).toBe(true)
      expect(isValidYearlyPeriod("2026-11")).toBe(false)
      expect(isValidYearlyPeriod("26")).toBe(false)
      expect(isValidYearlyPeriod("")).toBe(false)
    })
  })

  describe("generated copy", () => {
    it("fills a blank yearly title from the year", () => {
      expect(eventTitle(yearly(), t, "en")).toBe("Popular Movie of the Year - 2026")
    })

    it("keeps admin text when the admin wrote a title", () => {
      expect(eventTitle(yearly({ title: { en: "Custom", th: "" } }), t, "en")).toBe("Custom")
    })

    it("has the generated copy keys in both locales", () => {
      expectKeysInBothLocales([
        "events.generated.title.vote_yearly.movie",
        "events.generated.title.vote_yearly.tv",
        "events.generated.body.vote_yearly.movie",
        "events.generated.body.vote_yearly.tv",
        "awards.badge.labelYearly",
      ])
    })
  })

  describe("validateEventForm", () => {
    const base: AdminEventBody = {
      type: "vote_yearly",
      media_type: "movie",
      period: "2026",
      title_en: "",
      title_th: "",
      body_en: "",
      body_th: "",
      link_url: "",
      starts_at: "2027-01-15T17:00:00.000Z",
      ends_at: "2027-02-14T17:00:00.000Z",
      show_popup: false,
      popup_priority: 0,
      status: "draft",
    }

    it("still requires titles for an announcement", () => {
      const errors = validateEventForm({ ...base, type: "announcement" })
      expect(errors.title_en).toBe("admin.events.errors.titleRequired")
      expect(errors.title_th).toBe("admin.events.errors.titleRequired")
    })
  })

  describe("eventErrorKey", () => {
    it("maps MONTHLY_ROUNDS_PENDING to its own message", () => {
      expect(
        eventErrorKey({ response: { status: 409, data: { code: "MONTHLY_ROUNDS_PENDING" } } }),
      ).toBe("admin.events.errors.monthlyRoundsPending")
    })
  })

  describe("roundTitleFromParams", () => {
    it("titles a yearly event notification from its params", () => {
      expect(roundTitleFromParams({ event_type: "vote_yearly", media_type: "movie", period: "2026" }, t, "en")).toBe(
        "Popular Movie of the Year - 2026",
      )
    })

    it("returns null for a malformed or incomplete payload", () => {
      expect(roundTitleFromParams({ event_type: "vote_yearly", media_type: "movie", period: "26" }, t, "en")).toBeNull()
      expect(roundTitleFromParams({ event_type: "announcement", media_type: "movie", period: "2026" }, t, "en")).toBeNull()
      expect(roundTitleFromParams({ event_type: "vote_yearly", media_type: "book", period: "2026" }, t, "en")).toBeNull()
      expect(roundTitleFromParams({ event_type: "vote_yearly", media_type: "movie" }, t, "en")).toBeNull()
    })
  })

  describe("eventErrorKey for yearly generation", () => {
    it("maps TOO_FEW_YEARLY_NOMINEES to its own message", () => {
      expect(
        eventErrorKey({ response: { status: 409, data: { code: "TOO_FEW_YEARLY_NOMINEES" } } }),
      ).toBe("admin.events.errors.tooFewYearlyNominees")
    })
  })
})
