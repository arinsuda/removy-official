/**
 * @vitest-environment happy-dom
 */

import { describe, it, expect, vi, beforeEach } from "vitest"
import { createPinia, setActivePinia } from "pinia"
import { mount } from "@vue/test-utils"
import { createRouter, createWebHistory } from "vue-router"
import ActorDetailView from "@/views/actor/ActorDetailView.vue"

// ── Mocks ────────────────────────────────────────────────────────────────────

vi.mock("@/api/api", () => ({
  actorApi: {
    getById: vi.fn(),
    getMovies: vi.fn(),
    getTVSeries: vi.fn(),
  },
}))

import { actorApi } from "@/api/api"

import { createI18n } from "vue-i18n"

const i18n = createI18n({
  legacy: false,
  locale: "en",
  messages: {
    en: {
      actor: {
        title: "Actor Details",
        knownFor: "Known For",
        born: "Born",
        died: "Died",
        age: "years old",
        agedAtDeath: "aged",
        biography: "Biography",
        noBiography: "No biography information available.",
        showMore: "Read More",
        showLess: "Show Less",
        workography: "Workography",
        moviesTab: "Movies",
        tvTab: "TV Series",
        asCharacter: "as {character}",
        unknownCharacter: "Uncredited Role",
        noMovieCredits: "No movie credits recorded yet.",
        noTVCredits: "No TV series credits recorded yet.",
        notFoundTitle: "Actor Not Found",
        notFoundDetail: "The requested actor could not be found.",
        errorTitle: "Failed to Load",
        errorDetail: "Unable to connect to server.",
        retry: "Retry",
        backToHome: "Back to Home",
        back: "Back",
      },
    },
  },
})

function createTestRouter() {
  return createRouter({
    history: createWebHistory(),
    routes: [
      { path: "/", name: "home", component: { template: "<div>Home</div>" } },
      { path: "/actors/:id", name: "actor-detail", component: ActorDetailView },
      { path: "/movies/:id", name: "movie-detail", component: { template: "<div>Movie</div>" } },
      { path: "/tv/:id", name: "tv-detail", component: { template: "<div>TV</div>" } },
    ],
  })
}

describe("ActorDetailView Component & Routing", () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
  })

  it("renders loading skeleton initially while requests are pending", async () => {
    vi.mocked(actorApi.getById).mockReturnValue(new Promise(() => {}))
    vi.mocked(actorApi.getMovies).mockReturnValue(new Promise(() => {}))
    vi.mocked(actorApi.getTVSeries).mockReturnValue(new Promise(() => {}))

    const router = createTestRouter()
    router.push("/actors/100")
    await router.isReady()

    const wrapper = mount(ActorDetailView, {
      global: {
        plugins: [pinia, router, i18n],
      },
    })

    expect(wrapper.find('[role="status"]').attributes("aria-busy")).toBe("true")
    expect(wrapper.find(".skeleton-block").exists()).toBe(true)
  })

  it("renders actor details, age, biography, and workography when loaded", async () => {
    vi.mocked(actorApi.getById).mockResolvedValueOnce({
      data: {
        id: 100,
        name: "Tom Holland",
        biography: "Thomas Stanley Holland is an English actor...",
        birthday: "1996-06-01",
        deathday: null,
        place_of_birth: "Kingston upon Thames, Surrey, England",
        profile_path: "/bP2F.jpg",
        known_for_department: "Acting",
        popularity: 85.4,
        gender: 2,
        imdb_id: "nm4043618",
        homepage: null,
      },
    } as Awaited<ReturnType<typeof actorApi.getById>>)

    vi.mocked(actorApi.getMovies).mockResolvedValueOnce({
      data: {
        cast: [
          {
            id: 315635,
            title: "Spider-Man: Homecoming",
            original_title: "Spider-Man: Homecoming",
            overview: "Peter Parker tries to balance high school life...",
            poster_path: "/555.jpg",
            backdrop_path: "/bg.jpg",
            release_date: "2017-07-05",
            vote_average: 7.4,
            vote_count: 20000,
            popularity: 90,
            genre_ids: [28, 12],
            adult: false,
            character: "Peter Parker / Spider-Man",
            credit_id: "c1",
          },
        ],
        crew: [],
      },
    } as Awaited<ReturnType<typeof actorApi.getMovies>>)

    vi.mocked(actorApi.getTVSeries).mockResolvedValueOnce({
      data: {
        cast: [],
        crew: [],
      },
    } as Awaited<ReturnType<typeof actorApi.getTVSeries>>)

    const router = createTestRouter()
    router.push("/actors/100")
    await router.isReady()

    const wrapper = mount(ActorDetailView, {
      global: {
        plugins: [pinia, router, i18n],
      },
    })

    await wrapper.vm.$nextTick()
    // Wait for promise resolution
    await new Promise((r) => setTimeout(r, 50))
    await wrapper.vm.$nextTick()

    expect(wrapper.find(".actor-name").text()).toBe("Tom Holland")
    expect(wrapper.find(".bio-text").text()).toContain("Thomas Stanley Holland")
    expect(wrapper.find(".media-title").text()).toBe("Spider-Man: Homecoming")
    expect(wrapper.find(".character-role").text()).toContain("as Peter Parker / Spider-Man")

    // Check RouterLink pointing to /movies/315635
    const movieLink = wrapper.find('a[href="/movies/315635"]')
    expect(movieLink.exists()).toBe(true)
  })

  it("handles deceased actor age calculation correctly", async () => {
    vi.mocked(actorApi.getById).mockResolvedValueOnce({
      data: {
        id: 200,
        name: "Chadwick Boseman",
        biography: "Chadwick Aaron Boseman was an American actor...",
        birthday: "1976-11-29",
        deathday: "2020-08-28",
        place_of_birth: "Anderson, South Carolina, USA",
        profile_path: "/cb.jpg",
        known_for_department: "Acting",
        popularity: 50,
        gender: 2,
        imdb_id: "nm1569276",
        homepage: null,
      },
    } as Awaited<ReturnType<typeof actorApi.getById>>)

    vi.mocked(actorApi.getMovies).mockResolvedValueOnce({
      data: { cast: [], crew: [] },
    } as Awaited<ReturnType<typeof actorApi.getMovies>>)
    vi.mocked(actorApi.getTVSeries).mockResolvedValueOnce({
      data: { cast: [], crew: [] },
    } as Awaited<ReturnType<typeof actorApi.getTVSeries>>)

    const router = createTestRouter()
    router.push("/actors/200")
    await router.isReady()

    const wrapper = mount(ActorDetailView, {
      global: {
        plugins: [pinia, router, i18n],
      },
    })

    await new Promise((r) => setTimeout(r, 50))
    await wrapper.vm.$nextTick()

    // 2020-08-28 minus 1976-11-29 = 43 years old at death
    expect(wrapper.find(".meta-pill.birth").text()).toContain("aged 43")
    expect(wrapper.find(".meta-pill.birth").text()).toContain("2020-08-28")
  })

  it("renders empty biography state when biography is null/empty", async () => {
    vi.mocked(actorApi.getById).mockResolvedValueOnce({
      data: {
        id: 300,
        name: "Unknown Actor",
        biography: "",
        birthday: null,
        deathday: null,
        place_of_birth: null,
        profile_path: null,
        known_for_department: "Acting",
        popularity: 1,
        gender: 0,
        imdb_id: null,
        homepage: null,
      },
    } as Awaited<ReturnType<typeof actorApi.getById>>)

    vi.mocked(actorApi.getMovies).mockResolvedValueOnce({
      data: { cast: [], crew: [] },
    } as Awaited<ReturnType<typeof actorApi.getMovies>>)
    vi.mocked(actorApi.getTVSeries).mockResolvedValueOnce({
      data: { cast: [], crew: [] },
    } as Awaited<ReturnType<typeof actorApi.getTVSeries>>)

    const router = createTestRouter()
    router.push("/actors/300")
    await router.isReady()

    const wrapper = mount(ActorDetailView, {
      global: {
        plugins: [pinia, router, i18n],
      },
    })

    await new Promise((r) => setTimeout(r, 50))
    await wrapper.vm.$nextTick()

    expect(wrapper.find(".bio-empty").text()).toBe("No biography information available.")
  })

  it("shows error state with Retry button on API error", async () => {
    vi.mocked(actorApi.getById).mockRejectedValueOnce(new Error("Network Error"))
    vi.mocked(actorApi.getMovies).mockRejectedValueOnce(new Error("Network Error"))
    vi.mocked(actorApi.getTVSeries).mockRejectedValueOnce(new Error("Network Error"))

    const router = createTestRouter()
    router.push("/actors/400")
    await router.isReady()

    const wrapper = mount(ActorDetailView, {
      global: {
        plugins: [pinia, router, i18n],
      },
    })

    await new Promise((r) => setTimeout(r, 50))
    await wrapper.vm.$nextTick()

    expect(wrapper.find(".error-state").exists()).toBe(true)
    expect(wrapper.text()).toContain("Failed to Load")
  })
})
