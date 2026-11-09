import { describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import SettingsDangerSection from "@/components/settings/SettingsDangerSection.vue"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ deleteUser: vi.fn(), push: vi.fn(), reset: vi.fn() }))
vi.mock("@/api/api", () => ({ userApi: { deleteUser: m.deleteUser }, authApi: {} }))
vi.mock("vue-router", () => ({ useRouter: () => ({ push: m.push }) }))
vi.mock("@/composables/useSessionReset", () => ({ useSessionReset: () => m.reset }))
vi.mock("@/composables/useSocket", () => ({ resume: vi.fn() }))

const i18n = useRealI18n()
const err = (status: number, code: string) => ({ response: { status, data: { code } } })

async function open() {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: 1, username: "arin" } as never
  const w = mount(SettingsDangerSection, { props: { user: { id: 1, username: "arin" } as never }, attachTo: document.body, global: { plugins: [i18n, pinia] } })
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
  document.body.querySelector<HTMLFormElement>("#delete-account-form")!.dispatchEvent(new Event("submit"))
  await flushPromises()
}

describe("SettingsDangerSection", () => {
  it("switches to username mode on CONFIRM_MISMATCH and back to password on WRONG_PASSWORD", async () => {
    await open()
    m.deleteUser.mockRejectedValueOnce(err(400, "CONFIRM_MISMATCH"))
    await type('input[type="password"]', "x")
    await submit()
    expect(document.body.querySelector('input[type="password"]')).toBeNull()
    m.deleteUser.mockRejectedValueOnce(err(400, "WRONG_PASSWORD"))
    await type("input", "arin")
    await submit()
    expect(document.body.querySelector('input[type="password"]')).not.toBeNull()
  })

  it("tears the session down locally on success and goes to login", async () => {
    m.deleteUser.mockResolvedValue({})
    await open()
    await type('input[type="password"]', "pw")
    await submit()
    expect(m.deleteUser).toHaveBeenLastCalledWith(1, { password: "pw" })
    expect(m.reset).toHaveBeenCalled()
    expect(useAuthStore().user).toBeNull()
    expect(m.push).toHaveBeenCalledWith({ name: "login" })
  })
})
