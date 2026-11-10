import { readFileSync } from "node:fs"
import { join } from "node:path"
import { describe, expect, it } from "vitest"
import { scanDir, scanSource } from "./helpers/hardcodeScan"

/**
 * Zero-hardcode guard. Every directory listed here must stay free of color literals,
 * raw z-index numbers above the local 0-10 range, hard-coded locales and untranslated template text.
 * Add a redesigned area here; never remove one. Exceptions: end the line with
 * `// hardcode-ok: <reason>` (or put `<!-- hardcode-ok -->` on it), or place the marker on the line above.
 */
export const ZERO_DIRS = [
  // Whole source tree (2026-11-10): every top-level area is at zero, so new files are covered too.
  "components",
  "views",
  "layouts",
  "composables",
  "stores",
  "api",
  "router",
  "views/library",
  "views/achievement",
  "components/achievement",
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
  "components/lists",
  "components/feed",
  "components/notifications",
  "components/actor",
  "views/actor",
  "components/auth",
  "components/donate",
  "components/admin",
  "views/admin",
  "views/auth",
  "views/settings",
  "components/movie",
  "components/social",
  "components/common",
  "components/ui",
  "components/profile",
  "components/share",
  "utils",
]

/** Single files (not whole directories) that must stay at zero as well. */
export const ZERO_FILES = [
  "views/FeedView.vue",
  "views/NotificationsView.vue",
  "stores/feed.ts",
  "stores/notification.ts",
  "types/notification.ts",
  "utils/feedActivity.ts",
  "utils/relativeTime.ts",
  "composables/useNotificationList.ts",
  "composables/useSearchResults.ts",
  "views/SearchResultView.vue",
  "views/auth/LoginView.vue",
  "views/auth/RegisterView.vue",
  "views/aboutus/DonateView.vue",
  "layouts/AdminLayout.vue",
  "stores/admin.ts",
  "utils/adminError.ts",
  "components/profile/UserReviews.vue",
  "components/profile/UserAchievements.vue",
  "components/share/AchievementShareSheet.vue",
  "views/user/UserProfileView.vue",
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

  it("flags user-facing copy written in script code", () => {
    const kinds = (src: string, vue = false) => scanSource(src, vue).map((v) => v.split(" ")[0])
    expect(kinds(`toast.success('Signed out')`)).toEqual(["script-text"])
    expect(kinds(`ui.addToast("Saved", "success")`)).toEqual(["script-text"])
    expect(kinds(`const msg = "บันทึกแล้ว"`)).toEqual(["script-text"])
    expect(kinds(`<script setup lang="ts">\nconst e = "โหลดไม่สำเร็จ"\n</script>`, true)).toEqual(["script-text"])
    // i18n calls, logs, comments and dev-only errors are fine
    expect(kinds(`toast.success(t("auth.signedOut"))`)).toEqual([])
    expect(kinds(`// ภาษาไทยในคอมเมนต์\nconsole.error("load failed", err)`)).toEqual([])
    expect(kinds(`throw new Error("useX requires a provider")`)).toEqual([])
    expect(kinds(`// hardcode-ok: endonym\nconst label = "ไทย"`)).toEqual([])
  })
})

describe.each(ZERO_DIRS)("no hardcode in %s", (dir) => {
  it("has no violations", () => {
    expect(scanDir(dir)).toEqual([])
  })
})

describe.each(ZERO_FILES)("no hardcode in %s", (file) => {
  it("has no violations", () => {
    const source = readFileSync(join(import.meta.dirname, "..", file), "utf8")
    expect(scanSource(source, file.endsWith(".vue"))).toEqual([])
  })
})
