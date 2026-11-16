import { describe, expect, it } from "vitest"
import { isWebUrl, knownNumber, plainDescription, statusKey } from "@/components/manga/mangaFormat"

describe("plainDescription", () => {
  it("turns line breaks into newlines and strips other tags", () => {
    expect(plainDescription("First<br><br>Second <i>word</i><br/>Third")).toBe("First\n\nSecond word\nThird")
  })

  it("collapses long runs of breaks", () => {
    expect(plainDescription("A<br><br><br><br>B")).toBe("A\n\nB")
  })
})

describe("knownNumber", () => {
  it("treats null and 0 as unknown", () => {
    expect(knownNumber(null)).toBeNull()
    expect(knownNumber(0)).toBeNull()
    expect(knownNumber(12)).toBe(12)
  })
})

describe("statusKey", () => {
  it("maps known statuses regardless of case and everything else to other", () => {
    expect(statusKey("RELEASING")).toBe("releasing")
    expect(statusKey("finished")).toBe("finished")
    expect(statusKey("HIATUS")).toBe("other")
  })
})

describe("isWebUrl", () => {
  it("accepts only http and https", () => {
    expect(isWebUrl("https://example.com/read")).toBe(true)
    expect(isWebUrl("HTTP://example.com")).toBe(true)
    expect(isWebUrl("javascript:alert(1)")).toBe(false)
    expect(isWebUrl("/relative")).toBe(false)
  })
})
