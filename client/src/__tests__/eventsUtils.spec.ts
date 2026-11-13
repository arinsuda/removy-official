import { describe, expect, it } from "vitest"
import type { AdminEventBody, LocalizedText } from "@/types/event"
import {
  dismissPopup,
  EVENT_BODY_MAX,
  EVENT_TITLE_MAX,
  eventErrorKey,
  isInternalLink,
  isPopupHidden,
  isValidLink,
  isoToLocalInput,
  localDateKey,
  localizedText,
  localInputToIso,
  validateEventForm,
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
  })
})
