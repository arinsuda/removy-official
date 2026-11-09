import { beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter } from "vue-router"
import BmolBoard from "@/components/bmol/BmolBoard.vue"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import type { BMOLItemResponse } from "@/types"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ getUserBMOL: vi.fn(), updateItem: vi.fn(), addItem: vi.fn(), removeItem: vi.fn() }))

vi.mock("@/api/api", () => ({
  bmolApi: m,
  movieApi: { search: vi.fn(), searchSeries: vi.fn() },
  libraryApi: { getVisibleUserLibrary: vi.fn() },
  reviewApi: { getUserReviews: vi.fn() },
}))
vi.mock("html-to-image", () => ({ toPng: vi.fn() }))

const i18n = useRealI18n()
const stub = (name: string) => defineComponent({ name, render: () => h("div", { "data-test": name }) })
const stubs = { BmolAddSheet: stub("BmolAddSheet"), BolShareSheet: stub("BolShareSheet"), ConfirmModal: stub("ConfirmModal"), BmolNoteEditor: stub("BmolNoteEditor") }

const item = (id: number, rank: number, title: string): BMOLItemResponse => ({
  id, rank, note: "", media_type: "movie", created_at: "",
  media: { id: id * 10, title, poster_url: "", media_type: "movie", genres: [], vote_average: 7, created_at: "" },
})
const httpError = (status: number) => ({ isAxiosError: true, response: { status } })
const ranks = (w: ReturnType<typeof mount>) => w.findAll(".bcard").map((c) => `${c.find(".bcard__title").text()}:${c.find(".bcard__rank").text()}`)

async function mountBoard(viewerId: number, ownerId = 1) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: viewerId } as never
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: "/", component: { render: () => null } }, { path: "/:p(.*)*", component: { render: () => null } }] })
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = mount(BmolBoard, {
    props: { userId: ownerId },
    global: { plugins: [i18n, pinia, router, [VueQueryPlugin, { queryClient: client }]], stubs },
  })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  vi.clearAllMocks()
  m.getUserBMOL.mockResolvedValue({ data: { items: [item(1, 1, "Alpha"), item(2, 2, "Beta"), item(3, 3, "Gamma")] } })
})

describe("BmolBoard controls", () => {
  it("gives the owner add, edit and share controls", async () => {
    const w = await mountBoard(1)
    const text = w.text()
    expect(text).toContain(i18n.global.t("library.bmol.actions.add"))
    expect(text).toContain(i18n.global.t("library.bmol.actions.edit"))
    expect(text).toContain(i18n.global.t("library.bmol.actions.share"))
  })

  it("gives a visitor a read-only board", async () => {
    const w = await mountBoard(2)
    expect(ranks(w)).toHaveLength(3)
    expect(w.text()).not.toContain(i18n.global.t("library.bmol.actions.add"))
    expect(w.text()).not.toContain(i18n.global.t("library.bmol.actions.edit"))
    expect(w.find(".bcard__handle").exists()).toBe(false)
    expect(w.find(".bcard__menu-btn").exists()).toBe(false)
  })

  it("shows the locked state on 403", async () => {
    m.getUserBMOL.mockRejectedValue(httpError(403))
    const w = await mountBoard(2)
    expect(w.text()).toContain(i18n.global.t("library.bmol.locked.title"))
    expect(w.find("[role=tablist]").exists()).toBe(false)
  })

  it("invites the owner to pick a number one when empty", async () => {
    m.getUserBMOL.mockResolvedValue({ data: { items: [] } })
    const w = await mountBoard(1)
    expect(w.text()).toContain(i18n.global.t("library.bmol.onboarding.title"))
    const visitor = await mountBoard(2)
    expect(visitor.text()).toContain(i18n.global.t("library.bmol.empty.title"))
  })
})

describe("BmolBoard reordering", () => {
  async function editing() {
    const w = await mountBoard(1)
    await w.findAll("button").find((b) => b.text() === i18n.global.t("library.bmol.actions.edit"))!.trigger("click")
    return w
  }

  it("rolls back an optimistic move and shows a toast when the server fails", async () => {
    let reject!: (e: unknown) => void
    m.updateItem.mockReturnValue(new Promise((_, r) => (reject = r)))
    const w = await editing()
    await w.findAll(".bcard__menu-btn")[0]!.trigger("click")
    const down = w.findAll("[role=menuitem]").find((b) => b.text() === i18n.global.t("library.bmol.card.moveDown"))!
    await down.trigger("click")
    await flushPromises()
    expect(m.updateItem).toHaveBeenCalledWith(1, { rank: 2 })
    expect(ranks(w)).toContain("Alpha:2")

    reject(httpError(500))
    await flushPromises()
    expect(ranks(w)).toContain("Alpha:1")
    expect(useUIStore().toasts.some((x) => x.message === i18n.global.t("library.bmol.toasts.moveFailed"))).toBe(true)
  })

  it("rolls back only the failed item when two changes are in flight", async () => {
    let rejectA!: (e: unknown) => void
    m.updateItem.mockImplementation((id: number) => (id === 1 ? new Promise((_, r) => (rejectA = r)) : new Promise(() => {})))
    const w = await editing()
    // Refetches stay pending so the cache shows only the optimistic state.
    m.getUserBMOL.mockReturnValue(new Promise(() => {}))
    const step = async (menu: number, label: string) => {
      await w.findAll(".bcard__menu-btn")[menu]!.trigger("click")
      await w.findAll("[role=menuitem]").find((b) => b.text() === i18n.global.t(label))!.trigger("click")
      await flushPromises()
    }
    await step(2, "library.bmol.card.moveUp") // Gamma 3 -> 2, still pending
    await step(0, "library.bmol.card.moveDown") // Alpha 1 -> 2, will fail
    expect(ranks(w)).toEqual(expect.arrayContaining(["Alpha:2", "Gamma:2"]))
    rejectA(httpError(500))
    await flushPromises()
    expect(ranks(w)).toEqual(expect.arrayContaining(["Alpha:1", "Gamma:2", "Beta:2"]))
  })

  it("moves with the keyboard: Space grabs, Arrow chooses, Enter drops", async () => {
    m.updateItem.mockResolvedValue({ data: { item: item(3, 1, "Gamma") } })
    const w = await editing()
    const handle = w.findAll(".bcard__handle")[2]!
    await handle.trigger("keydown", { key: " " })
    await handle.trigger("keydown", { key: "ArrowUp" })
    expect(w.find("[role=status]").text()).toContain(i18n.global.t("library.bmol.drag.rank", { rank: 2 }))
    await handle.trigger("keydown", { key: "ArrowUp" })
    await handle.trigger("keydown", { key: "Enter" })
    await flushPromises()
    expect(m.updateItem).toHaveBeenCalledTimes(1)
    expect(m.updateItem).toHaveBeenCalledWith(3, { rank: 1 })
  })

  it("cancels a keyboard move with Escape", async () => {
    const w = await editing()
    const handle = w.findAll(".bcard__handle")[1]!
    await handle.trigger("keydown", { key: " " })
    await handle.trigger("keydown", { key: "ArrowUp" })
    await handle.trigger("keydown", { key: "Escape" })
    expect(m.updateItem).not.toHaveBeenCalled()
  })
})
