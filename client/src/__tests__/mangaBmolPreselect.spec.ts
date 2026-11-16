import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import MangaBmolSheet from "@/components/manga/MangaBmolSheet.vue"
import type { MangaItem } from "@/types/manga"
import { clickAfterRender, waitForEnabled } from "./helpers/waitForEnabled"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ getUserBMOL: vi.fn(), addItem: vi.fn() }))

vi.mock("@/api/api", () => ({
  bmolApi: m,
  movieApi: { search: vi.fn(), searchSeries: vi.fn() },
  libraryApi: { getVisibleUserLibrary: vi.fn() },
  reviewApi: { getUserReviews: vi.fn() },
}))
vi.mock("@/api/endpoints/manga", () => ({ mangaApi: { search: vi.fn() } }))

const i18n = useRealI18n()
const VIEWED_MANGA_ID = 7
const manga = {
  id: VIEWED_MANGA_ID,
  media_type: "manga",
  display_title: "Viewed Manga",
  cover_url: "https://cdn.example/cover.jpg",
  start_year: 2019,
  ratings: { anilist: { available: true, average: 80 } },
} as MangaItem

const wrappers: VueWrapper[] = []
const SELECTED_ROW = ".add-sheet__row.is-selected"

beforeEach(() => {
  i18n.global.locale.value = "en"
  m.getUserBMOL.mockResolvedValue({ data: { items: [] } })
  m.addItem.mockResolvedValue({ data: { item: {} } })
})
afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
})

function mountSheet(items: unknown[] = []) {
  setActivePinia(createPinia())
  m.getUserBMOL.mockResolvedValue({ data: { items } })
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const w = mount(MangaBmolSheet, {
    props: { modelValue: true, userId: 1, manga },
    global: { plugins: [i18n, [VueQueryPlugin, { queryClient: client }]] },
    attachTo: document.body,
  })
  wrappers.push(w)
  return w
}

describe("MangaBmolSheet preselection", () => {
  it("shows the viewed manga already selected, without a search", async () => {
    mountSheet()
    await flushPromises()
    const row = document.body.querySelector(SELECTED_ROW)
    expect(row?.textContent).toContain("Viewed Manga")
  })

  it("adds the viewed manga at the rank the user keeps", async () => {
    mountSheet()
    const submit = await waitForEnabled(".add-sheet__footer button")
    clickAfterRender(submit)
    await vi.waitFor(() => expect(m.addItem).toHaveBeenCalledWith({ media_id: VIEWED_MANGA_ID, media_type: "manga", rank: 1 }))
  })

  it("does not re-add a manga that is already ranked", async () => {
    mountSheet([{ id: 1, rank: 1, note: "", media_type: "manga", created_at: "", media: { id: VIEWED_MANGA_ID, title: "Viewed Manga", media_type: "manga" } }])
    await flushPromises()
    // The ranked title shows as disabled; submitting sends nothing for it.
    expect(document.body.querySelector(".add-sheet__row.is-disabled")?.textContent).toContain("Viewed Manga")
    clickAfterRender(await waitForEnabled(".add-sheet__footer button"))
    await flushPromises()
    expect(m.addItem).not.toHaveBeenCalled()
  })
})
