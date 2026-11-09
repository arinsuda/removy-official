import { describe, expect, it } from "vitest"
import { scanDir, scanSource } from "./helpers/hardcodeScan"

/**
 * Zero-hardcode guard. Every directory listed here must stay free of color literals,
 * raw z-index numbers above the local 0-10 range, hard-coded locales and untranslated template text.
 * Add a redesigned area here; never remove one. Exceptions: end the line with
 * `// hardcode-ok: <reason>` (or put `<!-- hardcode-ok -->` on it), or place the marker on the line above.
 */
export const ZERO_DIRS = [
  "views/library",
  "components/browse",
  "components/upcoming",
  "components/about",
  "components/search",
  "components/detail",
  "components/bmol",
  "components/profile/tabs",
  "components/settings",
  "components/layout",
  "components/onboarding",
  "components/home",
]

describe("scanner", () => {
  it("flags literals and ignores i18n calls and exceptions", () => {
    const bad = `<template><p title="Hi">Hello {{ t("x") }}</p><span :title="t('a')">{{ n }}</span></template>
<style>.a { color: #fff; z-index: 50 }</style>
<script>const l = "th-TH"</script>`
    expect(scanSource(bad, true).map((v) => v.split(/[ :]/)[0])).toEqual(["color", "z-index", "locale", "attr", "text"])
    expect(scanSource(".a { z-index: 5 }", false)).toEqual([])
    const ok = `<template><p>{{ t("x") }}</p></template>
<style>
/* hardcode-ok: fixed */
.a { color: #fff }
.b { color: var(--color-brand) }
</style>`
    expect(scanSource(ok, true)).toEqual([])
  })
})

describe.each(ZERO_DIRS)("no hardcode in %s", (dir) => {
  it("has no violations", () => {
    expect(scanDir(dir)).toEqual([])
  })
})
