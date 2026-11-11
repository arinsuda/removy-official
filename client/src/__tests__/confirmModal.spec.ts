import { afterEach, beforeEach, describe, expect, it } from "vitest"
import { defineComponent, h, nextTick, ref } from "vue"
import { mount, type VueWrapper } from "@vue/test-utils"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import en from "@/i18n/locales/en/profile.json"
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

// Parent with an opener button, the way consumers use it.
function mountHarness(listType: "unfollow" | "watched" | "email_change" = "unfollow") {
  const events: string[] = []
  const Harness = defineComponent({
    setup() {
      const open = ref(false)
      return () => [
        h("button", { id: "opener", onClick: () => (open.value = true) }, "Open"),
        h(ConfirmModal, {
          modelValue: open.value,
          listType,
          itemName: "Jane",
          "onUpdate:modelValue": (v: boolean) => (open.value = v),
          onConfirm: () => {
            events.push("confirm")
            open.value = false
          },
          onCancel: () => {
            events.push("cancel")
            open.value = false
          },
        }),
      ]
    },
  })
  const wrapper = mount(Harness, { global: { plugins: [i18n] }, attachTo: document.body })
  wrappers.push(wrapper)
  return { events }
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

describe("ConfirmModal", () => {
  it("has its text in both locales", () => {
    expectKeysInBothLocales(
      ["unfollow", "watched", "email_change", "likes", "watchlist", "bmol_delete"].flatMap((k) =>
        ["title", "description", "suffix", "confirm"].map((f) => `profile.confirmModal.${k}.${f}`),
      ),
    )
  })

  it("renders a labelled alertdialog with the item name and focuses Cancel first", async () => {
    mountHarness()
    await openIt()
    const el = dialog()!
    expect(document.getElementById(el.getAttribute("aria-labelledby")!)?.textContent).toBe(en.confirmModal.unfollow.title)
    expect(el.textContent).toContain("Jane")
    expect(document.activeElement).toBe(btn("Cancel"))
  })

  it("uses the danger variant for destructive types", async () => {
    mountHarness("unfollow")
    await openIt()
    expect(btn(en.confirmModal.unfollow.confirm).classList.contains("app-btn--danger")).toBe(true)
  })

  it("emits confirm from the confirm button", async () => {
    const { events } = mountHarness()
    await openIt()
    clickAfterRender(btn(en.confirmModal.unfollow.confirm))
    expect(events).toEqual(["confirm"])
  })

  it("emits cancel from the cancel button and returns focus to the opener", async () => {
    const { events } = mountHarness()
    const opener = await openIt()
    clickAfterRender(btn("Cancel"))
    await nextTick()
    await nextTick()
    expect(events).toEqual(["cancel"])
    expect(document.activeElement).toBe(opener)
  })

  it("emits cancel on Escape and returns focus to the opener", async () => {
    const { events } = mountHarness()
    const opener = await openIt()
    document.activeElement!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }))
    await nextTick()
    await nextTick()
    expect(events).toEqual(["cancel"])
    expect(dialog()).toBeNull()
    expect(document.activeElement).toBe(opener)
  })
})
