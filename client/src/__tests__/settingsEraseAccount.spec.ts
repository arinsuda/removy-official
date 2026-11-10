import { afterEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import SettingsDangerSection from "@/components/settings/SettingsDangerSection.vue"
import SettingsEraseAccount from "@/components/settings/SettingsEraseAccount.vue"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ requestErasure: vi.fn(), push: vi.fn(), reset: vi.fn(), success: vi.fn() }))
vi.mock("@/api/api", () => ({ userApi: { requestErasure: m.requestErasure, deleteUser: vi.fn() }, authApi: {} }))
vi.mock("vue-router", () => ({ useRouter: () => ({ push: m.push }) }))
vi.mock("@/composables/useSessionReset", () => ({ useSessionReset: () => m.reset }))
vi.mock("@/composables/useSocket", () => ({ resume: vi.fn() }))
vi.mock("@/composables/useToast", () => ({ useToast: () => ({ success: m.success, error: vi.fn() }) }))

const i18n = useRealI18n()
const err = (status: number, code: string) => ({ response: { status, data: { code } } })
const user = { id: 1, username: "arin" } as never

const mounted: { unmount: () => void }[] = []
afterEach(() => {
  mounted.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  vi.clearAllMocks()
})

async function open() {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = user
  const w = mount(SettingsEraseAccount, { props: { user }, attachTo: document.body, global: { plugins: [i18n, pinia] } })
  mounted.push(w)
  await w.get("button").trigger("click")
  await flushPromises()
  return w
}
const type = async (sel: string, v: string) => {
  const el = document.body.querySelector<HTMLInputElement>(sel)!
  el.value = v
  el.dispatchEvent(new Event("input"))
  await flushPromises()
}
const submit = async () => {
  document.body.querySelector<HTMLFormElement>("#erase-account-form")!.dispatchEvent(new Event("submit"))
  await flushPromises()
}

describe("SettingsEraseAccount", () => {
  it("explains the grace period and keeps confirm disabled until re-auth is typed", async () => {
    await open()
    expect(document.body.textContent).toContain("30")
    const confirm = document.body.querySelector<HTMLButtonElement>('button[form="erase-account-form"]')!
    expect(confirm.disabled).toBe(true)
    await type('input[type="password"]', "pw")
    expect(confirm.disabled).toBe(false)
  })

  it("flips to the username check on CONFIRM_MISMATCH and shows the error", async () => {
    await open()
    m.requestErasure.mockRejectedValueOnce(err(400, "CONFIRM_MISMATCH"))
    await type('input[type="password"]', "x")
    await submit()
    expect(document.body.querySelector('input[type="password"]')).toBeNull()
    expect(m.reset).not.toHaveBeenCalled()
  })

  it("shows a specific error when the last admin is refused", async () => {
    await open()
    m.requestErasure.mockRejectedValueOnce(err(409, "FINAL_ADMIN"))
    await type('input[type="password"]', "x")
    await submit()
    expect(document.body.textContent).toContain(i18n.global.t("settings.danger.errors.finalAdmin"))
    expect(m.reset).not.toHaveBeenCalled()
  })

  it("signs out, shows the erasure date and goes to login on success", async () => {
    m.requestErasure.mockResolvedValue({ data: { erasure_due_at: "2027-01-05T00:00:00Z" } })
    await open()
    await type('input[type="password"]', "pw")
    await submit()
    expect(m.requestErasure).toHaveBeenLastCalledWith({ password: "pw" })
    expect(m.reset).toHaveBeenCalled()
    expect(useAuthStore().user).toBeNull()
    const date = new Intl.DateTimeFormat(i18n.global.locale.value, { dateStyle: "long" }).format(new Date("2027-01-05T00:00:00Z"))
    expect(m.success).toHaveBeenCalledWith(expect.stringContaining(date))
    expect(m.push).toHaveBeenCalledWith({ name: "login" })
  })

  it("is rendered inside the danger section", () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const w = mount(SettingsDangerSection, { props: { user }, global: { plugins: [i18n, pinia] } })
    expect(w.findComponent(SettingsEraseAccount).exists()).toBe(true)
  })
})
