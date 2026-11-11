import { createPinia, setActivePinia } from "pinia"
import { afterEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import WatchHistoryPanel from "@/components/movie/WatchHistoryPanel.vue"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/watchLog", () => ({
  watchLogApi: { getMyWatchLogs: vi.fn().mockResolvedValue({ data: { logs: [] } }), deleteWatchLog: vi.fn() },
}))

const i18n = useRealI18n()
let wrapper: VueWrapper | null = null

afterEach(() => {
  wrapper?.unmount()
  wrapper = null
  document.body.innerHTML = ""
})

describe("WatchHistoryPanel", () => {
  it("is a labelled modal dialog that locks scroll and closes on Escape", async () => {
    setActivePinia(createPinia())
    wrapper = mount(WatchHistoryPanel, {
      props: { mediaId: 1, mediaType: "movie" },
      global: { plugins: [i18n], stubs: { Teleport: true } },
    })
    await flushPromises()
    const dialog = wrapper.find('[role="dialog"]')
    expect(dialog.attributes("aria-modal")).toBe("true")
    expect(document.body.style.overflow).toBe("hidden")
    await dialog.trigger("keydown", { key: "Escape" })
    expect(wrapper.emitted("close")).toHaveLength(1)
    wrapper.unmount()
    wrapper = null
    expect(document.body.style.overflow).toBe("")
  })
})
