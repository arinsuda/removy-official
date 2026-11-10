import { afterEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia } from "pinia"
import EditReview from "@/components/profile/components/EditReview.vue"
import type { ReviewResponse } from "@/types"
import { useRealI18n } from "./helpers/realI18n"

const updateReview = vi.hoisted(() => vi.fn())
vi.mock("@/api/api", () => ({ reviewApi: { updateReview } }))

const i18n = useRealI18n()
i18n.global.locale.value = "en"
const t = (k: string) => i18n.global.t(k)

const review = {
  id: 7,
  rating: 4,
  body: "Great",
  visibility: "public",
  created_at: "2026-03-01T00:00:00Z",
  media: { id: 1, title: "Dune", poster_url: "", media_type: "movie", release_date: "2021-09-15" },
} as unknown as ReviewResponse

let wrapper: VueWrapper | null = null
afterEach(() => {
  wrapper?.unmount()
  wrapper = null
  document.body.innerHTML = ""
  updateReview.mockReset()
})

function saveButton(): HTMLButtonElement {
  const dialog = document.body.querySelector('[role="dialog"]')!
  return Array.from(dialog.querySelectorAll("button")).find((b) => b.textContent?.trim() === t("reviews.update")) as HTMLButtonElement
}

describe("EditReview modal", () => {
  it("shows the media context and enables save only after a change", async () => {
    wrapper = mount(EditReview, { props: { userId: 1, review }, global: { plugins: [createPinia(), i18n] }, attachTo: document.body })
    await flushPromises()
    const dialog = document.body.querySelector('[role="dialog"]')!
    expect(dialog.textContent).toContain("Dune")
    expect(dialog.textContent).toContain("2021")
    expect(saveButton().disabled).toBe(true)

    const textarea = dialog.querySelector("textarea")!
    textarea.value = "Great, even better on rewatch"
    textarea.dispatchEvent(new Event("input"))
    await flushPromises()
    expect(saveButton().disabled).toBe(false)
  })

  it("saves with Ctrl+Enter and trims the body", async () => {
    updateReview.mockResolvedValue({ data: { review } })
    wrapper = mount(EditReview, { props: { userId: 1, review }, global: { plugins: [createPinia(), i18n] }, attachTo: document.body })
    await flushPromises()
    const textarea = document.body.querySelector("textarea")!
    textarea.value = "  New take  "
    textarea.dispatchEvent(new Event("input"))
    await flushPromises()
    textarea.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", ctrlKey: true, bubbles: true }))
    await flushPromises()
    expect(updateReview).toHaveBeenCalledWith(7, { rating: 4, body: "New take", visibility: "public" })
    expect(wrapper.emitted("saved")).toHaveLength(1)
  })
})
