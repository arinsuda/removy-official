import { describe, expect, it } from "vitest"
import type { AdminEvent, AdminEventBody, LocalizedText } from "@/types/event"
import {
  dismissPopup,
  EVENT_BODY_MAX,
  EVENT_TITLE_MAX,
  type EventTranslate,
  eventBody,
  eventErrorKey,
  eventTitle,
  nextMonthlyPeriod,
  publishErrorKey,
  isInternalLink,
  isPopupHidden,
  isValidLink,
  isoToLocalInput,
  localDateKey,
  localizedText,
  localInputToIso,
  validateEventForm,
  voteErrorKey,
} from "@/utils/events"

describe("events utils", () => {
  describe("localizedText", () => {
    const text: LocalizedText = { en: "English Title", th: "ชื่อภาษาไทย" }

    it("returns Thai text when locale is th", () => {
      expect(localizedText(text, "th")).toBe("ชื่อภาษาไทย")
    })

    it("returns English text when locale is en", () => {
      expect(localizedText(text, "en")).toBe("English Title")
    })

    it("falls back to the other language when primary is empty", () => {
      expect(localizedText({ en: "Only EN", th: "  " }, "th")).toBe("Only EN")
      expect(localizedText({ en: "", th: "Only TH" }, "en")).toBe("Only TH")
    })

    it("handles null / undefined safely", () => {
      expect(localizedText(null, "en")).toBe("")
      expect(localizedText(undefined, "th")).toBe("")
    })
  })

  describe("localDateKey", () => {
    it("formats a Date to YYYY-MM-DD", () => {
      const d = new Date(2026, 10, 15, 10, 30) // November (month index 10)
      expect(localDateKey(d)).toBe("2026-11-15")
    })
  })

  describe("popup storage dismissal", () => {
    it("is not hidden initially", () => {
      localStorage.clear()
      sessionStorage.clear()
      expect(isPopupHidden(42)).toBe(false)
    })

    it("hides for the session when dismissed without checkbox", () => {
      localStorage.clear()
      sessionStorage.clear()
      dismissPopup(42, false)
      expect(isPopupHidden(42)).toBe(true)
      // Different user is not affected
      expect(isPopupHidden(43)).toBe(false)
    })

    it("hides for today when dismissed with checkbox", () => {
      localStorage.clear()
      sessionStorage.clear()
      const today = new Date(2026, 10, 15)
      dismissPopup(42, true, today)
      expect(isPopupHidden(42, today)).toBe(true)

      // Next day it is visible again if sessionStorage is cleared (new session)
      sessionStorage.clear()
      const tomorrow = new Date(2026, 10, 16)
      expect(isPopupHidden(42, tomorrow)).toBe(false)
    })
  })

  describe("links validation", () => {
    it("recognizes internal paths", () => {
      expect(isInternalLink("/events")).toBe(true)
      expect(isInternalLink("/movies/123")).toBe(true)
      expect(isInternalLink("//evil.com")).toBe(false)
      expect(isInternalLink("https://example.com")).toBe(false)
      expect(isInternalLink(null)).toBe(false)
    })

    it("validates external and internal URLs", () => {
      expect(isValidLink("")).toBe(true)
      expect(isValidLink("/events")).toBe(true)
      expect(isValidLink("https://example.com")).toBe(true)
      expect(isValidLink("http://example.com/page")).toBe(true)
      expect(isValidLink("//evil.com")).toBe(false)
      expect(isValidLink("javascript:alert(1)")).toBe(false)
      expect(isValidLink("ftp://example.com")).toBe(false)
    })
  })

  describe("datetime input conversions", () => {
    it("converts ISO to local input format and back", () => {
      const d = new Date(2026, 10, 15, 14, 30)
      const inputStr = isoToLocalInput(d.toISOString())
      expect(inputStr).toContain("2026-11-15T14:30")
      const roundtripIso = localInputToIso(inputStr)
      expect(new Date(roundtripIso).getTime()).toBe(d.getTime())
    })

    it("handles empty or invalid inputs", () => {
      expect(isoToLocalInput("")).toBe("")
      expect(isoToLocalInput("invalid")).toBe("")
      expect(localInputToIso("")).toBe("")
    })
  })

  describe("validateEventForm", () => {
    const validForm: AdminEventBody = {
      title_en: "Summer Festival",
      title_th: "เทศกาลฤดูร้อน",
      body_en: "Description EN",
      body_th: "คำอธิบาย TH",
      link_url: "/events",
      starts_at: "2026-11-15T10:00:00.000Z",
      ends_at: "2026-11-20T10:00:00.000Z",
      show_popup: true,
      popup_priority: 5,
      status: "published",
    }

    it("returns no errors for valid body", () => {
      const errs = validateEventForm(validForm)
      expect(Object.keys(errs)).toHaveLength(0)
    })

    it("validates required title_en and title_th", () => {
      expect(validateEventForm({ ...validForm, title_en: "   " }).title_en).toBe(
        "admin.events.errors.titleRequired",
      )
      expect(validateEventForm({ ...validForm, title_th: "" }).title_th).toBe(
        "admin.events.errors.titleRequired",
      )
    })

    it("validates title max runes", () => {
      const longTitle = "A".repeat(EVENT_TITLE_MAX + 1)
      expect(validateEventForm({ ...validForm, title_en: longTitle }).title_en).toBe(
        "admin.events.errors.titleTooLong",
      )
    })

    it("validates body max runes", () => {
      const longBody = "B".repeat(EVENT_BODY_MAX + 1)
      expect(validateEventForm({ ...validForm, body_en: longBody }).body_en).toBe(
        "admin.events.errors.bodyTooLong",
      )
    })

    it("validates date order", () => {
      expect(
        validateEventForm({
          ...validForm,
          starts_at: "2026-11-20T10:00:00.000Z",
          ends_at: "2026-11-15T10:00:00.000Z",
        }).ends_at,
      ).toBe("admin.events.errors.endBeforeStart")
    })
  })

  describe("eventErrorKey", () => {
    it("maps 413 to imageTooLarge", () => {
      expect(eventErrorKey({ response: { status: 413 } })).toBe("admin.events.errors.imageTooLarge")
    })

    it("maps 415 or UNSUPPORTED_MEDIA_TYPE to imageType", () => {
      expect(eventErrorKey({ response: { status: 415 } })).toBe("admin.events.errors.imageType")
      expect(
        eventErrorKey({ response: { status: 400, data: { code: "UNSUPPORTED_MEDIA_TYPE" } } }),
      ).toBe("admin.events.errors.imageType")
    })

    it("maps INVALID_INPUT", () => {
      expect(
        eventErrorKey({ response: { status: 400, data: { code: "INVALID_INPUT" } } }),
      ).toBe("admin.events.errors.invalidInput")
    })

    it("maps NOT_FOUND", () => {
      expect(eventErrorKey({ response: { status: 404 } })).toBe("admin.events.errors.notFound")
    })

    it("maps voting and generation error codes", () => {
      expect(
        eventErrorKey({ response: { status: 409, data: { code: "EVENT_ALREADY_EXISTS" } } }),
      ).toBe("admin.events.errors.eventExists")
      expect(
        eventErrorKey({ response: { status: 409, data: { code: "VOTING_LOCKED" } } }),
      ).toBe("admin.events.errors.votingLocked")
      expect(
        eventErrorKey({ response: { status: 409, data: { code: "EVENT_NOT_PUBLISHED" } } }),
      ).toBe("admin.events.errors.notPublished")
      expect(
        eventErrorKey({ response: { status: 409, data: { code: "VOTING_NOT_ENDED" } } }),
      ).toBe("admin.events.errors.votingNotEnded")
      expect(
        eventErrorKey({ response: { status: 422, data: { code: "NOT_ENOUGH_NOMINEES" } } }),
      ).toBe("admin.events.errors.notEnoughNominees")
    })
  })

  describe("publishErrorKey", () => {
    it("maps NOT_ENOUGH_NOMINEES on a publish to the publish-specific message", () => {
      expect(
        publishErrorKey({ response: { status: 422, data: { code: "NOT_ENOUGH_NOMINEES" } } }),
      ).toBe("admin.events.errors.publishNeedsNominees")
    })

    it("defers to eventErrorKey for every other code", () => {
      expect(
        publishErrorKey({ response: { status: 409, data: { code: "VOTING_LOCKED" } } }),
      ).toBe("admin.events.errors.votingLocked")
      expect(publishErrorKey(new Error("network"))).toBe("admin.errors.generic")
    })
  })

  describe("voteErrorKey", () => {
    it("maps voting error codes to i18n keys", () => {
      expect(voteErrorKey({ response: { data: { code: "VOTING_CLOSED" } } })).toBe(
        "events.voting.errors.closed",
      )
      expect(voteErrorKey({ response: { data: { code: "VOTING_NOT_STARTED" } } })).toBe(
        "events.voting.errors.notStarted",
      )
      expect(voteErrorKey({ response: { data: { code: "NOMINEE_NOT_FOUND" } } })).toBe(
        "events.voting.errors.nomineeNotFound",
      )
      expect(voteErrorKey({})).toBe("common.error")
    })
  })

  describe("nextMonthlyPeriod", () => {
    it("picks the round after the active one, rolling over on the 15th in Thailand time", () => {
      // May 14, 2026 19:00 BKK -> active 2026-05 -> next 2026-06
      expect(nextMonthlyPeriod(new Date("2026-05-14T12:00:00Z"))).toBe("2026-06")
      // May 14, 2026 23:59 BKK -> still 2026-05 active -> next 2026-06
      expect(nextMonthlyPeriod(new Date("2026-05-14T16:59:00Z"))).toBe("2026-06")
      // May 15, 2026 00:00 BKK -> active 2026-06 -> next 2026-07
      expect(nextMonthlyPeriod(new Date("2026-05-14T17:00:00Z"))).toBe("2026-07")
      // May 15, 2026 01:00 BKK -> active 2026-06 -> next 2026-07
      expect(nextMonthlyPeriod(new Date("2026-05-14T18:00:00Z"))).toBe("2026-07")
    })

    it("rolls the year over", () => {
      // Dec 14, 2026 BKK -> active 2026-12 -> next 2027-01
      expect(nextMonthlyPeriod(new Date("2026-12-14T12:00:00Z"))).toBe("2027-01")
      // Dec 15, 2026 BKK -> active 2027-01 -> next 2027-02
      expect(nextMonthlyPeriod(new Date("2026-12-15T12:00:00Z"))).toBe("2027-02")
    })
  })

  describe("eventTitle and eventBody", () => {
    const messages: Record<string, Record<string, string>> = {
      en: {
        "events.generated.title.vote_monthly.movie": "Popular Movie of the Month - {month}",
        "events.generated.body.vote_monthly.movie":
          "Vote for your favorite movie of {month}! Voting closes {closes} {tz}.",
        "events.generated.timezone": "(Thailand time)",
      },
      th: {
        "events.generated.title.vote_monthly.movie": "ภาพยนตร์ยอดนิยมประจำเดือน - {month}",
        "events.generated.body.vote_monthly.movie":
          "ร่วมโหวตภาพยนตร์ยอดนิยมประจำเดือน{month}! ปิดโหวต {closes} {tz}",
        "events.generated.timezone": "(เวลาประเทศไทย)",
      },
    }
    // Fake translator: looks the key up per locale and fills {named} placeholders.
    const t: EventTranslate = (key, named, { locale }) =>
      (messages[locale]?.[key] ?? key).replace(/\{(\w+)\}/g, (_, name: string) => named[name] ?? "")

    const monthly = (overrides: Partial<AdminEvent> = {}): AdminEvent =>
      ({
        id: 1,
        type: "vote_monthly",
        media_type: "movie",
        period: "2026-11",
        title: { en: "", th: "" },
        body: { en: "", th: "" },
        image_url: null,
        link_url: null,
        starts_at: "2026-10-15T00:00:00Z",
        ends_at: "2026-11-15T00:00:00Z",
        show_popup: false,
        popup_priority: 0,
        state: "active",
        status: "published",
        created_at: "",
        updated_at: "",
        ...overrides,
      }) as AdminEvent

    it("uses generated monthly copy in English when both admin titles are blank", () => {
      expect(eventTitle(monthly(), t, "en")).toBe("Popular Movie of the Month - November 2026")
    })

    it("uses generated monthly copy in Thai, with the Buddhist-era year", () => {
      const title = eventTitle(monthly(), t, "th")
      expect(title).toContain("ภาพยนตร์ยอดนิยมประจำเดือน - ")
      expect(title).toContain("พฤศจิกายน")
      expect(title).toContain("2569")
    })

    it("fills the body with the closing time and timezone label", () => {
      const body = eventBody(monthly(), t, "en")
      expect(body).toMatch(/^Vote for your favorite movie of November 2026! Voting closes .+ \(Thailand time\)\.$/)
    })

    it("lets admin text win over generated copy", () => {
      const ev = monthly({ title: { en: "Custom title", th: "" } })
      expect(eventTitle(ev, t, "en")).toBe("Custom title")
      expect(eventTitle(ev, t, "th")).toBe("Custom title")
    })

    it("falls back to the other language when the admin text is missing in this locale", () => {
      const ev = monthly({ title: { en: "", th: "ชื่อภาษาไทย" } })
      expect(eventTitle(ev, t, "en")).toBe("ชื่อภาษาไทย")
    })

    it("returns empty text for a blank announcement", () => {
      const ev = monthly({ type: "announcement", media_type: null, period: null })
      expect(eventTitle(ev, t, "en")).toBe("")
      expect(eventBody(ev, t, "th")).toBe("")
    })
  })

  describe("validateEventForm for voting rounds", () => {
    const votingForm: AdminEventBody = {
      type: "vote_monthly",
      media_type: "movie",
      period: "2026-11",
      title_en: "",
      title_th: "",
      body_en: "",
      body_th: "",
      link_url: "",
      starts_at: "2026-10-15T00:00:00Z",
      ends_at: "2026-11-15T00:00:00Z",
      show_popup: false,
      popup_priority: 0,
      status: "draft",
    }

    it("accepts blank titles for voting rounds", () => {
      expect(validateEventForm(votingForm)).toEqual({})
    })

    it("still requires titles for announcements", () => {
      expect(validateEventForm({ ...votingForm, type: "announcement", period: null }).title_en).toBe(
        "admin.events.errors.titleRequired",
      )
    })

    it("still requires titles for yearly rounds", () => {
      const errors = validateEventForm({ ...votingForm, type: "vote_yearly", period: "2026" })
      expect(errors.title_en).toBe("admin.events.errors.titleRequired")
      expect(errors.title_th).toBe("admin.events.errors.titleRequired")
    })
  })
})
