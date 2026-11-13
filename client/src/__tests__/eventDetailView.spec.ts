import { afterEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createMemoryHistory, createRouter } from "vue-router"
import EventDetailView from "@/views/events/EventDetailView.vue"
import { eventApi } from "@/api/endpoints/event"
import type { AppEvent } from "@/types/event"
import { clickAfterRender } from "./helpers/waitForEnabled"
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
  title: { en: "Detail Event", th: "กิจกรรมรายละเอียด" },
  body: { en: "Detail body", th: "เนื้อหารายละเอียด" },
  image_url: null,
  link_url: "/movies",
  starts_at: "2026-11-15T00:00:00.000Z",
  ends_at: "2026-11-20T00:00:00.000Z",
  show_popup: false,
  popup_priority: 0,
  state: "active",
  ...over,
})

// The view branches on axios error status, so the fake carries the same shape.
const httpError = (status: number) =>
  Object.assign(new Error(`HTTP ${status}`), { isAxiosError: true, response: { status } })

async function mountAt(path: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/events", name: "events", component: { template: "<div>Events</div>" } },
      { path: "/events/:id", name: "event-detail", component: EventDetailView },
    ],
  })
  await router.push(path)
  await router.isReady()
  const wrapper = mount(EventDetailView, { global: { plugins: [i18n, router] } })
  wrappers.push(wrapper)
  await flushPromises()
  return { wrapper, router }
}

describe("EventDetailView", () => {
  afterEach(() => {
    wrappers.forEach((w) => w.unmount())
    wrappers.length = 0
    vi.clearAllMocks()
  })

  it("renders the formatted event period", async () => {
    api.get.mockResolvedValueOnce({ data: { event: makeEvent() } } as never)

    const { wrapper } = await mountAt("/events/1")

    const time = wrapper.find("time")
    expect(time.exists()).toBe(true)
    expect(time.text()).toContain("–")
    expect(time.text()).not.toContain("Invalid")
  })

  it("shows the not-found state with a back link only on 404", async () => {
    api.get.mockRejectedValueOnce(httpError(404))

    const { wrapper } = await mountAt("/events/99")

    expect(wrapper.text()).toContain(i18n.global.t("events.detail.notFoundTitle"))
    expect(wrapper.text()).not.toContain(i18n.global.t("events.loadErrorDesc"))
    expect(wrapper.find('a[href="/events"]').exists()).toBe(true)
  })

  it("shows a generic error with retry for other failures, then recovers", async () => {
    api.get
      .mockRejectedValueOnce(httpError(500))
      .mockResolvedValueOnce({ data: { event: makeEvent() } } as never)

    const { wrapper } = await mountAt("/events/1")

    expect(wrapper.text()).toContain(i18n.global.t("events.loadErrorDesc"))
    expect(wrapper.text()).not.toContain(i18n.global.t("events.detail.notFoundTitle"))

    const retryLabel = i18n.global.t("common.retry")
    const retry = wrapper.findAll("button").find((b) => b.text() === retryLabel)
    expect(retry).toBeDefined()
    clickAfterRender(retry!.element as HTMLElement)

    await vi.waitFor(() => expect(api.get).toHaveBeenCalledTimes(2))
    await flushPromises()
    expect(wrapper.find("time").exists()).toBe(true)
  })

  it("refetches when the route id changes", async () => {
    api.get
      .mockResolvedValueOnce({ data: { event: makeEvent({ id: 1 }) } } as never)
      .mockResolvedValueOnce({ data: { event: makeEvent({ id: 2 }) } } as never)

    const { router } = await mountAt("/events/1")
    await router.push("/events/2")

    await vi.waitFor(() => expect(api.get).toHaveBeenLastCalledWith(2))
  })
})
