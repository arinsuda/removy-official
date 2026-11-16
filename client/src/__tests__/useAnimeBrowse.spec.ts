import { describe, expect, it, vi, beforeEach } from "vitest"
import { mount } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { ref, type Ref } from "vue"
import { animeApi } from "@/api/endpoints/anime"
import { movieApi } from "@/api/endpoints/movie"
import { DEFAULT_ANIME_STATE, type AnimeBrowseState } from "@/components/anime/animeQuery"
import { useAnimeBrowse } from "@/composables/useAnimeBrowse"

vi.mock("@/api/endpoints/anime", () => ({ animeApi: { list: vi.fn() } }))
vi.mock("@/api/endpoints/movie", () => ({
  actorApi: {},
  movieApi: { getGenres: vi.fn(), getSeriesGenres: vi.fn() },
}))

const page = (results: unknown[]) => ({ data: { page: 1, results, total_pages: 1, total_results: results.length } })
const tvItem = (id: number) => ({
  id,
  name: `Show ${id}`,
  original_name: "",
  overview: "",
  poster_path: "/p.jpg",
  backdrop_path: "",
  first_air_date: "2001-01-01",
  vote_average: 7,
  vote_count: 0,
  popularity: 0,
  genre_ids: [],
  adult: false,
})

/** Mounts the composable inside a query client, the way a page does. */
function setup(initial: AnimeBrowseState = DEFAULT_ANIME_STATE) {
  const state: Ref<AnimeBrowseState> = ref(initial)
  let data!: ReturnType<typeof useAnimeBrowse>
  mount(
    {
      setup() {
        data = useAnimeBrowse(state)
        return () => null
      },
    },
    {
      global: {
        plugins: [[VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]],
      },
    },
  )
  return { state, data: () => data }
}

describe("useAnimeBrowse", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("loads the first page with the default filters and maps the titles", async () => {
    vi.mocked(animeApi.list).mockResolvedValue(page([tvItem(1)]) as never)
    const { data } = setup()

    await vi.waitFor(() => expect(data().loaded.value).toHaveLength(1))
    expect(animeApi.list).toHaveBeenCalledWith("tv", { sort: "popular", country: "JP" }, 1)
    expect(data().loaded.value[0]).toMatchObject({ key: "tv-1", type: "tv", title: "Show 1" })
  })

  it("starts a new list with the new filters when the URL state changes", async () => {
    vi.mocked(animeApi.list).mockResolvedValue(page([tvItem(1)]) as never)
    const { state } = setup()
    await vi.waitFor(() => expect(animeApi.list).toHaveBeenCalledTimes(1))

    state.value = { ...DEFAULT_ANIME_STATE, type: "movie", sort: "newest", keyword: "isekai" }
    await vi.waitFor(() =>
      expect(animeApi.list).toHaveBeenLastCalledWith("movie", { sort: "newest", country: "JP", keyword: "isekai" }, 1),
    )
  })

  it("offers genres from the TV list without the Animation genre", async () => {
    vi.mocked(movieApi.getSeriesGenres).mockResolvedValue({
      data: {
        genres: [
          { id: 16, name: "Animation" },
          { id: 18, name: "Drama" },
        ],
      },
    } as never)
    const { data } = setup()

    await vi.waitFor(() => expect(data().genres.value).toEqual([{ id: 18, name: "Drama" }]))
    expect(movieApi.getSeriesGenres).toHaveBeenCalledTimes(1)
    expect(movieApi.getGenres).not.toHaveBeenCalled()
  })
})
