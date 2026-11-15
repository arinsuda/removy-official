import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import AdminAutoPublishToggle from "@/components/admin/events/AdminAutoPublishToggle.vue"
import { adminEventApi } from "@/api/endpoints/event"
import type { AdminEvent } from "@/types/event"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({
  adminEventApi: { setAutoPublish: vi.fn() },
}))

const toastError = vi.fn()
const toastSuccess = vi.fn()
vi.mock("@/composables/useToast", () => ({
  useToast: () => ({ error: toastError, success: toastSuccess }),
}))

const i18n = useRealI18n()
const api = vi.mocked(adminEventApi)
const tr = (key: string) => i18n.global.t(key)
const wrappers: VueWrapper[] = []

function makeEvent(overrides: Partial<AdminEvent> = {}): AdminEvent {
  return {
    id: 7,
    type: "vote_monthly",
    media_type: "movie",
    period: "2026-12",
    title: { en: "", th: "" },
    body: { en: "", th: "" },
    image_url: null,
    link_url: null,
    starts_at: "2026-12-15T00:00:00+07:00",
    ends_at: "2027-01-15T00:00:00+07:00",
    show_popup: false,
    popup_priority: 0,
    state: "upcoming",
    status: "draft",
    auto_generated: true,
    auto_publish: false,
    created_at: "2026-11-15T00:00:00Z",
    updated_at: "2026-11-15T00:00:00Z",
    ...overrides,
  }
}

function mountToggle(event: AdminEvent) {
  const w = mount(AdminAutoPublishToggle, { props: { event }, global: { plugins: [i18n] } })
  wrappers.push(w)
  return w
}

const switchButton = (w: VueWrapper) => w.find('button[role="switch"]')

beforeEach(() => {
  i18n.global.locale.value = "en"
  api.setAutoPublish.mockResolvedValue({} as never)
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  vi.clearAllMocks()
})

describe("AdminAutoPublishToggle", () => {
  it("renders for a voting draft and reflects the stored value", () => {
    const w = mountToggle(makeEvent({ auto_publish: true }))

    expect(switchButton(w).attributes("aria-checked")).toBe("true")
    expect(switchButton(w).classes()).toContain("is-on")
    expect(w.text()).toContain(tr("admin.events.autoPublish.label"))
  })

  it("renders nothing for an announcement or a round that is no longer a draft", () => {
    expect(mountToggle(makeEvent({ type: "announcement" })).find("button").exists()).toBe(false)
    expect(mountToggle(makeEvent({ status: "published" })).find("button").exists()).toBe(false)
  })

  it("turns the setting on, reports the change and confirms it", async () => {
    const w = mountToggle(makeEvent({ auto_publish: false }))

    await switchButton(w).trigger("click")
    await flushPromises()

    expect(api.setAutoPublish).toHaveBeenCalledWith(7, true)
    expect(w.emitted("changed")).toHaveLength(1)
    expect(toastSuccess).toHaveBeenCalledWith(tr("admin.events.autoPublish.saved"))
  })

  it("maps EVENT_NOT_DRAFT and keeps the switch unchanged on failure", async () => {
    api.setAutoPublish.mockRejectedValueOnce({ response: { status: 409, data: { code: "EVENT_NOT_DRAFT" } } })
    const w = mountToggle(makeEvent({ auto_publish: true }))

    await switchButton(w).trigger("click")
    await flushPromises()

    expect(api.setAutoPublish).toHaveBeenCalledWith(7, false)
    expect(toastError).toHaveBeenCalledWith(tr("admin.events.errors.notDraft"))
    expect(w.emitted("changed")).toBeUndefined()
  })

  it("maps VOTING_CLOSED when enabling after the window closed", async () => {
    api.setAutoPublish.mockRejectedValueOnce({ response: { status: 409, data: { code: "VOTING_CLOSED" } } })
    const w = mountToggle(makeEvent({ auto_publish: false }))

    await switchButton(w).trigger("click")
    await flushPromises()

    expect(toastError).toHaveBeenCalledWith(tr("admin.events.errors.votingClosed"))
  })
})
