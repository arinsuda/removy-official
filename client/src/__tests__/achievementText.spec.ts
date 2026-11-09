import { readFileSync } from "node:fs"
import { resolve } from "node:path"
import { describe, expect, it } from "vitest"
import { REQUIREMENT_KEY_PREFIX } from "@/components/achievement/text"
import { useRealI18n } from "./helpers/realI18n"

const SEEDER = resolve(__dirname, "../../../server/database/seeder/achievement.json")
const seeded = JSON.parse(readFileSync(SEEDER, "utf-8")) as { action_type: string; target_count: number }[]
const i18n = useRealI18n()

describe("achievement requirement copy", () => {
  it("has en and th text for every seeded action type", () => {
    const missing = [...new Set(seeded.map((a) => a.action_type))].flatMap((type) =>
      (["en", "th"] as const).filter((l) => !i18n.global.te(`${REQUIREMENT_KEY_PREFIX}${type}`, l)).map((l) => `${l}:${type}`),
    )
    expect(missing).toEqual([])
  })

  it("interpolates the target and pluralizes the first step in English", () => {
    const { t } = i18n.global
    i18n.global.locale.value = "en"
    expect(t(`${REQUIREMENT_KEY_PREFIX}watched_count`, { target: "1" }, 1)).toBe("Watch your first movie or series")
    expect(t(`${REQUIREMENT_KEY_PREFIX}watched_count`, { target: "10" }, 10)).toBe("Watch 10 movies or series")
    i18n.global.locale.value = "th"
    expect(t(`${REQUIREMENT_KEY_PREFIX}watched_count`, { target: "10" }, 10)).toContain("10")
  })
})
