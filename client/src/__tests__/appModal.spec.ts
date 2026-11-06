import { afterEach, describe, expect, it } from "vitest"
import { defineComponent, h, nextTick, ref } from "vue"
import { mount, type VueWrapper } from "@vue/test-utils"
import AppModal from "@/components/ui/AppModal.vue"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()
const wrappers: VueWrapper[] = []

function mountModal(props: Record<string, unknown> = {}, slots: Record<string, string> = {}) {
  const wrapper = mount(AppModal, {
    props: { modelValue: true, title: "Hello", ...props },
    slots: { default: '<button id="first">One</button><button id="last">Two</button>', ...slots },
    global: { plugins: [i18n] },
    attachTo: document.body,
  })
  wrappers.push(wrapper)
  return wrapper
}

const panel = () => document.body.querySelector<HTMLElement>('[aria-modal="true"]')
const key = (target: EventTarget, k: string, init: KeyboardEventInit = {}) =>
  target.dispatchEvent(new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true, ...init }))

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  document.body.style.overflow = ""
})

describe("AppModal", () => {
  it("renders an accessible dialog in body labelled by its title", () => {
    mountModal()
    const el = panel()!
    expect(el.getAttribute("role")).toBe("dialog")
    const labelledBy = el.getAttribute("aria-labelledby")!
    expect(document.getElementById(labelledBy)?.textContent).toContain("Hello")
  })

  it("supports alertdialog and aria-label", () => {
    mountModal({ role: "alertdialog", ariaLabel: "Confirm" })
    expect(panel()!.getAttribute("role")).toBe("alertdialog")
    expect(panel()!.getAttribute("aria-label")).toBe("Confirm")
    expect(panel()!.getAttribute("aria-labelledby")).toBeNull()
  })

  it("renders nothing when closed", () => {
    mountModal({ modelValue: false })
    expect(panel()).toBeNull()
  })

  it("focuses the first focusable and returns focus to the opener", async () => {
    const opener = document.createElement("button")
    document.body.appendChild(opener)
    opener.focus()
    const wrapper = mountModal({ modelValue: false, showClose: false })
    await wrapper.setProps({ modelValue: true })
    await nextTick()
    expect(document.activeElement?.id).toBe("first")
    await wrapper.setProps({ modelValue: false })
    await nextTick()
    expect(document.activeElement).toBe(opener)
  })

  it("focuses the panel when nothing is focusable", async () => {
    mountModal({ showClose: false }, { default: "<p>text</p>" })
    await nextTick()
    await nextTick()
    expect(document.activeElement).toBe(panel())
  })

  it("honors initialFocus", async () => {
    mountModal({ initialFocus: "#last" })
    await nextTick()
    await nextTick()
    expect(document.activeElement?.id).toBe("last")
  })

  it("wraps Tab between the last and first focusable", async () => {
    mountModal()
    await nextTick()
    await nextTick()
    const last = document.getElementById("last")!
    last.focus()
    const first = panel()!.querySelector<HTMLElement>("button")!
    key(last, "Tab")
    expect(document.activeElement).toBe(first)
    key(first, "Tab", { shiftKey: true })
    expect(document.activeElement).toBe(last)
  })

  it("closes on Escape with a reason", async () => {
    const wrapper = mountModal()
    key(document.getElementById("first")!, "Escape")
    expect(wrapper.emitted("close")![0]).toEqual(["escape"])
    expect(wrapper.emitted("update:modelValue")![0]).toEqual([false])
  })

  it("ignores Escape from a teleported menu outside the panel", () => {
    const wrapper = mountModal()
    const menu = document.createElement("div")
    document.body.appendChild(menu)
    key(menu, "Escape")
    expect(wrapper.emitted("close")).toBeUndefined()
  })

  it("only the top modal handles Escape", () => {
    const lower = mountModal({ title: "Lower" })
    const upper = mountModal({ title: "Upper" })
    key(document.body, "Escape")
    expect(upper.emitted("close")).toHaveLength(1)
    expect(lower.emitted("close")).toBeUndefined()
  })

  it("respects closeOnEscape and dismissDisabled", () => {
    const a = mountModal({ closeOnEscape: false })
    key(document.body, "Escape")
    expect(a.emitted("close")).toBeUndefined()
    const b = mountModal({ dismissDisabled: true })
    key(document.body, "Escape")
    expect(b.emitted("close")).toBeUndefined()
  })

  it("closes on backdrop only when press and release both land on it", () => {
    const wrapper = mountModal()
    const backdrop = panel()!.parentElement!
    const down = (t: Element) => t.dispatchEvent(new MouseEvent("mousedown", { bubbles: true }))
    const up = (t: Element) => t.dispatchEvent(new MouseEvent("mouseup", { bubbles: true }))
    down(panel()!)
    up(backdrop)
    expect(wrapper.emitted("close")).toBeUndefined()
    down(backdrop)
    up(panel()!)
    expect(wrapper.emitted("close")).toBeUndefined()
    down(backdrop)
    up(backdrop)
    expect(wrapper.emitted("close")![0]).toEqual(["backdrop"])
  })

  it("does not close on backdrop when closeOnBackdrop is false", () => {
    const wrapper = mountModal({ closeOnBackdrop: false })
    const backdrop = panel()!.parentElement!
    backdrop.dispatchEvent(new MouseEvent("mousedown", { bubbles: true }))
    backdrop.dispatchEvent(new MouseEvent("mouseup", { bubbles: true }))
    expect(wrapper.emitted("close")).toBeUndefined()
  })

  it("closes from the close button, and disables it when dismissDisabled", async () => {
    const wrapper = mountModal()
    document.body.querySelector<HTMLButtonElement>(".app-modal__header button")!.click()
    expect(wrapper.emitted("close")![0]).toEqual(["close-button"])
    await wrapper.setProps({ dismissDisabled: true })
    expect(document.body.querySelector<HTMLButtonElement>(".app-modal__header button")!.disabled).toBe(true)
  })

  it("ref-counts the scroll lock across stacked modals", async () => {
    const a = mountModal()
    const b = mountModal()
    expect(document.body.style.overflow).toBe("hidden")
    await b.setProps({ modelValue: false })
    expect(document.body.style.overflow).toBe("hidden")
    await a.setProps({ modelValue: false })
    expect(document.body.style.overflow).toBe("")
  })

  it("works with v-model through a parent", async () => {
    const Parent = defineComponent({
      setup() {
        const open = ref(true)
        return () => h(AppModal, { modelValue: open.value, "onUpdate:modelValue": (v: boolean) => (open.value = v), title: "T" })
      },
    })
    const wrapper = mount(Parent, { global: { plugins: [i18n] }, attachTo: document.body })
    wrappers.push(wrapper)
    key(document.body, "Escape")
    await nextTick()
    expect(panel()).toBeNull()
  })
})
