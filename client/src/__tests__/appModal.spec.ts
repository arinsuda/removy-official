import { readFileSync } from "node:fs"
import { join } from "node:path"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h, nextTick, ref } from "vue"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import AppModal from "@/components/ui/AppModal.vue"
import { focusablesIn } from "@/composables/useDialogFocus"
import { useRealI18n } from "./helpers/realI18n"
import { clickAfterRender } from "./helpers/waitForEnabled"

const i18n = useRealI18n()
const wrappers: VueWrapper[] = []
const origGetClientRects = HTMLElement.prototype.getClientRects

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

beforeEach(() => {
  // happy-dom has no layout engine (getClientRects() is empty unless stubbed);
  // stub HTMLElement.prototype.getClientRects so the production filter is exercised.
  HTMLElement.prototype.getClientRects = function (this: HTMLElement) {
    const visible = this.isConnected && !this.hidden && this.style.display !== "none" && this.style.visibility !== "hidden"
    return (visible ? [{ width: 1, height: 1 }] : []) as unknown as DOMRectList
  }
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  document.body.style.overflow = ""
  HTMLElement.prototype.getClientRects = origGetClientRects
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

  it("omits aria-labelledby when neither title nor header slot is given, falls back to attr aria-labelledby, and warns in dev", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {})
    mountModal({ title: undefined })
    expect(panel()!.getAttribute("aria-labelledby")).toBeNull()
    expect(warn).toHaveBeenCalledOnce()
    warn.mockClear()
    mountModal({ title: undefined, ariaLabel: "Custom" })
    expect(warn).not.toHaveBeenCalled()
    document.body.innerHTML = ""
    mountModal({ title: undefined, "aria-labelledby": "ext-heading" })
    expect(panel()!.getAttribute("aria-labelledby")).toBe("ext-heading")
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
    const root = document.createElement("div")
    root.innerHTML = [
      '<button id="ok">OK</button>',
      '<div id="neg" tabindex="-2">Neg</div>',
      '<button id="disabled-tabindex" disabled tabindex="0">Disabled</button>',
      '<button id="hidden" style="display: none">Hidden</button>',
    ].join("")
    document.body.appendChild(root)
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

  it("wraps Tab, pulls Tab and Shift+Tab from body or outside elements into the top panel, and ignores teleported menus", async () => {
    const outside = document.createElement("button")
    outside.id = "outside"
    document.body.appendChild(outside)
    const dpMenu = document.createElement("div")
    dpMenu.className = "dp--menu-wrapper"
    const dpBtn = document.createElement("button")
    dpMenu.appendChild(dpBtn)
    document.body.appendChild(dpMenu)

    mountModal({ showClose: false })
    await nextTick()
    await nextTick()
    const first = document.getElementById("first")!
    const last = document.getElementById("last")!
    last.focus()
    key(last, "Tab")
    expect(document.activeElement).toBe(first)
    key(first, "Tab", { shiftKey: true })
    expect(document.activeElement).toBe(last)

    document.body.focus()
    key(document.body, "Tab")
    expect(document.activeElement).toBe(first)

    document.body.focus()
    key(document.body, "Tab", { shiftKey: true })
    expect(document.activeElement).toBe(last)

    outside.focus()
    key(outside, "Tab")
    expect(document.activeElement).toBe(first)

    dpBtn.focus()
    key(dpBtn, "Tab")
    expect(document.activeElement).toBe(dpBtn)
  })

  it("focuses the panel when the focused button becomes disabled on the top modal only", async () => {
    const lower = mountModal({ showClose: false, title: "Lower" }, { default: '<button id="low-btn">Low</button>' })
    const upper = mountModal({ showClose: false, title: "Upper" })
    await nextTick()
    await nextTick()
    const first = document.getElementById("first") as HTMLButtonElement
    expect(document.activeElement).toBe(first)

    // Disabling a control inside the non-top modal does not steal focus from the top modal.
    const lowBtn = document.getElementById("low-btn") as HTMLButtonElement
    lowBtn.disabled = true
    await nextTick()
    await flushPromises()
    expect(document.activeElement).toBe(first)

    // Disabling the focused control inside the top modal moves focus to the top panel itself.
    first.disabled = true
    const upperPanel = Array.from(document.body.querySelectorAll<HTMLElement>('[aria-modal="true"]'))[1]!
    await vi.waitFor(() => expect(document.activeElement).toBe(upperPanel))
    void lower
    void upper
  })

  it("does not pull focus into a lower modal when its control is disabled while focus is on body", async () => {
    mountModal({ showClose: false, title: "Lower" }, { default: '<button id="low-btn">Low</button>' })
    mountModal({ showClose: false, title: "Upper" })
    await nextTick()
    await nextTick()
    const lowBtn = document.getElementById("low-btn") as HTMLButtonElement
    const lowerPanel = Array.from(document.body.querySelectorAll<HTMLElement>('[aria-modal="true"]'))[0]!
    ;(document.activeElement as HTMLElement).blur()
    expect(document.activeElement).toBe(document.body)

    lowBtn.disabled = true
    await nextTick()
    await flushPromises()
    expect(document.activeElement).not.toBe(lowerPanel)
    expect(lowerPanel.contains(document.activeElement)).toBe(false)
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
    const a = mountModal({ title: "A", showClose: false }, { default: '<button id="a-btn">A</button>' })
    await nextTick()
    await nextTick()
    const b = mountModal({ title: "B", showClose: false }, { default: '<button id="b-btn">B</button>' })
    await nextTick()
    await nextTick()
    const offsetOfTitle = (title: string) => {
      const m = Array.from(document.body.querySelectorAll<HTMLElement>(".app-modal")).find((el) =>
        el.textContent?.includes(title),
      )!
      const match = m.style.zIndex.match(/\+\s*(\d+)/)
      return match ? Number(match[1]) : 0
    }
    expect(offsetOfTitle("B")).toBeGreaterThan(offsetOfTitle("A"))
    expect(document.activeElement?.id).toBe("b-btn")

    // Closing the lower modal `a` while `b` is on top must not steal focus from `b`.
    await a.setProps({ modelValue: false })
    await nextTick()
    expect(document.activeElement?.id).toBe("b-btn")

    await a.setProps({ modelValue: true })
    await nextTick()
    await nextTick()
    expect(offsetOfTitle("A")).toBeGreaterThan(offsetOfTitle("B"))
    expect(document.activeElement?.id).toBe("a-btn")
    key(document.body, "Escape")
    expect(a.emitted("close")).toHaveLength(1)
    expect(b.emitted("close")).toBeUndefined()
  })

  it("defines shared z-index tokens in order: legacy (99999) < modal < popover < toast", () => {
    const css = readFileSync(join(import.meta.dirname, "../assets/styles/main.css"), "utf8")
    const token = (name: string) => Number(css.match(new RegExp(`${name}:\\s*(\\d+)`))?.[1])
    const modalBase = token("--z-modal-base")
    const popover = token("--z-popover")
    const toast = token("--z-toast")
    expect(modalBase).toBeGreaterThan(99999)
    expect(popover).toBeGreaterThan(modalBase)
    expect(toast).toBeGreaterThan(popover)
    expect(css).toMatch(/body\s+\.dp--menu-wrapper\s*\{[^}]*z-index:\s*var\(--z-popover\)/)
    expect(css).not.toContain(".dp--menu-index")
  })

  it("does not emit close twice when dismissed during leave after modelValue turns false", async () => {
    const wrapper = mount(AppModal, {
      props: { modelValue: true, title: "Hello" },
      global: { plugins: [i18n], stubs: { transition: false } },
      attachTo: document.body,
    })
    wrappers.push(wrapper)
    await nextTick()
    await flushPromises()
    const closeBtn = document.body.querySelector<HTMLButtonElement>(".app-modal__header button")!
    clickAfterRender(closeBtn)
    expect(wrapper.emitted("close")).toHaveLength(1)
    await wrapper.setProps({ modelValue: false })
    clickAfterRender(closeBtn)
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

  it("closes on primary-button backdrop click only, keeping focus on mousedown and rejecting mixed buttons", () => {
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
    // Secondary mousedown followed by primary mouseup must not close.
    down(backdrop, 2)
    up(backdrop, 0)
    expect(wrapper.emitted("close")).toBeUndefined()
    // Primary mousedown followed by secondary mouseup must not close, and resets pressed state.
    down(backdrop, 0)
    up(backdrop, 2)
    expect(wrapper.emitted("close")).toBeUndefined()
    up(backdrop, 0)
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
    await nextTick()
    await flushPromises()
    clickAfterRender(document.body.querySelector<HTMLButtonElement>(".app-modal__header button")!)
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

  it("leaves the page scrollable when a legacy writer locks before AppModal and clears before AppModal closes", async () => {
    document.body.style.overflow = "hidden"
    const modal = mountModal()
    expect(document.body.style.overflow).toBe("hidden")
    // Legacy modal closes and clears overflow directly while AppModal is still open.
    document.body.style.overflow = ""
    await modal.setProps({ modelValue: false })
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
