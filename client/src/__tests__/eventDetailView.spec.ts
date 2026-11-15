import { afterEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createMemoryHistory, createRouter } from "vue-router"
import { createPinia, setActivePinia } from "pinia"
import EventDetailView from "@/views/events/EventDetailView.vue"
import { eventApi } from "@/api/endpoints/event"
import { useAuthStore } from "@/stores/auth"
import type { AppEvent } from "@/types/event"
import { clickAfterRender } from "./helpers/waitForEnabled"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({
  eventApi: {
    list: vi.fn(),
    get: vi.fn(),
    popup: vi.fn(),
    vote: vi.fn(),
    cancelVote: vi.fn(),
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

async function mountAt(path: string, pinia = createPinia()) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/events", name: "events", component: { template: "<div>Events</div>" } },
      { path: "/events/:id", name: "event-detail", component: EventDetailView },
    ],
  })
  await router.push(path)
  await router.isReady()
  const wrapper = mount(EventDetailView, { global: { plugins: [i18n, router, pinia] } })
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

  it("renders nominees and handles voting on active voting events", async () => {
    const voteEvent = makeEvent({
      id: 5,
      type: "vote_monthly",
      media_type: "movie",
      period: "2026-11",
      state: "active",
      nominees: [
        {
          id: 10,
          event_id: 5,
          media_id: 101,
          media_type: "movie",
          position: 1,
          media_title: "Dune Part Two",
          poster_path: "/dune.jpg",
          is_winner: false,
        },
        {
          id: 11,
          event_id: 5,
          media_id: 102,
          media_type: "movie",
          position: 2,
          media_title: "Oppenheimer",
          poster_path: "/oppen.jpg",
          is_winner: false,
        },
      ],
    })

    api.get.mockResolvedValueOnce({ data: { event: voteEvent } } as never)

    const { wrapper } = await mountAt("/events/5")
    expect(wrapper.text()).toContain("Dune Part Two")
    expect(wrapper.text()).toContain("Oppenheimer")
    expect(wrapper.text()).toContain(i18n.global.t("events.voting.title"))

    // Nominee cards should be rendered
    const nomineeCards = wrapper.findAll(".event-nominee-card")
    expect(nomineeCards.length).toBe(2)
    expect(wrapper.find("img.event-nominee-card__poster").attributes("src")).toBe(
      "https://image.tmdb.org/t/p/w342/dune.jpg",
    )
  })

  it("locks every vote control while a vote is in flight", async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    useAuthStore().user = { id: 7 } as never
    api.vote.mockReturnValueOnce(new Promise(() => {}) as never)
    const voteEvent = makeEvent({
      id: 8,
      type: "vote_monthly",
      state: "active",
      my_vote_nominee_id: null,
      nominees: [
        { id: 30, event_id: 8, media_id: 301, media_type: "movie", position: 1, media_title: "Alpha", is_winner: false },
        { id: 31, event_id: 8, media_id: 302, media_type: "movie", position: 2, media_title: "Beta", is_winner: false },
      ],
    })
    api.get.mockResolvedValueOnce({ data: { event: voteEvent } } as never)

    const { wrapper } = await mountAt("/events/8", pinia)
    const voteButtons = () => wrapper.findAll(".event-nominee-card__actions button")
    expect(voteButtons().length).toBe(2)
    clickAfterRender(voteButtons()[0].element as HTMLElement)
    await flushPromises()

    expect(api.vote).toHaveBeenCalledTimes(1)
    voteButtons().forEach((b) => expect(b.attributes("disabled")).toBeDefined())
  })

  it("renders winner badges and vote results when voting event is past", async () => {
    const pastVoteEvent = makeEvent({
      id: 6,
      type: "vote_monthly",
      media_type: "movie",
      period: "2026-10",
      state: "past",
      total_votes: 150,
      is_finalized: true,
      nominees: [
        {
          id: 20,
          event_id: 6,
          media_id: 201,
          media_type: "movie",
          position: 1,
          media_title: "Past Winner Movie",
          vote_count: 100,
          percentage: 66.67,
          is_winner: true,
        },
        {
          id: 21,
          event_id: 6,
          media_id: 202,
          media_type: "movie",
          position: 2,
          media_title: "Runner Up Movie",
          vote_count: 50,
          percentage: 33.33,
          is_winner: false,
        },
      ],
    })

    api.get.mockResolvedValueOnce({ data: { event: pastVoteEvent } } as never)
    // English bundle so the plural and percent assertions read as plain text.
    i18n.global.locale.value = "en"

    const { wrapper } = await mountAt("/events/6")
    expect(wrapper.text()).toContain("Past Winner Movie")
    expect(wrapper.text()).toContain("Runner Up Movie")
    expect(wrapper.find(".event-nominee-card__winner-badge").exists()).toBe(true)
    expect(wrapper.text()).toContain(i18n.global.t("events.voting.winner"))
    expect(wrapper.text()).toContain("Total votes: 150")
    expect(wrapper.text()).toContain("100 votes (67%)")
    expect(wrapper.text()).toContain("50 votes (33%)")
  })
})
