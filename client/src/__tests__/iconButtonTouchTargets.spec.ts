import { readdirSync, readFileSync, statSync } from "node:fs"
import { join, relative } from "node:path"
import { describe, expect, it } from "vitest"

/**
 * Icon-only <button>s (aria-label or title, no visible text) must be AppButton, carry the
 * `touch-target` utility (44px hit area on coarse pointers), or opt out with `data-touch-ok`
 * when the control is already 44px. Areas owned by redesigns still in flight are not listed yet;
 * add a directory here once it is clean and never remove one.
 */
const TOUCH_DIRS = [
  "components/layout",
  "components/ui",
  "components/notifications",
  "components/social",
  "components/lists",
  "components/bmol",
  "components/share",
  "components/movie",
  "components/common",
  "components/auth",
  "components/settings",
  "components/onboarding",
]

const SRC = join(__dirname, "..")

function vueFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const p = join(dir, name)
    if (statSync(p).isDirectory()) return vueFiles(p)
    return p.endsWith(".vue") ? [p] : []
  })
}

function offenders(): string[] {
  const out: string[] = []
  for (const dir of TOUCH_DIRS) {
    for (const file of vueFiles(join(SRC, dir))) {
      const src = readFileSync(file, "utf8")
      const template = src.split(/<script|<style/)[0] ?? ""
      for (const m of template.matchAll(/<button\b([^>]*)>([\s\S]*?)<\/button>/g)) {
        const attrs = m[1] ?? ""
        const text = (m[2] ?? "").replace(/<[^>]+>/g, "").replace(/\{\{[\s\S]*?\}\}/g, "T").trim()
        if (text || !/aria-label|title/.test(attrs)) continue
        if (/touch-target|data-touch-ok/.test(attrs)) continue
        out.push(`${relative(SRC, file)}:${template.slice(0, m.index).split("\n").length}`)
      }
    }
  }
  return out
}

describe("touch targets", () => {
  it("every icon-only button has a 44px hit area on coarse pointers", () => {
    expect(offenders()).toEqual([])
  })
})
