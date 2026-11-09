import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import AchievementsView from "@/views/achievement/AchievementsView.vue"
import { categoryOf, ctaFor } from "@/components/achievement/config"
import { buildUnits, countByStatus, filterUnits, groupByCategory, latestUnlock, nextUp, sortUnits, summarize } from "@/components/achievement/units"
import type { UserAchievement } from "@/types/achievement"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ list: vi.fn(), profile: vi.fn() }))
vi.mock("@/api/api", () => ({ userApi: { getProfile: m.profile } }))
vi.mock("@/api/endpoints/achievement", () => ({ achievementApi: { listByUser: m.list } }))

let nextId = 1
const ach = (name: string, type: string, target: number, current: number, at: string | null = null): UserAchievement => {
  const id = nextId++
  const unlocked = current >= target
  return { achievement_id: id, achievement: { id, name, description: `${name} description`, exp_reward: 10, action_type: type, target_count: target }, current_count: current, is_unlocked: unlocked, unlocked_at: unlocked ? (at ?? "2025-01-01T00:00:00Z") : null, progress_pct: Math.round((current / target) * 100) }
}
const sample = () => [
  ach("Critic I", "review_count", 1, 1, "2025-03-01T00:00:00Z"),
  ach("Critic II", "review_count", 10, 4),
  ach("Critic III", "review_count", 50, 4),
  ach("Binger", "watched_count", 10, 10, "2025-05-01T00:00:00Z"),
  ach("Socialite", "follower_count", 5, 0),
  ach("Shelf", "watchlist_count", 20, 15),
]

describe("achievement units", () => {
  beforeEach(() => (nextId = 1))

  it("maps action types to categories with a fallback", () => {
    expect(categoryOf("review_count")).toBe("reviews")
    expect(categoryOf("genre_watch_horror")).toBe("genres")
    expect(categoryOf("studio_watch_pixar")).toBe("cinema")
    expect(categoryOf("brand_new_type")).toBe("journey")
  })

  it("builds families from shared action types and picks the current tier", () => {
    const units = buildUnits(sample())
    expect(units).toHaveLength(4)
    const critic = units.find((u) => u.key === "family:review_count")!
    expect(critic.tiers.map((t) => t.achievement.name)).toEqual(["Critic I", "Critic II", "Critic III"])
    expect(critic.current.achievement.name).toBe("Critic II")
    expect(critic.status).toBe("progress")
    expect(units.find((u) => u.key.startsWith("single:") && u.current.achievement.name === "Binger")!.status).toBe("unlocked")
  })

  it("computes summary, latest unlock and next up", () => {
    const list = sample()
    expect(summarize(list)).toMatchObject({ unlocked: 2, total: 6, pct: 33, xp: 20 })
    expect(latestUnlock(list)?.achievement.name).toBe("Binger")
    expect(nextUp(list)?.achievement.name).toBe("Shelf")
    expect(nextUp(list.filter((a) => a.is_unlocked))).toBeNull()
  })

  it("filters by status, category and search, and counts statuses", () => {
    const units = buildUnits(sample())
    expect(countByStatus(units)).toEqual({ all: 4, unlocked: 1, progress: 2, locked: 1 })
    expect(filterUnits(units, { status: "locked", category: "all", search: "" })).toHaveLength(1)
    expect(filterUnits(units, { status: "all", category: "reviews", search: "" })).toHaveLength(1)
    expect(filterUnits(units, { status: "all", category: "all", search: "critic iii" })).toHaveLength(1)
  })

  it("sorts and groups into category sections", () => {
    const units = buildUnits(sample())
    expect(sortUnits(units, "recent", "en")[0]!.current.achievement.name).toBe("Binger")
    expect(sortUnits(units, "progress", "en")[0]!.current.achievement.name).toBe("Shelf")
    expect(sortUnits(units, "name", "en").map((u) => u.current.achievement.name)).toEqual(["Binger", "Critic II", "Shelf", "Socialite"])
    const sections = groupByCategory(units)
    expect(sections.map((s) => s.key)).toEqual(["watching", "reviews", "social", "library"])
    expect(sections.find((s) => s.key === "reviews")).toMatchObject({ unlocked: 1, total: 3 })
  })

  it("only offers tips for actionable types", () => {
    expect(ctaFor("review_count")?.labelKey).toBe("achievements.cta.review")
    expect(ctaFor("login_streak")).toBeNull()
    expect(ctaFor("library_total_count")?.to(5)).toEqual({ name: "user-library", params: { userId: 5 } })
  })
})

const i18n = useRealI18n()
const router = createRouter({
  history: createMemoryHistory(),
  routes: ["movies", "feed", "settings", "home", "user-profile", "user-library"].map((name) => ({ name, path: `/${name}/:userId?`, component: { render: () => null } })),
})
const wrappers: VueWrapper[] = []

function mountView(viewerId: number) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: viewerId } as never
  const wrapper = mount(AchievementsView, {
    global: { plugins: [i18n, pinia, router, [VueQueryPlugin, { queryClient: new QueryClient() }]] },
    attachTo: document.body,
  })
  wrappers.push(wrapper)
  return wrapper
}

vi.mock("vue-router", async (orig) => ({ ...(await orig<typeof import("vue-router")>()), useRoute: () => ({ params: { userId: "1" } }) }))

describe("AchievementsView", () => {
  beforeEach(async () => {
    nextId = 1
    i18n.global.locale.value = "en"
    await router.push("/movies")
    m.profile.mockResolvedValue({ data: { user: { id: 1, username: "arin", display_name: "Arin", avatar_url: null, level: 4 } } })
    m.list.mockResolvedValue({ data: sample(), pagination: { page: 1, limit: 100, total: 6, total_pages: 1 } })
  })
  afterEach(() => {
    wrappers.splice(0).forEach((w) => w.unmount())
    document.body.innerHTML = ""
    vi.clearAllMocks()
  })

  it("renders hero, sections and families for the owner", async () => {
    const w = mountView(1)
    await flushPromises()
    expect(w.text()).toContain("My achievements")
    expect(w.text()).toContain("Level 4")
    expect(w.text()).toContain("Latest unlock")
    expect(w.text()).toContain("Next up")
    expect(w.findAll(".as")).toHaveLength(4)
    expect(w.findAll(".af")).toHaveLength(1)
    expect(w.findAll(".af__step")).toHaveLength(3)
    expect(w.find(".af__step.is-current").exists()).toBe(true)
  })

  it("uses visitor wording and hides the owner tip in the modal", async () => {
    const w = mountView(2)
    await flushPromises()
    expect(w.text()).toContain("Arin's achievements")
    await w.find(".af__main").trigger("click")
    await flushPromises()
    const dialog = document.body.querySelector('[role="dialog"]')!
    expect(dialog.textContent).toContain("Critic II description")
    expect(dialog.textContent).toContain("4 / 10")
    expect(dialog.textContent).not.toContain("Find something")
  })

  it("shows the action tip for the owner on a locked badge", async () => {
    const w = mountView(1)
    await flushPromises()
    await w.find(".af__main").trigger("click")
    await flushPromises()
    const dialog = document.body.querySelector('[role="dialog"]')!
    expect(dialog.textContent).toContain("Write a review")
    expect(dialog.textContent).toContain("Tier II")
  })

  it("filters by status and clears to the no-match state", async () => {
    const w = mountView(1)
    await flushPromises()
    await w.findAll(".tb__chip")[3]!.trigger("click")
    expect(w.findAll(".as")).toHaveLength(1)
    await w.find(".tb__input").setValue("zzzz")
    expect(w.findAll(".as")).toHaveLength(0)
    expect(w.text()).toContain("Clear")
  })

  it("shows the empty, error and not found states", async () => {
    m.list.mockResolvedValue({ data: [], pagination: { page: 1, limit: 100, total: 0, total_pages: 0 } })
    let w = mountView(1)
    await flushPromises()
    expect(w.text()).toContain("No achievements yet")
    w.unmount()

    m.list.mockRejectedValue(new Error("boom"))
    vi.spyOn(console, "error").mockImplementation(() => {})
    w = mountView(1)
    await flushPromises()
    expect(w.text()).toContain("Could not load achievements")
    w.unmount()

    m.list.mockResolvedValue({ data: sample(), pagination: { page: 1, limit: 100, total: 6, total_pages: 1 } })
    m.profile.mockRejectedValue(Object.assign(new Error("nf"), { isAxiosError: true, response: { status: 404 } }))
    w = mountView(1)
    await flushPromises()
    expect(w.text()).toContain("User not found")
  })
})
