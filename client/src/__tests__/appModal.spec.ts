import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h, nextTick, ref } from "vue"
import { mount, type VueWrapper } from "@vue/test-utils"
import AppModal from "@/components/ui/AppModal.vue"
import { focusablesIn, setFocusableVisibilityCheck } from "@/composables/useDialogFocus"
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
const panels = () => Array.from(document.body.querySelectorAll<HTMLElement>('[aria-modal="true"]'))
const key = (target: EventTarget, k: string, init: KeyboardEventInit = {}) =>
  target.dispatchEvent(new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true, ...init }))

beforeEach(() => {
  // happy-dom has no layout engine (getClientRects() is empty unless overridden);
  // treat connected elements without `hidden` or `display: none` as visible.
  setFocusableVisibilityCheck(
    (el) => el.isConnected && !el.hidden && el.style.display !== "none" && el.style.visibility !== "hidden",
  )
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  document.body.style.overflow = ""
  setFocusableVisibilityCheck((el) => el.getClientRects().length > 0)
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

  it("omits aria-labelledby when neither title nor header slot is given and warns in dev", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {})
    mountModal({ title: undefined })
    expect(panel()!.getAttribute("aria-labelledby")).toBeNull()
    expect(warn).toHaveBeenCalledOnce()
    warn.mockClear()
    mountModal({ title: undefined, ariaLabel: "Custom" })
    expect(warn).not.toHaveBeenCalled()
    warn.mockRestore()
  })

  it("forwards consumer class to the dialog panel", () => {
    mountModal({ class: "custom-panel" })
    expect(panel()!.classList.contains("custom-panel")).toBe(true)
    expect(panel()!.classList.contains("app-modal__panel")).toBe(true)
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

  it("restores focus and releases scroll lock when unmounted with v-if while open", async () => {
    const opener = document.createElement("button")
    document.body.appendChild(opener)
    opener.focus()
    const wrapper = mountModal({ showClose: false })
    await nextTick()
    await nextTick()
    expect(document.activeElement?.id).toBe("first")
    expect(document.body.style.overflow).toBe("hidden")
    wrapper.unmount()
    expect(document.activeElement).toBe(opener)
    expect(document.body.style.overflow).toBe("")
  })

  it("does not focus a detached opener when the opener was removed before close", async () => {
    const opener = document.createElement("button")
    document.body.appendChild(opener)
    const focusSpy = vi.spyOn(opener, "focus")
    opener.focus()
    focusSpy.mockClear()
    const wrapper = mountModal({ showClose: false })
    await nextTick()
    await nextTick()
    expect(document.activeElement?.id).toBe("first")
    opener.remove()
    await wrapper.setProps({ modelValue: false })
    await nextTick()
    expect(focusSpy).not.toHaveBeenCalled()
  })

  it("filters out negative-tabindex, disabled, and zero-rect elements from focusables", () => {
    setFocusableVisibilityCheck((el) => el.getClientRects().length > 0)
    const root = document.createElement("div")
    root.innerHTML = [
      '<button id="ok">OK</button>',
      '<div id="neg" tabindex="-2">Neg</div>',
      '<button id="disabled-tabindex" disabled tabindex="0">Disabled</button>',
      '<button id="hidden">Hidden</button>',
    ].join("")
    document.body.appendChild(root)
    const rect = () => ({ length: 1 }) as unknown as DOMRectList
    const emptyRect = () => ({ length: 0 }) as unknown as DOMRectList
    root.querySelector<HTMLElement>("#ok")!.getClientRects = rect
    root.querySelector<HTMLElement>("#neg")!.getClientRects = rect
    root.querySelector<HTMLElement>("#disabled-tabindex")!.getClientRects = rect
    root.querySelector<HTMLElement>("#hidden")!.getClientRects = emptyRect
    expect(focusablesIn(root).map((el) => el.id)).toEqual(["ok"])
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

  it("wraps Tab between the last and first focusable, and pulls Tab from body into the top panel", async () => {
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
    document.body.focus()
    key(document.body, "Tab")
    expect(document.activeElement).toBe(first)
  })

  it("keeps focus inside the panel when the focused button becomes disabled", async () => {
    mountModal({ showClose: false })
    await nextTick()
    await nextTick()
    const first = document.getElementById("first") as HTMLButtonElement
    expect(document.activeElement).toBe(first)
    first.disabled = true
    await nextTick()
    await new Promise((r) => setTimeout(r, 0))
    expect(document.activeElement?.id).toBe("last")
  })

  it("closes on Escape with a reason", async () => {
    const wrapper = mountModal()
    key(document.getElementById("first")!, "Escape")
    expect(wrapper.emitted("close")![0]).toEqual(["escape"])
    expect(wrapper.emitted("update:modelValue")![0]).toEqual([false])
  })

  it("ignores Escape when defaultPrevented is already set", () => {
    const wrapper = mountModal()
    const ev = new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true })
    ev.preventDefault()
    document.getElementById("first")!.dispatchEvent(ev)
    expect(wrapper.emitted("close")).toBeUndefined()
  })

  it("ignores Escape from a teleported menu outside the panel", () => {
    const wrapper = mountModal()
    const menu = document.createElement("div")
    document.body.appendChild(menu)
    key(menu, "Escape")
    expect(wrapper.emitted("close")).toBeUndefined()
  })

  it("dismisses stacked modals one by one on successive Escapes", async () => {
    const lower = mountModal({ title: "Lower" })
    const upper = mountModal({ title: "Upper" })
    key(document.body, "Escape")
    expect(upper.emitted("close")).toHaveLength(1)
    expect(lower.emitted("close")).toBeUndefined()
    await upper.setProps({ modelValue: false })
    key(document.body, "Escape")
    expect(lower.emitted("close")).toHaveLength(1)
  })

  it("assigns increasing z-index by open order, including when a lower modal reopens on top", async () => {
    const a = mountModal({ title: "A" })
    const b = mountModal({ title: "B" })
    const zOfTitle = (title: string) => {
      const m = Array.from(document.body.querySelectorAll<HTMLElement>(".app-modal")).find((el) =>
        el.textContent?.includes(title),
      )!
      return Number(m.style.zIndex)
    }
    expect(zOfTitle("A")).toBeGreaterThan(99999)
    expect(zOfTitle("B")).toBeGreaterThan(zOfTitle("A"))
    await a.setProps({ modelValue: false })
    await a.setProps({ modelValue: true })
    await nextTick()
    expect(zOfTitle("A")).toBeGreaterThan(zOfTitle("B"))
    key(document.body, "Escape")
    expect(a.emitted("close")).toHaveLength(1)
    expect(b.emitted("close")).toBeUndefined()
  })

  it("does not emit close twice when dismissed during leave after modelValue turns false", async () => {
    const wrapper = mount(AppModal, {
      props: { modelValue: true, title: "Hello" },
      global: { plugins: [i18n], stubs: { transition: false } },
      attachTo: document.body,
    })
    wrappers.push(wrapper)
    const closeBtn = document.body.querySelector<HTMLButtonElement>(".app-modal__header button")!
    closeBtn.click()
    expect(wrapper.emitted("close")).toHaveLength(1)
    await wrapper.setProps({ modelValue: false })
    closeBtn.click()
    expect(wrapper.emitted("close")).toHaveLength(1)
  })

  it("respects closeOnEscape and dismissDisabled", () => {
    const a = mountModal({ closeOnEscape: false })
    key(document.body, "Escape")
    expect(a.emitted("close")).toBeUndefined()
    const b = mountModal({ dismissDisabled: true })
    key(document.body, "Escape")
    expect(b.emitted("close")).toBeUndefined()
  })

  it("closes on primary-button backdrop click only, keeping focus on mousedown", () => {
    const wrapper = mountModal({ showClose: false })
    const backdrop = panel()!.parentElement!
    const down = (t: Element, button = 0) =>
      t.dispatchEvent(new MouseEvent("mousedown", { bubbles: true, cancelable: true, button }))
    const up = (t: Element, button = 0) =>
      t.dispatchEvent(new MouseEvent("mouseup", { bubbles: true, cancelable: true, button }))
    down(panel()!)
    up(backdrop)
    expect(wrapper.emitted("close")).toBeUndefined()
    down(backdrop)
    up(panel()!)
    expect(wrapper.emitted("close")).toBeUndefined()
    down(backdrop, 2)
    up(backdrop, 2)
    expect(wrapper.emitted("close")).toBeUndefined()
    const prevented = !down(backdrop, 0)
    expect(prevented).toBe(true)
    up(backdrop, 0)
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
