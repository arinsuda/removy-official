import { readFileSync } from "node:fs"
import { join } from "node:path"
import { describe, expect, it } from "vitest"
import { BREAKPOINTS, maxWidthQuery, minWidthQuery } from "@/constants/breakpoints"

describe("breakpoints", () => {
  it("matches the --breakpoint-* tokens in main.css", () => {
    const css = readFileSync(join(import.meta.dirname, "..", "assets", "styles", "main.css"), "utf8")
    for (const [name, px] of Object.entries(BREAKPOINTS)) {
      const m = new RegExp(`--breakpoint-${name}:\\s*([\\d.]+)rem`).exec(css)
      expect(m, name).not.toBeNull()
      expect(Number(m![1]) * 16).toBe(px)
    }
  })

  it("builds mobile-first queries", () => {
    expect(minWidthQuery("md")).toBe("(min-width: 768px)")
    expect(maxWidthQuery("md")).toBe("(max-width: 767.98px)")
  })
})
