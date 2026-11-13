import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createMemoryHistory, createRouter } from "vue-router"
import EventsView from "@/views/events/EventsView.vue"
import { eventApi } from "@/api/endpoints/event"
import type { AppEvent } from "@/types/event"
import { useRealI18n } from "./helpers/realI18n"
import { clickAfterRender } from "./helpers/waitForEnabled"

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
  title: { en: "Test Event EN", th: "กิจกรรมทดสอบ TH" },
  body: { en: "Test Body EN", th: "เนื้อหาทดสอบ TH" },
  image_url: null,
  link_url: "/movies",
  starts_at: "2026-11-15T00:00:00.000Z",
  ends_at: "2026-11-20T00:00:00.000Z",
  show_popup: true,
  popup_priority: 1,
  state: "active",
  ...over,
})

describe("EventsView", () => {
  let router: ReturnType<typeof createRouter>

  beforeEach(() => {
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/events", name: "events", component: EventsView },
        { path: "/events/:id", name: "event-detail", component: { template: "<div>Detail</div>" } },
        { path: "/movies", name: "movies", component: { template: "<div>Movies</div>" } },
      ],
    })
  })

  afterEach(() => {
    wrappers.forEach((w) => w.unmount())
    wrappers.length = 0
    vi.clearAllMocks()
  })

  it("renders empty state when there are no events", async () => {
    api.list.mockResolvedValueOnce({
      data: {
        events: [],
        pagination: { page: 1, limit: 12, total: 0, total_pages: 0 },
      },
    } as never)

    await router.push("/events")
    await router.isReady()

    const wrapper = mount(EventsView, {
      global: {
        plugins: [i18n, router],
      },
    })
    wrappers.push(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain("กิจกรรมและประกาศ")
    expect(wrapper.find(".app-state").exists()).toBe(true)
  })

  it("renders event cards when events are returned", async () => {
    api.list.mockResolvedValueOnce({
      data: {
        events: [makeEvent({ id: 10, title: { en: "Cool Event", th: "กิจกรรมเจ๋งๆ" } })],
        pagination: { page: 1, limit: 12, total: 1, total_pages: 1 },
      },
    } as never)

    await router.push("/events")
    await router.isReady()

    const wrapper = mount(EventsView, {
      global: {
        plugins: [i18n, router],
      },
    })
    wrappers.push(wrapper)
    await flushPromises()

    expect(wrapper.find(".events-grid").exists()).toBe(true)
    expect(wrapper.text()).toContain("กิจกรรมเจ๋งๆ")
  })

  it("fetches page 2 when Next is clicked", async () => {
    api.list
      .mockResolvedValueOnce({
        data: { events: [makeEvent()], pagination: { page: 1, limit: 12, total: 13, total_pages: 2 } },
      } as never)
      .mockResolvedValueOnce({
        data: { events: [makeEvent({ id: 2 })], pagination: { page: 2, limit: 12, total: 13, total_pages: 2 } },
      } as never)

    await router.push("/events")
    await router.isReady()

    const wrapper = mount(EventsView, { global: { plugins: [i18n, router] } })
    wrappers.push(wrapper)
    await flushPromises()

    const nextLabel = i18n.global.t("admin.pagination.next")
    const next = wrapper.findAll("button").find((b) => b.text() === nextLabel)
    expect(next).toBeDefined()
    clickAfterRender(next!.element as HTMLElement)

    await vi.waitFor(() => expect(api.list).toHaveBeenCalledTimes(2))
    expect(api.list).toHaveBeenLastCalledWith({ state: "active", page: 2, limit: 12 })
    expect(router.currentRoute.value.query.page).toBe("2")
  })
})
