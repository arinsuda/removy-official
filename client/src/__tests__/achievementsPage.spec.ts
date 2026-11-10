import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import AchievementsView from "@/views/achievement/AchievementsView.vue"
import AchievementSquareCard from "@/components/share/cards/AchievementSquareCard.vue"
import AchievementStoryCard from "@/components/share/cards/AchievementStoryCard.vue"
import AchievementShareSheet from "@/components/share/AchievementShareSheet.vue"
import { categoryOf, ctaFor } from "@/components/achievement/config"
import { buildUnits, countByStatus, filterUnits, groupByCategory, latestUnlock, matchedCount, nextUp, sortUnits, summarize } from "@/components/achievement/units"
import type { UserAchievement } from "@/types/achievement"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ list: vi.fn(), profile: vi.fn() }))
vi.mock("@/api/api", () => ({ userApi: { getProfile: m.profile } }))
vi.mock("@/api/endpoints/social", async (orig) => ({
  ...(await orig<typeof import("@/api/endpoints/social")>()),
  followApi: { getFollowStatus: vi.fn().mockResolvedValue({ data: {} }), follow: vi.fn(), unfollow: vi.fn() },
}))
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
    expect(countByStatus(filterUnits(units, { status: "all", category: "all", search: "" }))).toEqual({ all: 6, unlocked: 2, progress: 3, locked: 1 })
    expect(countByStatus(filterUnits(units, { status: "all", category: "reviews", search: "" }))).toEqual({ all: 3, unlocked: 1, progress: 2, locked: 0 })
    expect(filterUnits(units, { status: "locked", category: "all", search: "" })).toHaveLength(1)
    expect(filterUnits(units, { status: "all", category: "reviews", search: "" })).toHaveLength(1)
    expect(filterUnits(units, { status: "all", category: "all", search: "critic iii" })).toHaveLength(1)
  })

  it("keeps a family under a status filter when any tier matches and tracks the matching tiers", () => {
    const units = buildUnits(sample())
    const unlocked = filterUnits(units, { status: "unlocked", category: "all", search: "" })
    expect(unlocked.map((u) => u.key)).toContain("family:review_count")
    const critic = unlocked.find((u) => u.key === "family:review_count")!
    expect([...critic.matched!]).toEqual([critic.tiers[0]!.achievement_id])
    expect(matchedCount(unlocked)).toBe(2)
    expect(matchedCount(filterUnits(units, { status: "progress", category: "all", search: "" }))).toBe(3)
    const search = filterUnits(units, { status: "all", category: "all", search: "critic iii", describe: () => "" })
    expect(countByStatus(search)).toEqual({ all: 1, unlocked: 0, progress: 1, locked: 0 })
  })

  it("sorts and groups into category sections", () => {
    const units = buildUnits(sample())
    expect(sortUnits(units, "recent", "en")[0]!.current.achievement.name).toBe("Binger")
    expect(sortUnits(units, "progress", "en")[0]!.current.achievement.name).toBe("Shelf")
    expect(sortUnits(units, "name", "en").map((u) => u.current.achievement.name)).toEqual(["Binger", "Critic II", "Shelf", "Socialite"])
    const sections = groupByCategory(units)
    expect(sections.map((s) => s.key)).toEqual(["watching", "reviews", "social", "library"])
    expect(sections.find((s) => s.key === "reviews")).toMatchObject({ unlocked: 1, total: 3 })
    const filteredSections = groupByCategory(filterUnits(units, { status: "progress", category: "all", search: "" }))
    expect(filteredSections.find((s) => s.key === "reviews")).toMatchObject({ unlocked: 0, total: 2 })
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
    expect(dialog.textContent).toContain("Write 10 reviews")
    expect(dialog.textContent).toContain("4 / 10")
    expect(dialog.textContent).not.toContain("Find something")
  })

  it("falls back to the server description for an unknown action type", async () => {
    m.list.mockResolvedValue({ data: [ach("Mystery", "brand_new_type", 5, 1)], pagination: { page: 1, limit: 100, total: 1, total_pages: 1 } })
    const w = mountView(1)
    await flushPromises()
    await w.find(".ab").trigger("click")
    await flushPromises()
    expect(document.body.querySelector('[role="dialog"]')!.textContent).toContain("Mystery description")
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
    expect(w.text()).toContain("1 of 6 achievements")
    await w.find(".tb__input").setValue("zzzz")
    expect(w.findAll(".as")).toHaveLength(0)
    expect(w.text()).toContain("Clear")
  })

  it("counts individual achievements in the chips and the toolbar", async () => {
    const w = mountView(1)
    await flushPromises()
    const chips = w.findAll(".tb__chip").map((c) => c.text())
    expect(chips.join("|")).toMatch(/All.*6.*Unlocked.*2.*In progress.*3.*Locked.*1/)
    expect(w.text()).toContain("6 of 6 achievements")
    await w.findAll(".tb__chip")[1]!.trigger("click")
    expect(w.text()).toContain("2 of 6 achievements")
    expect(w.findAll(".af__step.is-dim")).toHaveLength(2)
    expect(w.findAll(".as__count").map((c) => c.text())).toEqual(["1 / 1 unlocked", "1 / 1 unlocked"])
  })

  it("marks a single badge after a family card as full-width on mobile", async () => {
    m.list.mockResolvedValue({
      data: [
        ach("Critic I", "review_count", 1, 1),
        ach("Critic II", "review_count", 10, 4),
        ach("Contrarian", "contrarian_review_count", 1, 0),
      ],
      pagination: { page: 1, limit: 100, total: 3, total_pages: 1 },
    })
    const w = mountView(1)
    await flushPromises()
    const cells = w.findAll(".as__cell")
    expect(cells).toHaveLength(2)
    expect(cells[0]!.classes()).toContain("as__cell--wide")
    expect(cells[1]!.classes()).toContain("as__cell--solo")
  })

  it("shows the locked state instead of an error when the list answers 403", async () => {
    m.list.mockRejectedValue(Object.assign(new Error("forbidden"), { isAxiosError: true, response: { status: 403 } }))
    const w = mountView(2)
    await flushPromises()
    expect(w.text()).toContain("This account is private")
    expect(w.text()).not.toContain("Could not load achievements")
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

  it("shows the share button only for the owner on an unlocked achievement", async () => {
    const owner = mountView(1)
    await flushPromises()
    // Click unlocked single badge "Binger"
    await owner.find(".ab").trigger("click")
    await flushPromises()
    expect(document.body.querySelector('[data-test="achievement-share-btn"]')).not.toBeNull()
    owner.unmount()
    document.body.innerHTML = ""

    // Owner clicks locked/in-progress family card ("Critic II")
    const ownerLocked = mountView(1)
    await flushPromises()
    await ownerLocked.find(".af__main").trigger("click")
    await flushPromises()
    expect(document.body.querySelector('[data-test="achievement-share-btn"]')).toBeNull()
    ownerLocked.unmount()
    document.body.innerHTML = ""

    // Visitor clicks unlocked single badge "Binger"
    const visitor = mountView(2)
    await flushPromises()
    await visitor.find(".ab").trigger("click")
    await flushPromises()
    expect(document.body.querySelector('[data-test="achievement-share-btn"]')).toBeNull()
  })
})

describe("Achievement share cards", () => {
  afterEach(() => {
    wrappers.splice(0).forEach((w) => w.unmount())
    document.body.innerHTML = ""
    vi.restoreAllMocks()
  })

  it("renders localized requirement text, multi-tier badge, formatted date, and rejects raw avatar URLs", () => {
    i18n.global.locale.value = "en"
    const items = sample()
    const criticI = items[0]!
    const familyTiers = items.slice(0, 3)

    for (const Card of [AchievementSquareCard, AchievementStoryCard]) {
      const w = mount(Card, {
        props: {
          achievement: criticI,
          tiers: familyTiers,
          username: "arin",
          avatarUrl: "https://cdn.example.com/raw-avatar.png",
          theme: "crimson",
        },
        global: { plugins: [i18n] },
      })
      wrappers.push(w)
      // Names come from i18n by action type + target (review_count_1), not from the server string.
      expect(w.text()).toContain(i18n.global.t("achievements.names.review_count_1"))
      expect(w.text()).not.toContain("Critic I")
      expect(w.text()).toContain("Write your first review")
      expect(w.text()).not.toContain("Critic I description")
      expect(w.text()).toContain("Tier I")
      expect(w.text()).toContain("Mar")
      expect(w.text()).toContain("@arin")
      expect(w.findAll("img")).toHaveLength(0)
      expect(w.html()).not.toContain("https://cdn.example.com/raw-avatar.png")
    }
  })

  it("renders proxied data URLs for avatars and switches description locale to Thai", () => {
    i18n.global.locale.value = "th"
    const binger = sample()[3]!
    const dataUrl = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="
    const w = mount(AchievementSquareCard, {
      props: {
        achievement: binger,
        tiers: [binger],
        username: "arin",
        avatarUrl: dataUrl,
        theme: "gold",
      },
      global: { plugins: [i18n] },
    })
    wrappers.push(w)
    expect(w.text()).toContain("ดูหนังหรือซีรีส์ครบ 10 เรื่อง")
    expect(w.text()).not.toContain("Binger description")
    expect(w.find("img").attributes("src")).toBe(dataUrl)
  })

  it("inlines the user avatar as a data URL without the TMDB-only proxy", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(new Blob(["img"], { type: "image/png" }), { status: 200 }),
    )
    const pinia = createPinia()
    setActivePinia(pinia)

    const binger = sample()[3]!
    const w = mount(AchievementShareSheet, {
      props: {
        modelValue: true,
        achievement: binger,
        tiers: [binger],
        username: "arin",
        avatarUrl: "/uploads/avatars/arin.png",
      },
      global: { plugins: [i18n, pinia] },
      attachTo: document.body,
    })
    wrappers.push(w)
    await flushPromises()

    expect(fetchSpy).toHaveBeenCalledTimes(1)
    const calledUrl = String(fetchSpy.mock.calls[0]![0])
    expect(calledUrl).toBe(`${window.location.origin}/uploads/avatars/arin.png`)
    expect(calledUrl).not.toContain("/share/proxy-image")
  })
})
