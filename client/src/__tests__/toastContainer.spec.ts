import { afterEach, beforeEach, describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import ToastContainer from "@/components/common/ToastContainer.vue"
import { useToast } from "@/composables/useToast"
import { useUIStore } from "@/stores/ui"
import { i18n } from "@/i18n"

describe("ToastContainer", () => {
  beforeEach(() => setActivePinia(createPinia()))
  afterEach(() => {
    document.body.innerHTML = ""
  })

  it("renders toasts raised through useToast() and removes them on close", async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const wrapper = mount(ToastContainer, { global: { plugins: [pinia, i18n] }, attachTo: document.body })
    useToast().error("Save failed")
    await wrapper.vm.$nextTick()
    expect(document.body.textContent).toContain("Save failed")
    expect(document.querySelectorAll(".toast-item")).toHaveLength(1)

    ;(document.querySelector(".toast-close-btn") as HTMLElement).click()
    expect(useUIStore().toasts).toHaveLength(0)
    wrapper.unmount()
  })
})
