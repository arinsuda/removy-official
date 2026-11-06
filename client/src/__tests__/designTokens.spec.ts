import { readdirSync, readFileSync } from "node:fs"
import { join } from "node:path"
import { describe, expect, it } from "vitest"

// Ratchet: hard-coded colors and legacy --c-* variables may only go down.
// When a count drops, LOWER the baseline below so the gain is locked in.
// Convert colors to design tokens when you edit a file; never add new ones.
const BASELINE = { hex: 1026, rgb: 1052, legacyVars: 101 }

const SRC = join(import.meta.dirname, "..")

function walk(dir: string, exts: string[], out: string[] = []): string[] {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) {
      // Test files are included on purpose; nothing under src is skipped.
      walk(full, exts, out)
    } else if (exts.some((e) => entry.name.endsWith(e))) {
      out.push(full)
    }
  }
  return out
}

function count(files: string[], re: RegExp): number {
  return files.reduce((n, f) => n + (readFileSync(f, "utf8").match(re)?.length ?? 0), 0)
}

describe("design token ratchet", () => {
  const vue = walk(SRC, [".vue"])
  const styled = walk(SRC, [".vue", ".css"])

  it("finds source files", () => {
    expect(vue.length).toBeGreaterThan(50)
  })

  it("does not add hex colors to .vue files", () => {
    expect(count(vue, /#[0-9a-fA-F]{3,8}\b/g)).toBeLessThanOrEqual(BASELINE.hex)
  })

  it("does not add rgb()/rgba() colors to .vue files", () => {
    expect(count(vue, /\brgba?\(/g)).toBeLessThanOrEqual(BASELINE.rgb)
  })

  it("does not add --c-* legacy variable definitions", () => {
    expect(count(styled, /^\s*--c-[a-z0-9-]+\s*:/gm)).toBeLessThanOrEqual(BASELINE.legacyVars)
  })
})
