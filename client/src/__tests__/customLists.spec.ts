import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import AddToListSheet from "@/components/lists/AddToListSheet.vue"
import LibraryAddSheet from "@/views/library/components/LibraryAddSheet.vue"
import ListEditorModal from "@/views/library/components/ListEditorModal.vue"
import { ADD_SEARCH_DEBOUNCE_MS } from "@/views/library/config"
import { listErrorKey } from "@/components/lists/listErrors"
import { moveIndex } from "@/views/library/useListReorder"
import { useUIStore } from "@/stores/ui"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({
  getOwn: vi.fn(),
  addItem: vi.fn(),
  removeItem: vi.fn(),
  get: vi.fn(),
  create: vi.fn(),
  search: vi.fn(),
  searchSeries: vi.fn(),
}))
vi.mock("@/api/endpoints/list", () => ({ listApi: m }))
vi.mock("@/api/api", () => ({
  libraryApi: { addItem: vi.fn() },
  movieApi: { search: m.search, searchSeries: m.searchSeries },
}))

const i18n = useRealI18n()
const wrappers: Array<{ unmount: () => void }> = []
// Modals teleport into body; unmount wrappers and drop what a previous test left there.
afterEach(() => {
  vi.useRealTimers()
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
})
const list = (id: number, contains: boolean, itemId?: number) => ({
  id,
  name: `L${id}`,
  item_count: 2,
  contains_media: contains,
  ...(itemId !== undefined ? { item_id: itemId } : {}),
  collage_posters: [],
  visibility: "public",
})

function mountWith(component: object, props: object) {
  i18n.global.locale.value = "en"
  setActivePinia(createPinia())
  const w = mount(component as never, {
    props,
    attachTo: document.body,
    global: { plugins: [i18n, [VueQueryPlugin, { queryClient: new QueryClient() }]] },
  })
  wrappers.push(w)
  return w
}

describe("moveIndex / listErrorKey", () => {
  it("moves one element and ignores no-ops", () => {
    expect(moveIndex([1, 2, 3, 4], 0, 2)).toEqual([2, 3, 1, 4])
    const same = [1, 2]
    expect(moveIndex(same, 1, 1)).toBe(same)
    expect(moveIndex(same, 0, 5)).toBe(same)
  })
  it("maps known server codes and falls back to generic", () => {
    expect(listErrorKey({ response: { data: { code: "ITEM_EXISTS" } } })).toBe("library.customLists.errors.ITEM_EXISTS")
    expect(listErrorKey({ response: { data: { code: "WAT" } } })).toBe("library.customLists.errors.generic")
  })
})

describe("AddToListSheet", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    m.getOwn.mockResolvedValue({ data: { lists: [list(1, false), list(2, true, 77)] } })
    m.addItem.mockResolvedValue({ data: { item: { id: 55 } } })
    m.removeItem.mockResolvedValue({})
  })

  it("adds to an unchecked list and removes from a checked one using item_id directly", async () => {
    mountWith(AddToListSheet, { modelValue: true, mediaId: 9, mediaType: "movie" })
    await flushPromises()
    expect(m.getOwn).toHaveBeenCalledWith({ media_type: "movie", media_id: 9 })
    const box = (id: number) => document.body.querySelector<HTMLInputElement>(`[data-list-id="${id}"]`)!
    expect(box(1).checked).toBe(false)
    expect(box(2).checked).toBe(true)
    box(1).click()
    await flushPromises()
    expect(m.addItem).toHaveBeenCalledWith(1, { media_id: 9, media_type: "movie" })
    box(2).click()
    await flushPromises()
    expect(m.removeItem).toHaveBeenCalledWith(2, 77)
    expect(m.get).not.toHaveBeenCalled()
  })
})

describe("ListEditorModal", () => {
  it("blocks an empty name and creates with trimmed values", async () => {
    m.create.mockResolvedValue({ data: { list: list(3, false) } })
    const w = mountWith(ListEditorModal, { modelValue: true })
    await flushPromises()
    const submit = () => document.body.querySelector<HTMLButtonElement>('button[type="submit"][form="list-editor-form"]')!
    expect(submit().disabled).toBe(true)
    const name = document.body.querySelector<HTMLInputElement>('[data-field="name"]')!
    name.value = "  Marathon  "
    name.dispatchEvent(new Event("input"))
    await flushPromises()
    expect(submit().disabled).toBe(false)
    document.body.querySelector<HTMLFormElement>("#list-editor-form")!.dispatchEvent(new Event("submit"))
    await flushPromises()
    expect(m.create).toHaveBeenCalledWith({ name: "Marathon", description: "", visibility: "public" })
    expect(w.emitted("saved")).toHaveLength(1)
  })
})

describe("AddToListSheet failure handling", () => {
  it("restores the checkbox and toasts when item_id is missing on a checked list", async () => {
    m.getOwn.mockResolvedValue({ data: { lists: [list(2, true)] } })
    mountWith(AddToListSheet, { modelValue: true, mediaId: 9, mediaType: "movie" })
    await flushPromises()
    const box = document.body.querySelector<HTMLInputElement>('[data-list-id="2"]')!
    box.click()
    await flushPromises()
    expect(m.removeItem).not.toHaveBeenCalled()
    expect(box.checked).toBe(true)
    expect(useUIStore().toasts.some((x) => x.type === "error")).toBe(true)
  })
})

describe("LibraryAddSheet custom list target", () => {
  it("searches and adds a title to the target custom list", async () => {
    m.search.mockResolvedValue({ data: { results: [{ id: 603, title: "The Matrix", poster_path: "/m.jpg", release_date: "1999-03-31" }] } })
    m.searchSeries.mockResolvedValue({ data: { results: [] } })
    m.addItem.mockResolvedValue({ data: { item: { id: 101 } } })

    const w = mountWith(LibraryAddSheet, { modelValue: true, savedKeys: new Set<string>(), listId: 42, listName: "Sci-Fi Classics" })
    await flushPromises()
    expect(document.body.textContent).toContain("Add to Sci-Fi Classics")

    const input = document.body.querySelector<HTMLInputElement>('input[type="search"].las__input')!
    input.value = "matrix"
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] })
    input.dispatchEvent(new Event("input"))
    await vi.advanceTimersByTimeAsync(ADD_SEARCH_DEBOUNCE_MS)
    vi.useRealTimers()
    await flushPromises()

    document.body.querySelector<HTMLButtonElement>('[data-action="add"]')!.click()
    await flushPromises()

    expect(m.addItem).toHaveBeenCalledWith(42, { media_id: 603, media_type: "movie" })
    expect(w.emitted("added")).toHaveLength(1)
  })
})
