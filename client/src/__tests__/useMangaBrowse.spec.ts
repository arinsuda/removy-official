import { describe, expect, it, vi, beforeEach } from "vitest"
import { mount } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { ref, type Ref } from "vue"
import { mangaApi } from "@/api/endpoints/manga"
import { DEFAULT_MANGA_STATE, type MangaBrowseState } from "@/components/manga/mangaQuery"
import { useMangaBrowse } from "@/composables/useMangaBrowse"

vi.mock("@/api/endpoints/manga", () => ({
  mangaApi: { list: vi.fn(), search: vi.fn(), genres: vi.fn(), detail: vi.fn(), similar: vi.fn() },
}))

const item = (id: number) => ({
  id,
  media_type: "manga",
  titles: { romaji: `Title ${id}`, english: "", native: "", thai: "" },
  display_title: `Title ${id}`,
  cover_url: "https://cdn.example.com/c.jpg",
  banner_url: "",
  format: "manga",
  country: "JP",
  kind: "manga",
  status: "RELEASING",
  start_year: 2020,
  chapters: 10,
  volumes: null,
  genres: [],
  ratings: { anilist: { average: 80, available: true, scale: 100 } },
})

const page = (results: unknown[], pageNo = 1, totalPages = 1) => ({
  data: { page: pageNo, results, total_pages: totalPages, total_results: results.length },
})

/** Mounts the composable inside a query client, the way a page does. */
function setup(initial: MangaBrowseState = DEFAULT_MANGA_STATE) {
  const state: Ref<MangaBrowseState> = ref(initial)
  const includeAdult = ref(false)
  let data!: ReturnType<typeof useMangaBrowse>
  mount(
    {
      setup() {
        data = useMangaBrowse(state, includeAdult)
        return () => null
      },
    },
    {
      global: {
        plugins: [[VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]],
      },
    },
  )
  return { state, includeAdult, data: () => data }
}

describe("useMangaBrowse", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("loads the first page with the default filters and keeps the cover URL absolute", async () => {
    vi.mocked(mangaApi.list).mockResolvedValue(page([item(1)]) as never)
    const { data } = setup()

    await vi.waitFor(() => expect(data().loaded.value).toHaveLength(1))
    expect(mangaApi.list).toHaveBeenCalledWith({ country: "JP", format: "all", sort: "popular", status: "all", include_adult: false }, 1)
    expect(data().loaded.value[0]).toMatchObject({ id: 1, display_title: "Title 1", cover_url: "https://cdn.example.com/c.jpg" })
  })

  it("starts a new list with the new filters when the URL state changes", async () => {
    vi.mocked(mangaApi.list).mockResolvedValue(page([item(1)]) as never)
    const { state } = setup()
    await vi.waitFor(() => expect(mangaApi.list).toHaveBeenCalledTimes(1))

    state.value = { ...DEFAULT_MANGA_STATE, country: "KR", format: "novel", genre: "Action" }
    await vi.waitFor(() =>
      expect(mangaApi.list).toHaveBeenLastCalledWith(
        { country: "KR", format: "novel", sort: "popular", status: "all", genre: "Action", include_adult: false },
        1,
      ),
    )
  })

  it("fetches the next page while more pages exist and drops repeated titles", async () => {
    vi.mocked(mangaApi.list)
      .mockResolvedValueOnce(page([item(1), item(2)], 1, 2) as never)
      .mockResolvedValueOnce(page([item(2), item(3)], 2, 2) as never)
    const { data } = setup()
    await vi.waitFor(() => expect(data().hasNextPage.value).toBe(true))

    await data().fetchNextPage()
    await vi.waitFor(() => expect(data().loaded.value.map((m) => m.id)).toEqual([1, 2, 3]))
    expect(mangaApi.list).toHaveBeenLastCalledWith(expect.any(Object), 2)
  })

  it("offers the genre list from the API", async () => {
    vi.mocked(mangaApi.list).mockResolvedValue(page([]) as never)
    vi.mocked(mangaApi.genres).mockResolvedValue({ data: { genres: ["Action", "Romance"] } } as never)
    const { data } = setup()

    await vi.waitFor(() => expect(data().genres.value).toEqual(["Action", "Romance"]))
    expect(mangaApi.genres).toHaveBeenCalledTimes(1)
  })
})
