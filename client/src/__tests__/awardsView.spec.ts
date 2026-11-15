import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import AwardsView from "@/views/events/AwardsView.vue"
import { eventApi } from "@/api/endpoints/event"
import type { AwardsResponse, EventAward } from "@/types/event"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({
  eventApi: { getAwards: vi.fn() },
}))

const i18n = useRealI18n()
const api = vi.mocked(eventApi)
const wrappers: VueWrapper[] = []

const winner = (over: Partial<EventAward>): EventAward => ({
  event_id: 1,
  kind: "monthly",
  period: "2026-11",
  media_id: 1,
  media_type: "movie",
  media_title: "Winner",
  poster_path: null,
  backdrop_path: null,
  votes: 4,
  ...over,
})

const response = (awards: EventAward[], years: number[] = [2026]): { data: AwardsResponse } => ({
  data: { awards, years },
})

function mountView() {
  const w = mount(AwardsView, {
    global: {
      plugins: [
        i18n,
        [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }],
      ],
      stubs: { RouterLink: { template: "<a><slot /></a>" } },
    },
    attachTo: document.body,
  })
  wrappers.push(w)
  return w
}

const tabs = () => [...document.body.querySelectorAll<HTMLElement>("[role=tab]")]

beforeEach(() => {
  i18n.global.locale.value = "en"
  api.getAwards.mockResolvedValue(response([]))
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  vi.clearAllMocks()
})

describe("AwardsView", () => {
  it("shows the empty state when there are no winners", async () => {
    mountView()
    await flushPromises()
    expect(document.body.textContent).toContain("No awards yet")
  })

  it("lists yearly winners first, then monthly rounds newest first, grouping co-winners", async () => {
    api.getAwards.mockResolvedValue(
      response([
        winner({ media_id: 1, media_title: "Monthly Old", kind: "monthly", period: "2026-09" }),
        winner({ media_id: 2, media_title: "Monthly A", kind: "monthly", period: "2026-11" }),
        winner({ media_id: 3, media_title: "Monthly B", kind: "monthly", period: "2026-11" }),
        winner({ media_id: 4, media_title: "Yearly Best", kind: "yearly", period: "2026" }),
      ]),
    )
    const w = mountView()
    await flushPromises()

    const titles = w.findAll(".awards-group__title").map((h) => h.text())
    expect(titles).toEqual(["Yearly winners · 2026", "Monthly winners · November 2026", "Monthly winners · September 2026"])
    expect(w.text()).toContain("Yearly Best")
    // Two winners share the November round, so both are marked as co-winners.
    expect(w.findAll(".award-card__co-winner")).toHaveLength(2)
    expect(w.text()).toContain("4 votes")
  })

  it("requests the default media type and year, then switches to TV", async () => {
    mountView()
    await flushPromises()
    expect(api.getAwards).toHaveBeenLastCalledWith(expect.objectContaining({ media_type: "movie" }))

    tabs()[1]!.click()
    await flushPromises()
    expect(api.getAwards).toHaveBeenLastCalledWith(expect.objectContaining({ media_type: "tv" }))
  })

  it("moves to the newest year that has winners when the selected year has none", async () => {
    api.getAwards
      .mockResolvedValueOnce(response([], [1999]))
      .mockResolvedValueOnce(response([winner({ kind: "yearly", period: "1999", media_title: "Old Classic" })], [1999]))
    const w = mountView()
    await flushPromises()

    expect(api.getAwards).toHaveBeenLastCalledWith(expect.objectContaining({ year: 1999 }))
    expect(w.text()).toContain("Old Classic")
  })

  it("shows the error state with a retry", async () => {
    api.getAwards.mockRejectedValueOnce(new Error("boom"))
    mountView()
    await flushPromises()

    expect(document.body.textContent).toContain("Something went wrong while loading awards")
    expect(document.body.querySelector(".app-state")).not.toBeNull()
  })
})
