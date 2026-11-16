import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import { createPinia, setActivePinia } from "pinia"
import { AxiosError, AxiosHeaders } from "axios"
import MangaDetailView from "@/views/manga/MangaDetailView.vue"
import { mangaApi } from "@/api/endpoints/manga"
import { MANGA_ADULT_ACK_KEY_PREFIX } from "@/components/manga/mangaConfig"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/manga", () => ({
  mangaApi: { list: vi.fn(), search: vi.fn(), genres: vi.fn(), detail: vi.fn(), similar: vi.fn() },
}))
// The action bar and reviews mount with the detail page; their endpoints are stubbed here.
vi.mock("@/api/endpoints/media", () => ({
  mediaApi: { getMediaStats: vi.fn(async () => ({ data: { stats: null } })), likeMedia: vi.fn(), unlikeMedia: vi.fn() },
}))
vi.mock("@/api/endpoints/library", () => ({
  libraryApi: { getOwnMediaStatus: vi.fn(async () => ({ data: { in_lists: [] } })), addItem: vi.fn(), removeItem: vi.fn() },
}))
vi.mock("@/api/endpoints/social", () => ({
  reviewApi: {
    getMediaRating: vi.fn(async () => ({ data: { rating: { distribution: [] } } })),
    getMediaReviews: vi.fn(async () => ({ data: { reviews: [], pagination: { page: 1, page_size: 10, total: 0, has_more: false } } })),
    getMyReview: vi.fn(async () => ({ data: { review: null } })),
  },
}))

const i18n = useRealI18n()
const api = vi.mocked(mangaApi)
const wrappers: VueWrapper[] = []

const summary = (id: number, title: string) => ({
  id,
  media_type: "manga",
  titles: { romaji: title, english: "", native: "", thai: "" },
  display_title: title,
  cover_url: "https://cdn.example.com/c.jpg",
  banner_url: "",
  format: "manga",
  country: "JP",
  kind: "manga",
  is_adult: false,
  status: "RELEASING",
  start_year: 2019,
  chapters: 120,
  volumes: null,
  genres: ["Action", "Fantasy"],
  ratings: { anilist: { average: 88, available: true, scale: 100 } },
})

const detail = {
  ...summary(7, "Berserk Tale"),
  description: "First line<br><br>Second <i>line</i>",
  tags: [{ name: "Swords" }],
  read_links: [
    { site: "Official", url: "https://example.com/read", language: "EN" },
    { site: "Broken", url: "javascript:alert(1)", language: "" },
  ],
  relations: [],
}

/** A 403 the way the API sends it, so the view's error check reads the real shape. */
const ageRestricted = () =>
  new AxiosError("forbidden", "ERR_BAD_REQUEST", undefined, undefined, {
    status: 403,
    statusText: "Forbidden",
    headers: {},
    config: { headers: new AxiosHeaders() },
    data: { code: "AGE_RESTRICTED", error: "age restricted" },
  })

async function render(path = "/manga/7", dateOfBirth: string | null = null) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = (dateOfBirth ? { id: 1, date_of_birth: dateOfBirth } : null) as never
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/manga", name: "manga", component: { template: "<div />" } },
      { path: "/login", name: "login", component: { template: "<div />" } },
      { path: "/manga/:id(\\d+)", name: "manga-detail", component: MangaDetailView },
    ],
  })
  await router.push(path)
  await router.isReady()
  const w = mount(MangaDetailView, {
    global: {
      plugins: [router, i18n, pinia, [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]],
    },
  })
  wrappers.push(w)
  await flushPromises()
  return w
}

describe("MangaDetailView", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    i18n.global.locale.value = "en"
    api.detail.mockResolvedValue({ data: { manga: detail } } as never)
    api.similar.mockResolvedValue({ data: { results: [summary(9, "Similar One")] } } as never)
  })

  afterEach(() => {
    wrappers.splice(0).forEach((w) => w.unmount())
  })

  it("renders the title, facts and the AniList score out of 100", async () => {
    const w = await render()
    expect(api.detail).toHaveBeenCalledWith(7)
    await vi.waitFor(() => expect(w.find("h1").text()).toBe("Berserk Tale"))
    expect(w.text()).toContain("Manga")
    expect(w.text()).toContain("2019")
    expect(w.text()).toContain("120")
    expect(w.text()).toContain("88/100")
    expect(w.text()).toContain("Unknown")
  })

  it("shows the synopsis as plain text", async () => {
    const w = await render()
    await vi.waitFor(() => expect(w.find(".manga-detail__description").exists()).toBe(true))
    const text = w.find(".manga-detail__description").text()
    expect(text).toContain("Second line")
    expect(text).not.toContain("<i>")
  })

  it("offers only web read links, opened in a new tab without opener access", async () => {
    const w = await render()
    await vi.waitFor(() => expect(w.findAll(".manga-read__link")).toHaveLength(1))
    const link = w.find(".manga-read__link")
    expect(link.attributes("href")).toBe("https://example.com/read")
    expect(link.attributes("target")).toBe("_blank")
    expect(link.attributes("rel")).toBe("noopener noreferrer")
  })

  it("lists similar titles and the attribution", async () => {
    const w = await render()
    await vi.waitFor(() => expect(w.text()).toContain("Similar One"))
    expect(w.text()).toContain("Data provided by AniList")
  })

  it("explains the age restriction and links to the profile settings", async () => {
    api.detail.mockRejectedValue(ageRestricted())
    const w = await render()
    await vi.waitFor(() => expect(w.text()).toContain("This title is restricted"))
    expect(w.find('a[href="/settings#profile"]').exists()).toBe(true)
    expect(w.find("h1").exists()).toBe(false)
  })

  it("keeps an adult title behind the 18+ confirm until the viewer confirms", async () => {
    api.detail.mockResolvedValue({ data: { manga: { ...detail, is_adult: true } } } as never)
    const w = await render("/manga/7", "1990-01-01")
    await vi.waitFor(() => expect(document.body.textContent).toContain("You are 18 or older"))
    expect(w.find(".manga-detail__description").exists()).toBe(false)

    const confirm = Array.from(document.body.querySelectorAll("button")).find((b) => b.textContent?.includes("I am 18 or older"))
    confirm!.click()
    await vi.waitFor(() => expect(w.find(".manga-detail__description").exists()).toBe(true))
    expect(localStorage.getItem(`${MANGA_ADULT_ACK_KEY_PREFIX}1`)).toBe("1")
    expect(api.similar).toHaveBeenCalledWith(7, true)
  })

  it("shows an error state with retry when the detail fails", async () => {
    api.detail.mockRejectedValue(new Error("boom"))
    const w = await render()
    await vi.waitFor(() => expect(w.text()).toContain("Could not load this title"))
  })
})
