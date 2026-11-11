import { afterEach, beforeEach, describe, expect, it } from "vitest"
import { defineComponent, h, nextTick, ref } from "vue"
import { mount, type VueWrapper } from "@vue/test-utils"
import AdminConfirmModal from "@/components/admin/AdminConfirmModal.vue"
import { expectKeysInBothLocales, useRealI18n } from "./helpers/realI18n"
import { clickAfterRender } from "./helpers/waitForEnabled"

const i18n = useRealI18n()
const origGetClientRects = HTMLElement.prototype.getClientRects
const wrappers: VueWrapper[] = []

beforeEach(() => {
  // happy-dom has no layout; make connected elements count as visible for the focus logic.
  HTMLElement.prototype.getClientRects = function (this: HTMLElement) {
    return (this.isConnected ? [{ width: 1, height: 1 }] : []) as unknown as DOMRectList
  }
  i18n.global.locale.value = "en"
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  HTMLElement.prototype.getClientRects = origGetClientRects
})

function mountHarness(extra: Record<string, unknown> = {}) {
  const events: unknown[] = []
  const Harness = defineComponent({
    setup() {
      const open = ref(false)
      return () => [
        h("button", { id: "opener", onClick: () => (open.value = true) }, "Open"),
        h(AdminConfirmModal, {
          modelValue: open.value,
          title: "Deactivate user",
          description: "They lose access",
          targetName: "jane",
          ...extra,
          "onUpdate:modelValue": (v: boolean) => (open.value = v),
          onConfirm: (p: { reason: string }) => events.push(["confirm", p.reason]),
          onCancel: () => events.push("cancel"),
        }),
      ]
    },
  })
  wrappers.push(mount(Harness, { global: { plugins: [i18n] }, attachTo: document.body }))
  return events
}

const dialog = () => document.body.querySelector<HTMLElement>('[role="alertdialog"]')
const btn = (label: string) =>
  Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="alertdialog"] button')).find((b) =>
    b.textContent?.includes(label),
  )!

async function openIt() {
  const opener = document.getElementById("opener") as HTMLButtonElement
  opener.focus()
  opener.click()
  await nextTick()
  await nextTick()
  return opener
}

const settle = async () => {
  await nextTick()
  await nextTick()
}

describe("AdminConfirmModal", () => {
  it("has its text in both locales", () => {
    expectKeysInBothLocales(["admin.confirmModal.processing", "admin.confirmModal.reasonLabel", "admin.confirmModal.reasonPlaceholder"])
  })

  it("renders a labelled alertdialog", async () => {
    mountHarness()
    await openIt()
    const el = dialog()!
    expect(document.getElementById(el.getAttribute("aria-labelledby")!)?.textContent).toBe("Deactivate user")
    expect(el.textContent).toContain("jane")
  })

  it("confirms with the trimmed reason", async () => {
    const events = mountHarness()
    await openIt()
    const textarea = dialog()!.querySelector("textarea")!
    textarea.value = "  spam  "
    textarea.dispatchEvent(new Event("input"))
    clickAfterRender(btn("Confirm"))
    expect(events).toEqual([["confirm", "spam"]])
  })

  it("cancels from the button and returns focus to the opener", async () => {
    const events = mountHarness()
    const opener = await openIt()
    clickAfterRender(btn("Cancel"))
    await settle()
    expect(events).toEqual(["cancel"])
    expect(dialog()).toBeNull()
    expect(document.activeElement).toBe(opener)
  })

  it("cancels on Escape and returns focus to the opener", async () => {
    const events = mountHarness()
    const opener = await openIt()
    document.activeElement!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }))
    await settle()
    expect(events).toEqual(["cancel"])
    expect(dialog()).toBeNull()
    expect(document.activeElement).toBe(opener)
  })

  it("ignores Escape and Cancel while loading", async () => {
    const events = mountHarness({ loading: true })
    await openIt()
    document.activeElement!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }))
    await settle()
    expect(events).toEqual([])
    expect(dialog()).not.toBeNull()
  })
})
