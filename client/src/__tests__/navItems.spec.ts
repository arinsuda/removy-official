import { describe, expect, it } from "vitest"
import type { RouteLocationNormalizedLoaded } from "vue-router"
import { NAV_ITEMS } from "@/components/layout/navItems"

const route = (path: string) => ({ path, name: undefined }) as unknown as RouteLocationNormalizedLoaded

describe("NAV_ITEMS anime entry", () => {
  const anime = NAV_ITEMS.find((item) => item.key === "anime")

  it("links to /anime with the navigation label", () => {
    expect(anime).toBeDefined()
    expect(anime?.to).toBe("/anime")
    expect(anime?.labelKey).toBe("navigation.anime")
  })

  it("is active on the anime page and not on the other browse pages", () => {
    expect(anime?.isActive(route("/anime"))).toBe(true)
    expect(anime?.isActive(route("/tv"))).toBe(false)
    expect(anime?.isActive(route("/movies"))).toBe(false)
  })
})

describe("NAV_ITEMS manga entry", () => {
  const manga = NAV_ITEMS.find((item) => item.key === "manga")

  it("links to /manga after anime with the navigation label", () => {
    expect(manga).toBeDefined()
    expect(manga?.to).toBe("/manga")
    expect(manga?.labelKey).toBe("navigation.manga")
    expect(NAV_ITEMS.findIndex((item) => item.key === "manga")).toBe(NAV_ITEMS.findIndex((item) => item.key === "anime") + 1)
  })

  it("is active on the manga pages and not on the anime page", () => {
    expect(manga?.isActive(route("/manga"))).toBe(true)
    expect(manga?.isActive(route("/manga/7"))).toBe(true)
    expect(manga?.isActive(route("/anime"))).toBe(false)
  })
})
