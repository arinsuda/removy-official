import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import EventPopup from "@/components/events/EventPopup.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { eventApi } from "@/api/endpoints/event"
import { useAuthStore } from "@/stores/auth"
import type { AppEvent } from "@/types/event"
import { EVENT_POPUP_HIDE_KEY_PREFIX, EVENT_POPUP_SESSION_KEY_PREFIX, localDateKey } from "@/utils/events"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({
  eventApi: {
    list: vi.fn(),
    get: vi.fn(),
    popup: vi.fn(),
  },
}))

const i18n = useRealI18n()
const api = vi.mocked(eventApi)
const wrappers: VueWrapper[] = []

const makeEvent = (over: Partial<AppEvent> = {}): AppEvent => ({
  id: 1,
  type: "announcement",
  media_type: null,
  period: null,
  title: { en: "Popup Event", th: "กิจกรรมป๊อปอัป" },
  body: { en: "Important Announcement", th: "ประกาศสำคัญ" },
  image_url: null,
  link_url: "/events",
  starts_at: "2026-11-15T00:00:00.000Z",
  ends_at: "2026-11-20T00:00:00.000Z",
  show_popup: true,
  popup_priority: 10,
  state: "active",
  ...over,
})

describe("EventPopup", () => {
  let router: ReturnType<typeof createRouter>

  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    sessionStorage.clear()
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/", name: "home", component: { template: "<div>Home</div>" } },
        { path: "/admin", name: "admin", component: { template: "<div>Admin</div>" } },
        { path: "/events", name: "events", component: { template: "<div>Events</div>" } },
      ],
    })
  })

  afterEach(() => {
    wrappers.forEach((w) => w.unmount())
    wrappers.length = 0
    vi.clearAllMocks()
  })

  it("does not open when user is not logged in", async () => {
    const authStore = useAuthStore()
    authStore.user = null

    const wrapper = mount(EventPopup, {
      global: { plugins: [i18n, router] },
    })
    wrappers.push(wrapper)
    await flushPromises()

    expect(api.popup).not.toHaveBeenCalled()
    expect(wrapper.find(".app-modal").exists()).toBe(false)
  })

  it("opens modal and displays event when user is logged in", async () => {
    const authStore = useAuthStore()
    authStore.user = { id: 7, username: "alex", email: "alex@x.co", role: "user" } as never

    api.popup.mockResolvedValueOnce({
      data: { events: [makeEvent()] },
    } as never)

    const wrapper = mount(EventPopup, {
      global: { plugins: [i18n, router] },
      attachTo: document.body,
    })
    wrappers.push(wrapper)
    await flushPromises()

    expect(api.popup).toHaveBeenCalledTimes(1)
    expect(document.body.textContent).toContain("กิจกรรมป๊อปอัป")
  })

  it("saves to localStorage when 'Don't show again today' is checked on close", async () => {
    const authStore = useAuthStore()
    authStore.user = { id: 7, username: "alex", email: "alex@x.co", role: "user" } as never

    api.popup.mockResolvedValueOnce({
      data: { events: [makeEvent()] },
    } as never)

    const wrapper = mount(EventPopup, {
      global: { plugins: [i18n, router] },
      attachTo: document.body,
    })
    wrappers.push(wrapper)
    await flushPromises()

    const checkbox = document.body.querySelector('input[type="checkbox"]') as HTMLInputElement
    expect(checkbox).not.toBeNull()
    checkbox.checked = true
    checkbox.dispatchEvent(new Event("change"))
    await flushPromises()

    // Find and click the close button in the footer
    const buttons = Array.from(document.body.querySelectorAll("button"))
    const closeBtn = buttons.find((b) => b.textContent?.includes("Close") || b.textContent?.includes("ปิด"))
    expect(closeBtn).toBeDefined()
    closeBtn?.click()
    await flushPromises()

    expect(localStorage.getItem(`${EVENT_POPUP_HIDE_KEY_PREFIX}7`)).toBe(localDateKey())
  })

  it("records a session dismissal when AppModal reports close from the X button", async () => {
    const authStore = useAuthStore()
    authStore.user = { id: 7, username: "alex", email: "alex@x.co", role: "user" } as never
    api.popup.mockResolvedValueOnce({ data: { events: [makeEvent()] } } as never)

    const wrapper = mount(EventPopup, { global: { plugins: [i18n, router] }, attachTo: document.body })
    wrappers.push(wrapper)
    await flushPromises()

    wrapper.findComponent(AppModal).vm.$emit("close", "close-button")
    await flushPromises()

    expect(sessionStorage.getItem(`${EVENT_POPUP_SESSION_KEY_PREFIX}7`)).toBe("1")
    expect(localStorage.getItem(`${EVENT_POPUP_HIDE_KEY_PREFIX}7`)).toBeNull()
  })

  it("keeps 'don't show again today' when the modal is dismissed with Escape", async () => {
    const authStore = useAuthStore()
    authStore.user = { id: 7, username: "alex", email: "alex@x.co", role: "user" } as never
    api.popup.mockResolvedValueOnce({ data: { events: [makeEvent()] } } as never)

    const wrapper = mount(EventPopup, { global: { plugins: [i18n, router] }, attachTo: document.body })
    wrappers.push(wrapper)
    await flushPromises()

    const checkbox = document.body.querySelector('input[type="checkbox"]') as HTMLInputElement
    checkbox.checked = true
    checkbox.dispatchEvent(new Event("change"))
    await flushPromises()

    document.body.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }))
    await flushPromises()

    expect(localStorage.getItem(`${EVENT_POPUP_HIDE_KEY_PREFIX}7`)).toBe(localDateKey())
  })

  it("shows the position counter through i18n", async () => {
    const authStore = useAuthStore()
    authStore.user = { id: 7, username: "alex", email: "alex@x.co", role: "user" } as never
    api.popup.mockResolvedValueOnce({
      data: { events: [makeEvent({ id: 1 }), makeEvent({ id: 2 })] },
    } as never)

    const wrapper = mount(EventPopup, { global: { plugins: [i18n, router] }, attachTo: document.body })
    wrappers.push(wrapper)
    await flushPromises()

    expect(document.body.querySelector(".event-popup__counter")?.textContent?.trim()).toBe("1 / 2")
  })
})
