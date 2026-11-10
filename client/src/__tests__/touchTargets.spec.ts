import { readFileSync } from "node:fs"
import { resolve } from "node:path"
import { describe, expect, it } from "vitest"

const read = (p: string) => readFileSync(resolve(__dirname, "..", p), "utf8")

describe("touch layout contracts", () => {
  const bmol = read("components/bmol/BmolCard.vue")
  it("BMOL: touch-action none only on the handle, 44px handle on coarse pointers", () => {
    const rules = [...bmol.matchAll(/([^{}]+)\{[^}]*touch-action: none[^}]*\}/g)].map((m) => m[1]!.trim())
    expect(rules).toEqual([".bcard__handle"])
    expect(bmol).toMatch(/@media \(pointer: coarse\)[\s\S]*\.bcard__handle[\s\S]*width: 2\.75rem/)
  })
  it("BMOL: move up/down stay reachable from the card menu", () => {
    expect(bmol).toContain("act('move-up')")
    expect(bmol).toContain("act('move-down')")
  })
  it("PopupCard and the hover popup carry no fixed widths or fixed 5-column action row", () => {
    expect(read("components/movie/PopupCard.vue")).not.toMatch(/repeat\(5,/)
    expect(read("components/home/HomePosterCard.vue")).not.toMatch(/^\s*width: 320px/m)
  })
})
