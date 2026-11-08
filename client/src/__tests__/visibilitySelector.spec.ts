import { afterEach, describe, expect, it, vi } from "vitest"
import { mount } from "@vue/test-utils"
import { nextTick } from "vue"
import VisibilitySelector from "@/components/common/VisibilitySelector.vue"
import { i18n } from "@/i18n"

function rect(over: Partial<DOMRect>): DOMRect {
  return { left: 0, top: 0, right: 0, bottom: 0, width: 0, height: 0, x: 0, y: 0, toJSON: () => ({}), ...over } as DOMRect
}

async function openAt(r: Partial<DOMRect>) {
  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockReturnValue(rect(r))
  const w = mount(VisibilitySelector, { global: { plugins: [i18n] }, attachTo: document.body })
  await w.get("button").trigger("click")
  await nextTick()
  await nextTick()
  return w
}

describe("VisibilitySelector placement", () => {
  afterEach(() => {
    vi.restoreAllMocks()
    document.body.innerHTML = ""
  })

  it("aligns left when the trigger is near the left edge", async () => {
    const w = await openAt({ left: 10, right: 130, top: 100, bottom: 130 })
    const menu = w.get("[role=listbox]")
    expect(menu.classes()).toContain("left-0")
    expect(menu.classes()).not.toContain("right-0")
    expect(w.get("button").attributes("aria-controls")).toBe(menu.attributes("id"))
  })

  it("aligns right when there is room on the left", async () => {
    const w = await openAt({ left: 700, right: 900, top: 100, bottom: 130 })
    expect(w.get("[role=listbox]").classes()).toContain("right-0")
  })

  it("flips up near the bottom of the viewport", async () => {
    const w = await openAt({ left: 700, right: 900, top: window.innerHeight - 60, bottom: window.innerHeight - 30 })
    expect(w.get("[role=listbox]").classes()).toContain("bottom-full")
  })

  it("opens downward when there is room below", async () => {
    const w = await openAt({ left: 700, right: 900, top: 50, bottom: 80 })
    expect(w.get("[role=listbox]").classes()).toContain("top-full")
  })
})
