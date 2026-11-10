import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import AdminLayout from "@/layouts/AdminLayout.vue"
import AdminDashboardView from "@/views/admin/AdminDashboardView.vue"
import AdminUsersView from "@/views/admin/AdminUsersView.vue"
import AdminReviewsView from "@/views/admin/AdminReviewsView.vue"
import AdminAuditLogsView from "@/views/admin/AdminAuditLogsView.vue"
import { adminApi } from "@/api/api"
import { useUIStore } from "@/stores/ui"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", async (orig) => ({
  ...(await orig<typeof import("@/api/api")>()),
  adminApi: {
    getOverview: vi.fn(),
    getGrowth: vi.fn(),
    listUsers: vi.fn(),
    updateUserRole: vi.fn(),
    updateUserStatus: vi.fn(),
    deleteUser: vi.fn(),
    eraseUser: vi.fn(),
    listReviews: vi.fn(),
    deleteReview: vi.fn(),
    listAuditLogs: vi.fn(),
  },
}))

const i18n = useRealI18n()
const api = vi.mocked(adminApi)
const wrappers: VueWrapper[] = []

const page = <T>(items: T[]) => ({ data: { items, total: items.length, page: 1, limit: 20, total_pages: 1 } }) as never
const user = (over = {}) => ({
  id: 5,
  username: "jane",
  email: "jane@x.co",
  display_name: "Jane",
  avatar_url: null,
  role: "user",
  is_active: true,
  created_at: "2026-01-02T00:00:00Z",
  deleted_at: null,
  erasure_due_at: null,
  review_count: 3,
  ...over,
})
const review = (over = {}) => ({
  id: 9,
  user_id: 5,
  username: "jane",
  media_id: 1,
  media_type: "movie",
  media_title: "Dune",
  rating: 4.5,
  body: "Great film",
  is_public: true,
  like_count: 2,
  created_at: "2026-01-02T00:00:00Z",
  ...over,
})

async function mountView(view: object, path = "/admin") {
  const stub = { template: "<div />" }
  const routed = view === AdminLayout ? stub : view
  setActivePinia(createPinia())
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/admin", name: "admin-dashboard", component: routed },
      { path: "/admin/users", name: "admin-users", component: routed },
      { path: "/admin/reviews", name: "admin-reviews", component: routed },
      { path: "/admin/audit-logs", name: "admin-audit-logs", component: routed },
      { path: "/:p(.*)*", name: "home", component: { template: "<div />" } },
    ],
  })
  await router.push(path)
  const w = mount(view, { global: { plugins: [i18n, router] }, attachTo: document.body })
  wrappers.push(w)
  await flushPromises()
  return w
}

const menuTrigger = (label: string) => document.body.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`)!
const menuItem = (label: string) =>
  Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="menuitem"]')).find((b) => b.textContent?.includes(label))!
const dialogButton = (label: string) =>
  Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="alertdialog"] button')).find((b) => b.textContent?.includes(label))!

beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
  i18n.global.locale.value = "en"
  HTMLElement.prototype.getClientRects = function (this: HTMLElement) {
    return (this.isConnected ? [{ width: 1, height: 1 }] : []) as unknown as DOMRectList
  }
})
afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
})

describe("AdminLayout", () => {
  it("renders the nav with the current page marked and a drawer toggle", async () => {
    const w = await mountView(AdminLayout, "/admin/users")
    const links = w.findAll("nav a")
    expect(links.map((l) => l.attributes("href"))).toEqual(["/admin", "/admin/users", "/admin/reviews", "/admin/audit-logs"])
    expect(links[1]!.attributes("aria-current")).toBe("page")
    const menu = w.get(".admin-topbar__menu")
    expect(menu.attributes("aria-expanded")).toBe("false")
    await menu.trigger("click")
    expect(w.get(".admin-topbar__menu").attributes("aria-expanded")).toBe("true")
    await w.get("aside").trigger("keydown", { key: "Escape" })
    expect(w.get(".admin-topbar__menu").attributes("aria-expanded")).toBe("false")
  })

  it("collapses the sidebar and remembers it", async () => {
    const w = await mountView(AdminLayout)
    await w.get(".admin-sidebar__toggle").trigger("click")
    expect(w.get(".admin-shell").classes()).toContain("is-collapsed")
    expect(localStorage.getItem("admin_sidebar_collapsed")).toBe("true")
  })
})

describe("AdminDashboardView", () => {
  const overview = {
    total_registered_users: 120,
    active_users_count: 110,
    inactive_users_count: 10,
    current_month_registrations: 20,
    previous_month_registrations: 10,
    absolute_growth: 10,
    growth_percentage: 100,
    growth_status: "growth",
    unique_online_users: 4,
    total_activity_events: 900,
    activity_events_today: 12,
    dau_today: 7,
    wau_7d: 30,
    mau_30d: 80,
    total_reviews: 55,
    total_media_likes: 66,
  }

  it("renders KPIs and an accessible chart with a data table", async () => {
    api.getOverview.mockResolvedValue({ data: { overview } } as never)
    api.getGrowth.mockResolvedValue({ data: { growth: [{ month: "2026-01", user_count: 3 }, { month: "2026-02", user_count: 9 }] } } as never)
    const w = await mountView(AdminDashboardView)
    expect(w.get("h1").text()).toBe("Admin Dashboard")
    expect(w.text()).toContain("120")
    expect(w.text()).toContain("+100.0%")
    expect(w.findAll("table tbody tr").map((r) => r.text().replace(/\s+/g, " "))).toEqual([expect.stringContaining("3"), expect.stringContaining("9")])
    expect(w.text()).toContain("Bar chart of new registrations per month. 2 months, peak 9")
    expect(w.findAll('[role="meter"]')).toHaveLength(3)
  })

  it("shows an error state with retry", async () => {
    api.getOverview.mockRejectedValue(new Error("x"))
    api.getGrowth.mockResolvedValue({ data: { growth: [] } } as never)
    const w = await mountView(AdminDashboardView)
    expect(w.text()).toContain("Couldn't load the dashboard")
    api.getOverview.mockResolvedValue({ data: { overview } } as never)
    await w.get("button").trigger("click")
    await flushPromises()
    expect(w.text()).toContain("120")
  })
})

describe("AdminUsersView", () => {
  it("lists users with sortable headers and requests the chosen sort", async () => {
    api.listUsers.mockResolvedValue(page([user()]))
    const w = await mountView(AdminUsersView)
    expect(w.text()).toContain("@jane")
    const joined = w.findAll("th").find((th) => th.text().includes("Joined"))!
    expect(joined.attributes("aria-sort")).toBe("descending")
    await w.findAll("th button").find((b) => b.text().includes("Email"))!.trigger("click")
    await flushPromises()
    expect(api.listUsers).toHaveBeenLastCalledWith(expect.objectContaining({ sort_by: "email", sort_order: "asc", page: 1 }))
  })

  it("never deactivates in one click: the menu opens a confirm modal first", async () => {
    api.listUsers.mockResolvedValue(page([user()]))
    api.updateUserStatus.mockResolvedValue({} as never)
    await mountView(AdminUsersView)
    menuTrigger("Actions for @jane").click()
    await flushPromises()
    menuItem("Deactivate").click()
    await flushPromises()
    expect(api.updateUserStatus).not.toHaveBeenCalled()
    const dialog = document.body.querySelector('[role="alertdialog"]')!
    expect(dialog.textContent).toContain("signed out of every device")
    dialogButton("Deactivate account").click()
    await flushPromises()
    expect(api.updateUserStatus).toHaveBeenCalledWith(5, false, undefined)
  })

  it("maps the final-admin protection error to a localized toast", async () => {
    api.listUsers.mockResolvedValue(page([user({ role: "admin" })]))
    api.updateUserRole.mockRejectedValue({
      response: { status: 409, data: { code: "FINAL_ADMIN", error: "cannot demote or deactivate final active admin" } },
    })
    await mountView(AdminUsersView)
    menuTrigger("Actions for @jane").click()
    await flushPromises()
    menuItem("Remove admin access").click()
    await flushPromises()
    dialogButton("Remove admin access").click()
    await flushPromises()
    expect(api.updateUserRole).toHaveBeenCalledWith(5, "user", undefined)
    expect(useUIStore().toasts.map((t) => t.message)).toContain("You can't demote, deactivate or delete the last active admin.")
  })

  it("requires confirmation to delete and disables role change for inactive users", async () => {
    api.listUsers.mockResolvedValue(page([user({ is_active: false })]))
    await mountView(AdminUsersView)
    menuTrigger("Actions for @jane").click()
    await flushPromises()
    expect(menuItem("Make admin").disabled).toBe(true)
    menuItem("Delete account").click()
    await flushPromises()
    expect(api.deleteUser).not.toHaveBeenCalled()
    expect(document.body.querySelector('[role="alertdialog"]')!.textContent).toContain("soft-deleted")
  })

  it("filters by deleted status and offers only erasure on deleted rows", async () => {
    api.listUsers.mockResolvedValue(page([user()]))
    const w = await mountView(AdminUsersView)
    const statusSelect = w.findAll("select").find((s) => s.findAll("option").some((o) => o.text() === "Deleted"))!
    api.listUsers.mockResolvedValue(page([user({ is_active: false, deleted_at: "2026-02-03T00:00:00Z" })]))
    await statusSelect.setValue("deleted")
    await flushPromises()
    expect(api.listUsers).toHaveBeenLastCalledWith(expect.objectContaining({ status: "deleted", page: 1 }))
    expect(w.text()).toContain("Deleted")
    menuTrigger("Actions for @jane").click()
    await flushPromises()
    const items = Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="menuitem"]')).map((b) => b.textContent!.trim())
    expect(items).toEqual(["Erase permanently"])
  })

  it("erases a deleted user only after the username is typed, then refreshes and toasts", async () => {
    api.listUsers.mockResolvedValue(page([user({ is_active: false, deleted_at: "2026-02-03T00:00:00Z" })]))
    api.eraseUser.mockResolvedValue({} as never)
    await mountView(AdminUsersView)
    menuTrigger("Actions for @jane").click()
    await flushPromises()
    menuItem("Erase permanently").click()
    await flushPromises()
    const dialog = document.body.querySelector('[role="alertdialog"]')!
    expect(dialog.textContent).toContain("cannot be undone")
    const input = dialog.querySelector<HTMLInputElement>("input")!
    input.value = "jane"
    input.dispatchEvent(new Event("input"))
    await flushPromises()
    api.listUsers.mockClear()
    dialogButton("Erase permanently").click()
    await flushPromises()
    expect(api.eraseUser).toHaveBeenCalledWith(5, { confirm_username: "jane", reason: undefined })
    expect(api.listUsers).toHaveBeenCalled()
    expect(useUIStore().toasts.map((t) => t.message)).toContain("Permanently erased @jane")
  })

  it("shows an error state and an empty-with-filters state", async () => {
    api.listUsers.mockRejectedValueOnce(new Error("x"))
    const w = await mountView(AdminUsersView)
    expect(w.text()).toContain("Couldn't load data")
    api.listUsers.mockResolvedValue(page([]))
    await w.get(".app-state button").trigger("click")
    await flushPromises()
    expect(w.text()).toContain("No users found")
  })

  it("renders Thai", async () => {
    i18n.global.locale.value = "th"
    api.listUsers.mockResolvedValue(page([user()]))
    const w = await mountView(AdminUsersView)
    expect(w.get("h1").text()).toBe("จัดการผู้ใช้")
  })
})

describe("AdminReviewsView", () => {
  it("deletes only after the confirm modal", async () => {
    api.listReviews.mockResolvedValue(page([review()]))
    api.deleteReview.mockResolvedValue({} as never)
    const w = await mountView(AdminReviewsView)
    expect(w.text()).toContain("Dune")
    menuTrigger("Actions for #9").click()
    await flushPromises()
    menuItem("Delete Review").click()
    await flushPromises()
    expect(api.deleteReview).not.toHaveBeenCalled()
    dialogButton("Delete review").click()
    await flushPromises()
    expect(api.deleteReview).toHaveBeenCalledWith(9, undefined)
  })

  it("opens the full review in a dialog", async () => {
    api.listReviews.mockResolvedValue(page([review()]))
    const w = await mountView(AdminReviewsView)
    await w.get(".body-button").trigger("click")
    await flushPromises()
    expect(document.body.textContent).toContain("Review #9")
    expect(document.body.textContent).toContain("Great film")
  })
})

describe("AdminAuditLogsView", () => {
  const log = (over = {}) => ({
    id: 1,
    admin_id: 1,
    admin_username: "root",
    action: "USER_DEACTIVATED",
    target_type: "user",
    target_id: 5,
    reason: "spam",
    meta_data: { previous: "active" },
    created_at: "2026-01-02T10:00:00Z",
    ...over,
  })

  it("renders a timeline with localized actions, reason and details", async () => {
    api.listAuditLogs.mockResolvedValue(page([log(), log({ id: 2, action: "REVIEW_DELETED", target_type: "review", reason: null, meta_data: null })]))
    const w = await mountView(AdminAuditLogsView)
    const entries = w.findAll(".entry")
    expect(entries).toHaveLength(2)
    expect(entries[0]!.text()).toContain("Account deactivated")
    expect(entries[0]!.text()).toContain("by @root")
    expect(entries[0]!.text()).toContain("spam")
    expect(entries[0]!.find("details").exists()).toBe(true)
    expect(entries[1]!.text()).toContain("No reason given")
    expect(entries[1]!.find("details").exists()).toBe(false)
  })

  it("filters by action and shows an empty state", async () => {
    api.listAuditLogs.mockResolvedValue(page([log()]))
    const w = await mountView(AdminAuditLogsView)
    api.listAuditLogs.mockResolvedValue(page([]))
    await w.findAll("select")[0]!.setValue("USER_DELETED")
    await flushPromises()
    expect(api.listAuditLogs).toHaveBeenLastCalledWith(expect.objectContaining({ action: "USER_DELETED", page: 1 }))
    expect(w.text()).toContain("No results")
  })
})
