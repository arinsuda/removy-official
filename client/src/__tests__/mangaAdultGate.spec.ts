import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import MangaView from "@/views/manga/MangaView.vue"
import { mangaApi } from "@/api/endpoints/manga"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/manga", () => ({
  mangaApi: { list: vi.fn(), search: vi.fn(), genres: vi.fn(), detail: vi.fn(), similar: vi.fn() },
}))

const i18n = useRealI18n()
const api = vi.mocked(mangaApi)
const wrappers: VueWrapper[] = []

const yearsAgo = (years: number) => `${new Date().getFullYear() - years}-01-01`

async function render(dateOfBirth: string | null) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = (dateOfBirth ? { id: 3, date_of_birth: dateOfBirth } : null) as never
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/manga", name: "manga", component: MangaView },
      { path: "/manga/:id(\\d+)", name: "manga-detail", component: { template: "<div />" } },
    ],
  })
  await router.push("/manga")
  await router.isReady()
  const w = mount(MangaView, {
    global: {
      plugins: [router, i18n, pinia, [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]],
    },
  })
  wrappers.push(w)
  await flushPromises()
  return { w, router }
}

/** Ticks the box the way a browser does, then fires change. */
async function toggleOn(w: VueWrapper) {
  const input = w.find(".adult-toggle input").element as HTMLInputElement
  input.checked = true
  input.dispatchEvent(new Event("change"))
  await flushPromises()
}

const confirmButton = () =>
  Array.from(document.body.querySelectorAll("button")).find((b) => b.textContent?.includes("I am 18 or older"))

describe("manga adult gate", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    i18n.global.locale.value = "en"
    api.list.mockResolvedValue({ data: { page: 1, results: [], total_pages: 1, total_results: 0 } } as never)
    api.genres.mockResolvedValue({ data: { genres: [] } } as never)
  })

  afterEach(() => {
    wrappers.splice(0).forEach((w) => w.unmount())
  })

  it("hides the 18+ toggle for guests", async () => {
    const { w } = await render(null)
    expect(w.find(".adult-toggle").exists()).toBe(false)
  })

  it("hides the 18+ toggle for signed-in viewers under 18", async () => {
    const { w } = await render(yearsAgo(10))
    expect(w.find(".adult-toggle").exists()).toBe(false)
  })

  it("shows the toggle for an 18+ viewer, and asks for confirmation before sending include_adult", async () => {
    const { w } = await render(yearsAgo(30))
    expect(w.find(".adult-toggle").exists()).toBe(true)
    expect(api.list).not.toHaveBeenLastCalledWith(expect.objectContaining({ include_adult: true }), 1)

    await toggleOn(w)
    await flushPromises()
    expect(document.body.textContent).toContain("You are 18 or older and want to see adult content?")
    expect(api.list.mock.calls.some(([query]) => (query as { include_adult?: boolean }).include_adult === true)).toBe(false)
    expect((w.find(".adult-toggle input").element as HTMLInputElement).checked).toBe(false)

    confirmButton()!.click()
    await vi.waitFor(() => expect(api.list).toHaveBeenLastCalledWith(expect.objectContaining({ include_adult: true }), 1))
    expect(localStorage.getItem("popcorn-manga-adult-ack:3")).toBe("1")
  })

  it("does not ask again once acknowledged", async () => {
    localStorage.setItem("popcorn-manga-adult-ack:3", "1")
    const { w } = await render(yearsAgo(30))
    await toggleOn(w)
    await flushPromises()
    expect(document.body.textContent).not.toContain("You are 18 or older")
    expect(api.list).toHaveBeenLastCalledWith(expect.objectContaining({ include_adult: true }), 1)
  })

  it("never sends include_adult from a shared URL alone", async () => {
    localStorage.clear()
    await render(yearsAgo(30))
    expect(api.list.mock.calls.every(([query]) => (query as { include_adult?: boolean }).include_adult !== true)).toBe(true)
  })
})
