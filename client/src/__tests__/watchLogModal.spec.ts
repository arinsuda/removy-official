import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { nextTick } from "vue"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import WatchLogModal from "@/components/movie/WatchLogModal.vue"
import { expectKeysInBothLocales, useRealI18n } from "./helpers/realI18n"

const { createWatchLog, toastSuccess, toastError } = vi.hoisted(() => ({
  createWatchLog: vi.fn(),
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}))

vi.mock("@/api/endpoints/watchLog", () => ({ watchLogApi: { createWatchLog } }))
vi.mock("@/composables/useToast", () => ({
  useToast: () => ({ success: toastSuccess, error: toastError, warning: vi.fn(), toast: vi.fn() }),
}))

// Emits the same open/closed events as the real picker.
vi.mock("@vuepic/vue-datepicker", async () => {
  const { defineComponent, h } = await import("vue")
  const VueDatePicker = defineComponent({
    emits: ["open", "closed"],
    setup(_, { emit }) {
      return () =>
        h("div", [
          h("button", { id: "dp-open", onClick: () => emit("open") }),
          h("button", { id: "dp-close", onClick: () => emit("closed") }),
        ])
    },
  })
  return { VueDatePicker }
})
vi.mock("@vuepic/vue-datepicker/dist/main.css", () => ({}))

const i18n = useRealI18n()
const origGetClientRects = HTMLElement.prototype.getClientRects
let wrapper: VueWrapper

beforeEach(() => {
  HTMLElement.prototype.getClientRects = function (this: HTMLElement) {
    return (this.isConnected ? [{ width: 1, height: 1 }] : []) as unknown as DOMRectList
  }
  i18n.global.locale.value = "en"
  vi.clearAllMocks()
})

afterEach(() => {
  wrapper?.unmount()
  document.body.innerHTML = ""
  HTMLElement.prototype.getClientRects = origGetClientRects
})

async function mountModal() {
  wrapper = mount(WatchLogModal, {
    props: { mediaId: 7, mediaType: "movie" },
    global: {
      plugins: [i18n],
      stubs: { VisibilitySelector: true },
    },
    attachTo: document.body,
  })
  await nextTick()
  await nextTick()
}

const saveBtn = () =>
  Array.from(document.body.querySelectorAll<HTMLButtonElement>("[role='dialog'] button")).find((b) =>
    b.textContent?.includes("Save"),
  )!
const pressEscape = () =>
  document.activeElement!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }))

describe("WatchLogModal", () => {
  it("has its text in both locales", () => {
    expectKeysInBothLocales(["watchLog.save", "watchLog.createSuccess", "watchLog.createFailed", "watchLog.recordWatch"])
  })

  it("emits logged and close only after the POST succeeds", async () => {
    let resolve!: () => void
    createWatchLog.mockReturnValue(new Promise<void>((r) => (resolve = r)))
    await mountModal()
    saveBtn().click()
    await nextTick()
    expect(createWatchLog).toHaveBeenCalledWith("movie", 7, expect.objectContaining({ visibility: "public" }))
    expect(wrapper.emitted("logged")).toBeUndefined()
    expect(wrapper.emitted("close")).toBeUndefined()
    resolve()
    await flushPromises()
    expect(wrapper.emitted("logged")).toHaveLength(1)
    expect(wrapper.emitted("close")).toHaveLength(1)
    expect(toastSuccess).toHaveBeenCalledOnce()
  })

  it("shows an error toast and stays open when the POST fails", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {})
    createWatchLog.mockRejectedValue(new Error("boom"))
    await mountModal()
    saveBtn().click()
    await flushPromises()
    expect(toastError).toHaveBeenCalledOnce()
    expect(wrapper.emitted("logged")).toBeUndefined()
    expect(wrapper.emitted("close")).toBeUndefined()
    expect(document.body.querySelector("[role='dialog']")).not.toBeNull()
  })

  it("closes on Escape, but not while the datepicker menu is open", async () => {
    await mountModal()
    document.getElementById("dp-open")!.click()
    await nextTick()
    pressEscape()
    await nextTick()
    expect(wrapper.emitted("close")).toBeUndefined()
    document.getElementById("dp-close")!.click()
    await nextTick()
    pressEscape()
    await nextTick()
    expect(wrapper.emitted("close")).toHaveLength(1)
  })
})
