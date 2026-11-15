import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { mount, type VueWrapper } from "@vue/test-utils"
import { nextTick } from "vue"
import { createPinia, setActivePinia } from "pinia"
import EventVotingSection from "@/components/events/EventVotingSection.vue"
import type { AppEvent } from "@/types/event"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({ eventApi: { vote: vi.fn(), cancelVote: vi.fn() } }))

const i18n = useRealI18n()
const NOW = Date.parse("2026-11-16T00:00:00.000Z")
const WAIT_MS = 60_000
const wrappers: VueWrapper[] = []

const votingEvent = (over: Partial<AppEvent> = {}): AppEvent => ({
  id: 8,
  type: "vote_monthly",
  media_type: "movie",
  period: "2026-11",
  title: { en: "Round", th: "รอบ" },
  body: { en: "", th: "" },
  image_url: null,
  link_url: null,
  starts_at: "2026-11-15T00:00:00.000Z",
  ends_at: "2026-11-20T00:00:00.000Z",
  show_popup: false,
  popup_priority: 0,
  state: "active",
  nominees: [
    { id: 30, event_id: 8, media_id: 301, media_type: "movie", position: 1, media_title: "Alpha", is_winner: false },
  ],
  vote_eligible_at: new Date(NOW + WAIT_MS).toISOString(),
  ...over,
})

const mountSection = (event: AppEvent) => {
  const wrapper = mount(EventVotingSection, { props: { event }, global: { plugins: [i18n, createPinia()] } })
  wrappers.push(wrapper)
  return wrapper
}

describe("EventVotingSection eligibility timer", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers({ now: NOW })
  })

  afterEach(() => {
    wrappers.forEach((w) => w.unmount())
    wrappers.length = 0
    vi.useRealTimers()
  })

  it("lifts the lock by itself when the eligibility time arrives", async () => {
    const wrapper = mountSection(votingEvent())
    const button = () => wrapper.find(".event-nominee-card__actions button").element as HTMLButtonElement
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    expect(button().disabled).toBe(true)

    await vi.advanceTimersByTimeAsync(WAIT_MS)
    await nextTick()

    expect(wrapper.find('[role="status"]').exists()).toBe(false)
    expect(button().disabled).toBe(false)
  })

  it("clears the timer on unmount", () => {
    const wrapper = mountSection(votingEvent())
    expect(vi.getTimerCount()).toBe(1)
    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })

  it("shows no eligibility notice for a past round", () => {
    const wrapper = mountSection(votingEvent({ state: "past" }))
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
  })
})
