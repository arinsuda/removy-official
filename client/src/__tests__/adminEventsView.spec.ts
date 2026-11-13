import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import AdminEventsView from "@/views/admin/AdminEventsView.vue"
import { adminEventApi } from "@/api/endpoints/event"
import type { AdminEvent } from "@/types/event"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({
  adminEventApi: {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
    uploadImage: vi.fn(),
    removeImage: vi.fn(),
  },
}))

const i18n = useRealI18n()
const api = vi.mocked(adminEventApi)
const wrappers: VueWrapper[] = []

const makeAdminEvent = (over: Partial<AdminEvent> = {}): AdminEvent => ({
  id: 1,
  type: "announcement",
  media_type: null,
  period: null,
  title: { en: "Admin Event EN", th: "กิจกรรมผู้ดูแล TH" },
  body: { en: "Admin Body EN", th: "เนื้อหาผู้ดูแล TH" },
  image_url: null,
  link_url: "/movies",
  starts_at: "2026-11-15T00:00:00.000Z",
  ends_at: "2026-11-20T00:00:00.000Z",
  show_popup: true,
  popup_priority: 5,
  state: "active",
  status: "published",
  created_at: "2026-11-10T00:00:00.000Z",
  updated_at: "2026-11-10T00:00:00.000Z",
  ...over,
})

describe("AdminEventsView", () => {
  let router: ReturnType<typeof createRouter>

  beforeEach(() => {
    setActivePinia(createPinia())
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/admin/events", name: "admin-events", component: AdminEventsView },
        { path: "/events/:id", name: "event-detail", component: { template: "<div>Detail</div>" } },
      ],
    })
  })

  afterEach(() => {
    wrappers.forEach((w) => w.unmount())
    wrappers.length = 0
    vi.clearAllMocks()
  })

  it("renders event list from api", async () => {
    api.list.mockResolvedValueOnce({
      data: {
        events: [makeAdminEvent({ id: 5, title: { en: "Grand Launch", th: "เปิดตัวยิ่งใหญ่" } })],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 },
      },
    } as never)

    await router.push("/admin/events")
    await router.isReady()

    const wrapper = mount(AdminEventsView, {
      global: { plugins: [i18n, router] },
    })
    wrappers.push(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain("Grand Launch")
    expect(wrapper.text()).toContain("#5")
  })

  it("opens create event modal when clicking 'New Event'", async () => {
    api.list.mockResolvedValueOnce({
      data: {
        events: [],
        pagination: { page: 1, limit: 20, total: 0, total_pages: 0 },
      },
    } as never)

    await router.push("/admin/events")
    await router.isReady()

    const wrapper = mount(AdminEventsView, {
      global: { plugins: [i18n, router] },
      attachTo: document.body,
    })
    wrappers.push(wrapper)
    await flushPromises()

    const newBtn = wrapper.find(".app-btn--primary")
    expect(newBtn.exists()).toBe(true)
    await newBtn.trigger("click")
    await flushPromises()

    expect(document.body.textContent).toContain("ชื่อกิจกรรม (EN)")
    expect(document.body.textContent).toContain("ชื่อกิจกรรม (TH)")
  })
})
