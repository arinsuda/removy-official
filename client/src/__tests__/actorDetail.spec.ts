/**
 * @vitest-environment happy-dom
 */
import { beforeEach, describe, expect, it, vi } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { flushPromises, mount } from "@vue/test-utils"
import { createMemoryHistory, createRouter } from "vue-router"
import ActorDetailView from "@/views/actor/ActorDetailView.vue"
import { i18n } from "@/i18n"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", () => ({
  actorApi: { getById: vi.fn(), getMovies: vi.fn(), getTVSeries: vi.fn() },
}))

import { actorApi } from "@/api/api"

useRealI18n()

const stub = { template: "<div />" }
const person = (over: Record<string, unknown> = {}) => ({
  id: 100,
  name: "Tom Holland",
  biography: "Thomas Stanley Holland is an English actor.",
  birthday: "1996-06-01",
  deathday: null,
  place_of_birth: "Kingston upon Thames, England",
  profile_path: "/p.jpg",
  known_for_department: "Acting",
  popularity: 85,
  gender: 2,
  imdb_id: "nm4043618",
  homepage: null,
  ...over,
})
const movie = (id: number, title: string, date: string, popularity: number, character?: string) => ({
  id,
  title,
  original_title: title,
  overview: "",
  poster_path: `/m${id}.jpg`,
  backdrop_path: "",
  release_date: date,
  vote_average: 7,
  vote_count: 100,
  popularity,
  genre_ids: [],
  adult: false,
  character,
})
const show = (id: number, name: string, date: string, popularity: number, character?: string) => ({
  id,
  name,
  original_name: name,
  overview: "",
  poster_path: `/t${id}.jpg`,
  backdrop_path: "",
  first_air_date: date,
  vote_average: 7,
  vote_count: 100,
  popularity,
  genre_ids: [],
  adult: false,
  character,
})

function mockApi(opts: { person?: unknown; movies?: { cast: unknown[]; crew: unknown[] }; tv?: { cast: unknown[]; crew: unknown[] } }) {
  vi.mocked(actorApi.getById).mockResolvedValue({ data: opts.person ?? person() } as never)
  vi.mocked(actorApi.getMovies).mockResolvedValue({ data: opts.movies ?? { cast: [], crew: [] } } as never)
  vi.mocked(actorApi.getTVSeries).mockResolvedValue({ data: opts.tv ?? { cast: [], crew: [] } } as never)
}

async function mountView(id = "100") {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/", component: stub },
      { path: "/actors/:id", component: ActorDetailView },
      { name: "movie-detail", path: "/movies/:id", component: stub },
      { name: "tv-detail", path: "/tv/:id", component: stub },
    ],
  })
  router.push(`/actors/${id}`)
  await router.isReady()
  const wrapper = mount(ActorDetailView, { global: { plugins: [createPinia(), router, i18n] } })
  await flushPromises()
  return wrapper
}

describe("ActorDetailView", () => {
  beforeEach(() => {
    i18n.global.locale.value = "en"
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it("shows the skeleton while loading", async () => {
    vi.mocked(actorApi.getById).mockReturnValue(new Promise(() => {}))
    vi.mocked(actorApi.getMovies).mockReturnValue(new Promise(() => {}))
    vi.mocked(actorApi.getTVSeries).mockReturnValue(new Promise(() => {}))
    const wrapper = await mountView()
    expect(wrapper.find('[aria-busy="true"]').exists()).toBe(true)
  })

  it("renders hero, stats, known-for rail and filmography", async () => {
    mockApi({
      movies: { cast: [movie(1, "Homecoming", "2017-07-05", 90, "Peter Parker"), movie(2, "Cherry", "2021-02-26", 40, "Cherry")], crew: [] },
      tv: { cast: [show(3, "The Crowded Room", "2023-06-09", 60, "Danny")], crew: [] },
    })
    const wrapper = await mountView()
    expect(wrapper.find("h1").text()).toBe("Tom Holland")
    expect(wrapper.text()).toContain("Thomas Stanley Holland")
    const stats = wrapper.find("dl").text()
    expect(stats).toContain("3")
    expect(stats).toContain("2017 – 2023")
    expect(wrapper.text()).toContain("Known for")
    expect(wrapper.findAll('a[href="/movies/1"]').length).toBeGreaterThan(0)
    expect(wrapper.find('a[href="https://www.imdb.com/name/nm4043618"]').exists()).toBe(true)
  })

  it("sorts by newest and filters to TV series", async () => {
    mockApi({
      movies: { cast: [movie(1, "Old Movie", "2001-01-01", 99), movie(2, "New Movie", "2022-01-01", 1)], crew: [] },
      tv: { cast: [show(3, "A Show", "2010-01-01", 50)], crew: [] },
    })
    const wrapper = await mountView()
    const titles = () => wrapper.findAll(".credit-grid .poster-title").map((n) => n.text())
    expect(titles()).toEqual(["Old Movie", "A Show", "New Movie"])

    await wrapper.find("select").setValue("newest")
    expect(titles()).toEqual(["New Movie", "A Show", "Old Movie"])

    const filterTabs = wrapper.findAll('[role="tablist"]')[1]!
    await filterTabs.findAll('[role="tab"]')[2]!.trigger("click")
    expect(titles()).toEqual(["A Show"])
  })

  it("switches to the timeline grouped by year", async () => {
    mockApi({ movies: { cast: [movie(1, "Alpha", "2001-01-01", 5), movie(2, "Beta", "2001-06-01", 9), movie(3, "Gamma", "2005-01-01", 1)], crew: [] } })
    const wrapper = await mountView()
    const viewTabs = wrapper.findAll('[role="tablist"]')[2]!
    await viewTabs.findAll('[role="tab"]')[1]!.trigger("click")
    const years = wrapper.findAll(".timeline__label").map((n) => n.text())
    expect(years).toEqual(["2005", "2001"])
    expect(wrapper.findAll(".timeline__row")).toHaveLength(3)
  })

  it("opens on the crew tab for a person without acting credits", async () => {
    mockApi({ person: person({ known_for_department: "Directing" }), movies: { cast: [], crew: [movie(9, "Directed Film", "2015-01-01", 10)] } })
    const wrapper = await mountView()
    const selected = wrapper.findAll('[role="tablist"]')[0]!.find('[aria-selected="true"]')
    expect(selected.text()).toContain("Crew")
    expect(wrapper.text()).toContain("Directed Film")
  })

  it("shows the empty filmography state", async () => {
    mockApi({})
    const wrapper = await mountView()
    expect(wrapper.text()).toContain("No acting credits yet.")
  })

  it("formats a deceased person's dates and age", async () => {
    mockApi({ person: person({ birthday: "1976-11-29", deathday: "2020-08-28" }) })
    const wrapper = await mountView()
    expect(wrapper.find(".actor-hero__facts").text()).toContain("aged 43")
  })

  it("shows the not-found state on 404 and an error state with retry otherwise", async () => {
    vi.mocked(actorApi.getById).mockRejectedValue({ response: { status: 404 } })
    vi.mocked(actorApi.getMovies).mockRejectedValue({ response: { status: 404 } })
    vi.mocked(actorApi.getTVSeries).mockRejectedValue({ response: { status: 404 } })
    let wrapper = await mountView()
    expect(wrapper.text()).toContain("Actor not found")

    vi.spyOn(console, "error").mockImplementation(() => {})
    vi.mocked(actorApi.getById).mockRejectedValue({ response: { status: 500 } })
    wrapper = await mountView()
    expect(wrapper.find('[role="alert"]').text()).toContain("Could not load this actor")
    mockApi({})
    await wrapper.find('[role="alert"] button').trigger("click")
    await flushPromises()
    expect(wrapper.find("h1").text()).toBe("Tom Holland")
  })
})
