import { afterEach, describe, expect, it, vi } from "vitest"
import { mount, type VueWrapper } from "@vue/test-utils"
import { nextTick } from "vue"
import VisibilitySelector from "@/components/common/VisibilitySelector.vue"
import { i18n } from "@/i18n"

function rect(over: Partial<DOMRect>): DOMRect {
  return { left: 0, top: 0, right: 0, bottom: 0, width: 0, height: 0, x: 0, y: 0, toJSON: () => ({}), ...over } as DOMRect
}

const mounted: VueWrapper[] = []

async function openAt(r: Partial<DOMRect>) {
  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockReturnValue(rect(r))
  const w = mount(VisibilitySelector, { global: { plugins: [i18n] }, attachTo: document.body })
  mounted.push(w)
  await w.get("button").trigger("click")
  await nextTick()
  await nextTick()
  return w
}

/** The menu is teleported to <body>, so it is looked up there instead of inside the wrapper. */
function menu(): HTMLElement {
  return document.body.querySelector("[role=listbox]") as HTMLElement
}

describe("VisibilitySelector placement", () => {
  it("keeps the menu open when an option inside the teleported menu is clicked", async () => {
    const w = await openAt({ left: 10, right: 130, top: 100, bottom: 130 })
    const option = menu().querySelectorAll("[role=option]")[2] as HTMLElement
    option.click()
    await nextTick()
    expect(w.emitted("update:modelValue")?.[0]).toEqual(["private"])
  })

  afterEach(() => {
    mounted.splice(0).forEach((w) => w.unmount())
    vi.restoreAllMocks()
    document.body.innerHTML = ""
  })

  it("aligns left when the trigger is near the left edge", async () => {
    const w = await openAt({ left: 10, right: 130, top: 100, bottom: 130 })
    const el = menu()
    expect(w.element.contains(el)).toBe(false)
    expect(el.style.left).toBe("10px")
    expect(el.style.right).toBe("")
    expect(w.get("button").attributes("aria-controls")).toBe(el.id)
  })

  it("aligns right when there is room on the left", async () => {
    await openAt({ left: 700, right: 900, top: 100, bottom: 130 })
    expect(menu().style.right).toBe(`${window.innerWidth - 900}px`)
    expect(menu().style.left).toBe("")
  })

  it("flips up near the bottom of the viewport", async () => {
    await openAt({ left: 700, right: 900, top: window.innerHeight - 60, bottom: window.innerHeight - 30 })
    expect(menu().style.bottom).toBe(`${60 + 4}px`)
    expect(menu().style.top).toBe("")
  })

  it("opens downward when there is room below", async () => {
    await openAt({ left: 700, right: 900, top: 50, bottom: 80 })
    expect(menu().style.top).toBe(`${80 + 4}px`)
  })
})
