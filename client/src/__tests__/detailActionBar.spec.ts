import { beforeEach, describe, expect, it, vi } from "vitest"
import { defineComponent, h } from "vue"
import { mount } from "@vue/test-utils"
import DetailActionBar from "@/components/detail/DetailActionBar.vue"
import { i18n } from "@/i18n"

const LogModal = defineComponent({
  emits: ["close", "logged"],
  setup: (_, { emit }) => () => h("button", { "data-test": "modal-save", onClick: () => (emit("logged"), emit("close")) }),
})

function mountBar(props: Partial<InstanceType<typeof DetailActionBar>["$props"]> = {}) {
  return mount(DetailActionBar, {
    props: { mediaId: 1, mediaType: "movie", watchlisted: false, hasOwnReview: false, requireLogin: () => true, ...props },
    global: { plugins: [i18n], stubs: { WatchLogModal: LogModal } },
  })
}

describe("DetailActionBar", () => {
  beforeEach(() => {
    i18n.global.locale.value = "en"
  })

  it("prompts guests with specific reasons instead of rating, logging, or opening lists", async () => {
    const requireLogin = vi.fn(() => false)
    const w = mountBar({ requireLogin })
    await w.get('[data-test="action-rate"]').trigger("click")
    await w.get('[data-test="action-log"]').trigger("click")
    await w.get('[data-test="action-list"]').trigger("click")
    expect(requireLogin.mock.calls).toEqual([["rate"], ["watchLog"], ["list"]])
    expect(w.emitted("rate")).toBeUndefined()
    expect(w.find('[data-test="modal-save"]').exists()).toBe(false)
  })

  it("emits rate for signed-in users and relabels when a review exists", async () => {
    const w = mountBar({ hasOwnReview: true })
    expect(w.get('[data-test="action-rate"]').text()).toBe("Edit rating")
    await w.get('[data-test="action-rate"]').trigger("click")
    expect(w.emitted("rate")).toHaveLength(1)
  })

  it("reflects the watchlist state in aria-pressed", async () => {
    const w = mountBar()
    const btn = w.get('[data-test="action-watchlist"]')
    expect(btn.attributes("aria-pressed")).toBe("false")
    await btn.trigger("click")
    expect(w.emitted("watchlist")).toHaveLength(1)
    await w.setProps({ watchlisted: true })
    expect(btn.attributes("aria-pressed")).toBe("true")
  })

  it("opens the log modal and reports a saved log", async () => {
    const w = mountBar()
    await w.get('[data-test="action-log"]').trigger("click")
    await w.get('[data-test="modal-save"]').trigger("click")
    expect(w.emitted("logged")).toHaveLength(1)
    expect(w.find('[data-test="modal-save"]').exists()).toBe(false)
  })

  it("closes the log modal when the title changes", async () => {
    const w = mountBar()
    await w.get('[data-test="action-log"]').trigger("click")
    expect(w.find('[data-test="modal-save"]').exists()).toBe(true)
    await w.setProps({ mediaId: 2 })
    expect(w.find('[data-test="modal-save"]').exists()).toBe(false)
  })

  it("emits share", async () => {
    const w = mountBar()
    await w.get('[data-test="action-share"]').trigger("click")
    expect(w.emitted("share")).toHaveLength(1)
  })
})
