/**
 * @vitest-environment happy-dom
 */
import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import NotificationsView from "@/views/NotificationsView.vue"
import NotificationDropdown from "@/components/notifications/NotificationDropdown.vue"
import NotificationItem from "@/components/notifications/NotificationItem.vue"
import { i18n } from "@/i18n"
import { useAuthStore } from "@/stores/auth"
import { useNotificationStore } from "@/stores/notification"
import type { AppNotification } from "@/types/notification"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/notification", () => ({
  notiApi: {
    list: vi.fn(),
    getUnreadCount: vi.fn(),
    markAsRead: vi.fn().mockResolvedValue({}),
    markAllAsRead: vi.fn().mockResolvedValue({}),
    deleteNotification: vi.fn().mockResolvedValue({}),
    clearAll: vi.fn(),
  },
}))
vi.mock("@/api/api", () => ({ followApi: { acceptRequest: vi.fn(), rejectRequest: vi.fn() } }))

import { notiApi } from "@/api/endpoints/notification"
import { followApi } from "@/api/api"

useRealI18n()

const stub = { template: "<div />" }
const makeRouter = () =>
  createRouter({
    history: createMemoryHistory(),
    routes: [
      { name: "notifications", path: "/notifications", component: NotificationsView },
      { path: "/users/:id", component: stub },
      { path: "/movies/:id", component: stub },
    ],
  })

const NOW = new Date()
const hoursAgo = (h: number) => new Date(NOW.getTime() - h * 3_600_000).toISOString()
const daysAgo = (d: number) => new Date(NOW.getFullYear(), NOW.getMonth(), NOW.getDate() - d, 9).toISOString()

let nextId = 100
const noti = (over: Partial<AppNotification> = {}): AppNotification => ({
  id: nextId++,
  type: "review_liked",
  category: "social",
  message: "Someone liked your review",
  is_read: false,
  created_at: new Date().toISOString(),
  ...over,
})

const respond = (notifications: AppNotification[], total_pages = 1) =>
  ({ data: { notifications, unread_count: notifications.filter((n) => !n.is_read).length, pagination: { page: 1, limit: 20, total: notifications.length, total_pages } } }) as never

async function mountPage(query = "") {
  const router = makeRouter()
  router.push(`/notifications${query}`)
  await router.isReady()
  const wrapper = mount(NotificationsView, { global: { plugins: [createPinia(), router, i18n] } })
  await flushPromises()
  return { wrapper, router }
}

beforeEach(() => {
  i18n.global.locale.value = "en"
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.spyOn(console, "error").mockImplementation(() => {})
})

describe("NotificationsView", () => {
  it("shows skeletons, then groups by Today / Yesterday / Earlier", async () => {
    vi.mocked(notiApi.list).mockResolvedValue(
      respond([noti({ id: 1, created_at: hoursAgo(0) }), noti({ id: 2, created_at: daysAgo(1), is_read: true }), noti({ id: 3, created_at: daysAgo(5), is_read: true })]),
    )
    const { wrapper } = await mountPage()
    expect(wrapper.findAll(".noti-group__title").map((h) => h.text())).toEqual(["Today", "Yesterday", "Earlier"])
    expect(wrapper.findAll(".noti")).toHaveLength(3)
    expect(wrapper.findAll(".noti.is-unread")).toHaveLength(1)
  })

  it("passes category and unread filters from the URL to the API", async () => {
    vi.mocked(notiApi.list).mockResolvedValue(respond([]))
    await mountPage("?category=social&unread=1")
    expect(notiApi.list).toHaveBeenCalledWith(expect.objectContaining({ category: "social", unread: true }))
  })

  it("pushes filter changes to the URL and reloads", async () => {
    vi.mocked(notiApi.list).mockResolvedValue(respond([noti()]))
    const { wrapper, router } = await mountPage()
    await wrapper.findAll('[role="tab"]')[2]!.trigger("click")
    await flushPromises()
    expect(router.currentRoute.value.query.category).toBe("media")
    await wrapper.find(".noti-filters__unread").trigger("click")
    await flushPromises()
    expect(router.currentRoute.value.query.unread).toBe("1")
    expect(notiApi.list).toHaveBeenLastCalledWith(expect.objectContaining({ category: "media", unread: true }))
  })

  it("shows an empty state that matches the filter, and an error with retry", async () => {
    vi.mocked(notiApi.list).mockResolvedValue(respond([]))
    expect((await mountPage()).wrapper.text()).toContain("You are all caught up")
    expect((await mountPage("?unread=1")).wrapper.text()).toContain("No unread notifications")

    vi.mocked(notiApi.list).mockRejectedValueOnce({ response: { status: 500 } })
    const { wrapper } = await mountPage()
    expect(wrapper.find('[role="alert"]').text()).toContain("Could not load notifications")
    await wrapper.find('[role="alert"] button').trigger("click")
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })

  it("marks all as read from the header button", async () => {
    vi.mocked(notiApi.list).mockResolvedValue(respond([noti({ id: 1 }), noti({ id: 2 })]))
    const { wrapper } = await mountPage()
    const store = useNotificationStore()
    store.unreadCount = 2
    await flushPromises()
    await wrapper.find(".noti-page__header button").trigger("click")
    await flushPromises()
    expect(notiApi.markAllAsRead).toHaveBeenCalled()
    expect(wrapper.findAll(".noti.is-unread")).toHaveLength(0)
  })

  it("animates a notification that arrives over the socket, respecting the filter", async () => {
    vi.mocked(notiApi.list).mockResolvedValue(respond([noti({ id: 1 })]))
    const { wrapper } = await mountPage("?category=social")
    const store = useNotificationStore()

    store.prependNotification(noti({ id: 500, category: "media" }))
    await flushPromises()
    expect(wrapper.findAll(".noti")).toHaveLength(1)

    store.prependNotification(noti({ id: 501, category: "social" }))
    await flushPromises()
    expect(wrapper.findAll(".noti")).toHaveLength(2)
    expect(wrapper.find(".noti.is-fresh").exists()).toBe(true)
    expect(wrapper.find(".noti__new").text()).toBe("New")
  })

  it("loads more pages", async () => {
    vi.mocked(notiApi.list).mockResolvedValueOnce(respond([noti({ id: 1 })], 2))
    const { wrapper } = await mountPage()
    vi.mocked(notiApi.list).mockResolvedValueOnce(respond([noti({ id: 2 })], 2))
    await wrapper.find(".noti-page__more button").trigger("click")
    await flushPromises()
    expect(wrapper.findAll(".noti")).toHaveLength(2)
    expect(notiApi.list).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }))
  })
})

describe("NotificationItem", () => {
  const mountItem = (n: AppNotification) => {
    const pinia = createPinia()
    setActivePinia(pinia)
    useAuthStore().user = { id: 7 } as never
    return mount(NotificationItem, { props: { notification: n }, global: { plugins: [pinia, i18n] } })
  }
  const request = () => noti({ type: "follow_requested", category: "social", actor: { id: 3, username: "kenji", display_name: "Kenji" }, message: "kenji requested to follow you" })

  it("builds follow messages from the actor name in the current language", () => {
    expect(mountItem(request()).find(".noti__message").text()).toBe("Kenji requested to follow you")
  })

  it("accepts a follow request inline and reports it", async () => {
    vi.mocked(followApi.acceptRequest).mockResolvedValue({} as never)
    const w = mountItem(request())
    await w.find(".noti__actions button").trigger("click")
    await flushPromises()
    expect(followApi.acceptRequest).toHaveBeenCalledWith(7, 3)
    expect(w.find(".noti__result").text()).toBe("Request accepted")
    expect(w.emitted("responded")).toHaveLength(1)
  })

  it("declines a request and keeps the buttons when the call fails", async () => {
    vi.mocked(followApi.rejectRequest).mockResolvedValueOnce({} as never)
    const w = mountItem(request())
    await w.findAll(".noti__actions button")[1]!.trigger("click")
    await flushPromises()
    expect(followApi.rejectRequest).toHaveBeenCalledWith(7, 3)
    expect(w.find(".noti__result").text()).toBe("Request declined")

    vi.mocked(followApi.acceptRequest).mockRejectedValueOnce({ response: { status: 404 } })
    const failed = mountItem(request())
    await failed.find(".noti__actions button").trigger("click")
    await flushPromises()
    expect(failed.find(".noti__actions button").exists()).toBe(true)
  })
})

describe("NotificationDropdown", () => {
  it("shows only the latest five with a link to all", async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const store = useNotificationStore()
    store.notifications = Array.from({ length: 8 }, (_, i) => noti({ id: i + 1 }))
    store.unreadCount = 8
    const router = makeRouter()
    router.push("/notifications")
    await router.isReady()
    const w = mount(NotificationDropdown, { global: { plugins: [pinia, router, i18n] } })
    expect(w.findAll(".noti")).toHaveLength(5)
    expect(w.find('a[href="/notifications"]').text()).toContain("View all notifications")
    await w.find(".noti-dropdown__head button").trigger("click")
    await flushPromises()
    expect(notiApi.markAllAsRead).toHaveBeenCalled()
  })

  it("shows the empty state", () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const w = mount(NotificationDropdown, { global: { plugins: [pinia, makeRouter(), i18n] } })
    expect(w.text()).toContain("No notifications at this time")
  })
})
