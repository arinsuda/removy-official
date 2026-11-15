import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import AdminGenerateRoundModal from "@/components/admin/events/AdminGenerateRoundModal.vue"
import { adminEventApi } from "@/api/endpoints/event"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({
  adminEventApi: {
    generateMonthly: vi.fn(),
    generateYearlyRound: vi.fn(),
    suggestNominees: vi.fn(),
  },
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

/** Mid-November 2026 in Thailand: the next monthly round is 2026-12, the next yearly round is 2026. */
const NOW = new Date("2026-11-15T12:00:00Z")

async function openModal() {
  const w = mount(AdminGenerateRoundModal, {
    props: { modelValue: false, "onUpdate:modelValue": () => {} },
    global: { plugins: [i18n] },
    attachTo: document.body,
  })
  wrappers.push(w)
  await w.setProps({ modelValue: true })
  await flushPromises()
  return w
}

const periodInput = () => document.body.querySelector<HTMLInputElement>("#gen-period")!
const lateCheckbox = () => document.body.querySelector<HTMLInputElement>('input[name="generate-late-start"]')
const kindRadio = (value: string) =>
  document.body.querySelector<HTMLInputElement>(`input[name="generate-round-kind"][value="${value}"]`)!
const mediaRadio = (value: string) =>
  document.body.querySelector<HTMLInputElement>(`input[name="generate-media-type"][value="${value}"]`)!
const submit = () => document.body.querySelector("form.event-form")!.dispatchEvent(new Event("submit"))

async function choose(input: HTMLInputElement) {
  input.checked = true
  input.dispatchEvent(new Event("change"))
  await flushPromises()
}

beforeEach(() => {
  i18n.global.locale.value = "en"
  vi.useFakeTimers({ toFake: ["Date"] })
  vi.setSystemTime(NOW)
  api.generateMonthly.mockResolvedValue({} as never)
  api.generateYearlyRound.mockResolvedValue({} as never)
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  vi.useRealTimers()
  vi.clearAllMocks()
})

describe("AdminGenerateRoundModal", () => {
  it("opens on the monthly kind with the next monthly period", async () => {
    await openModal()
    expect(kindRadio("monthly").checked).toBe(true)
    expect(periodInput().value).toBe("2026-12")
    expect(document.body.textContent).toContain(tr("admin.events.generateRoundModalTitle"))
  })

  it("switches to the yearly kind with the next yearly period and no nominee preview", async () => {
    await openModal()
    await choose(kindRadio("yearly"))

    expect(periodInput().value).toBe("2026")
    expect(document.body.querySelector("#gen-period")?.getAttribute("placeholder")).toBe(
      tr("admin.events.yearlyPeriodPlaceholder"),
    )
    expect(document.body.textContent).toContain(tr("admin.events.yearlyNomineesHint"))
    expect(document.body.textContent).not.toContain(tr("admin.events.previewNominees"))
  })

  it("rejects a yearly period that is not four digits and sends nothing", async () => {
    await openModal()
    await choose(kindRadio("yearly"))

    periodInput().value = "26"
    periodInput().dispatchEvent(new Event("input"))
    submit()
    await flushPromises()

    expect(api.generateYearlyRound).not.toHaveBeenCalled()
    expect(document.body.textContent).toContain(tr("admin.events.errors.yearlyPeriodInvalid"))
  })

  it("creates a yearly round for the chosen media type", async () => {
    await openModal()
    await choose(kindRadio("yearly"))
    await choose(mediaRadio("tv"))

    submit()
    await flushPromises()

    expect(api.generateYearlyRound).toHaveBeenCalledWith({ media_type: "tv", period: "2026", late_start: false })
    expect(api.generateMonthly).not.toHaveBeenCalled()
    expect(toastSuccess).toHaveBeenCalledWith(tr("admin.events.generateSuccess"))
  })

  it("keeps creating monthly rounds through the monthly endpoint", async () => {
    await openModal()
    submit()
    await flushPromises()

    expect(api.generateMonthly).toHaveBeenCalledWith({ media_type: "movie", period: "2026-12", late_start: false })
    expect(api.generateYearlyRound).not.toHaveBeenCalled()
  })

  it("offers a late start of the current monthly round and sends it with the current period", async () => {
    await openModal()
    expect(lateCheckbox()).not.toBeNull()

    await choose(lateCheckbox()!)
    expect(periodInput().value).toBe("2026-11")

    submit()
    await flushPromises()

    expect(api.generateMonthly).toHaveBeenCalledWith({ media_type: "movie", period: "2026-11", late_start: true })
  })

  it("hides the late start once too little of the current window is left", async () => {
    // Dec 12 2026 12:00 in Thailand: the round closes Dec 15, less than 72 hours after the late start.
    vi.setSystemTime(new Date("2026-12-12T05:00:00Z"))
    await openModal()

    expect(lateCheckbox()).toBeNull()
  })

  it("maps VOTING_WINDOW_TOO_SHORT to a readable toast", async () => {
    api.generateMonthly.mockRejectedValueOnce({
      response: { status: 409, data: { code: "VOTING_WINDOW_TOO_SHORT" } },
    })
    await openModal()
    await choose(lateCheckbox()!)

    submit()
    await flushPromises()

    expect(toastError).toHaveBeenCalledWith(tr("admin.events.errors.votingWindowTooShort"))
  })

  it("maps MONTHLY_ROUNDS_PENDING to a readable toast", async () => {
    api.generateYearlyRound.mockRejectedValueOnce({
      response: { status: 409, data: { code: "MONTHLY_ROUNDS_PENDING" } },
    })
    await openModal()
    await choose(kindRadio("yearly"))

    submit()
    await flushPromises()

    expect(toastError).toHaveBeenCalledWith(tr("admin.events.errors.monthlyRoundsPending"))
  })
})
