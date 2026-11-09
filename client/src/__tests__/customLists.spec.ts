import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import AddToListSheet from "@/components/lists/AddToListSheet.vue"
import ListEditorModal from "@/views/library/components/ListEditorModal.vue"
import { listErrorKey } from "@/components/lists/listErrors"
import { moveIndex } from "@/views/library/useListReorder"
import { useUIStore } from "@/stores/ui"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ getOwn: vi.fn(), addItem: vi.fn(), removeItem: vi.fn(), get: vi.fn(), create: vi.fn() }))
vi.mock("@/api/endpoints/list", () => ({ listApi: m }))

const i18n = useRealI18n()
// Modals teleport into body; drop what a previous test left there.
afterEach(() => (document.body.innerHTML = ""))
const list = (id: number, contains: boolean) => ({ id, name: `L${id}`, item_count: 2, contains_media: contains, collage_posters: [], visibility: "public" })

function mountWith(component: object, props: object) {
  setActivePinia(createPinia())
  return mount(component as never, {
    props,
    attachTo: document.body,
    global: { plugins: [i18n, [VueQueryPlugin, { queryClient: new QueryClient() }]] },
  })
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
    m.getOwn.mockResolvedValue({ data: { lists: [list(1, false), list(2, true)] } })
    m.addItem.mockResolvedValue({ data: { item: { id: 55 } } })
    m.removeItem.mockResolvedValue({})
    m.get.mockResolvedValue({ data: { items: [{ id: 77, media: { id: 9 }, media_type: "movie" }], pagination: { has_more: false } } })
  })

  it("adds to an unchecked list and removes from a checked one (finding the item id)", async () => {
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
  it("restores the checkbox and toasts when the item cannot be found to remove", async () => {
    m.getOwn.mockResolvedValue({ data: { lists: [list(2, true)] } })
    m.get.mockResolvedValue({ data: { items: [], pagination: { has_more: false } } })
    const pinia = createPinia()
    setActivePinia(pinia)
    mount(AddToListSheet, { props: { modelValue: true, mediaId: 9, mediaType: "movie" }, attachTo: document.body, global: { plugins: [i18n, pinia, [VueQueryPlugin, { queryClient: new QueryClient() }]] } })
    await flushPromises()
    const box = document.body.querySelector<HTMLInputElement>('[data-list-id="2"]')!
    box.click()
    await flushPromises()
    expect(m.removeItem).not.toHaveBeenCalled()
    expect(box.checked).toBe(true)
    expect(useUIStore().toasts.some((x) => x.type === "error")).toBe(true)
  })
})
