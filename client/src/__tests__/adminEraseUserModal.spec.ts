import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import AdminEraseUserModal from "@/components/admin/AdminEraseUserModal.vue"
import { adminApi } from "@/api/api"
import type { AdminUserRow } from "@/types"
import { expectKeysInBothLocales, useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", async (orig) => ({
  ...(await orig<typeof import("@/api/api")>()),
  adminApi: { eraseUser: vi.fn() },
}))

const i18n = useRealI18n()
const api = vi.mocked(adminApi)
const wrappers: VueWrapper[] = []
const origGetClientRects = HTMLElement.prototype.getClientRects

const target: AdminUserRow = {
  id: 7,
  username: "Jane_Doe",
  email: "jane@x.co",
  display_name: null,
  avatar_url: null,
  role: "user",
  is_active: false,
  created_at: "2026-01-02T00:00:00Z",
  deleted_at: "2026-02-03T00:00:00Z",
  review_count: 0,
}

beforeEach(() => {
  HTMLElement.prototype.getClientRects = function (this: HTMLElement) {
    return (this.isConnected ? [{ width: 1, height: 1 }] : []) as unknown as DOMRectList
  }
  vi.clearAllMocks()
  i18n.global.locale.value = "en"
  setActivePinia(createPinia())
})

afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  HTMLElement.prototype.getClientRects = origGetClientRects
})

function open() {
  const w = mount(AdminEraseUserModal, {
    props: { modelValue: true, user: target },
    global: { plugins: [i18n] },
    attachTo: document.body,
  })
  wrappers.push(w)
  return w
}

const input = () => document.body.querySelector<HTMLInputElement>("input")!
const confirmButton = () =>
  Array.from(document.body.querySelectorAll<HTMLButtonElement>("button")).find((b) => b.textContent?.includes("Erase permanently"))!

async function type(value: string) {
  input().value = value
  input().dispatchEvent(new Event("input"))
  await flushPromises()
}

describe("AdminEraseUserModal", () => {
  it("keeps the confirm button disabled until the username matches exactly", async () => {
    open()
    await flushPromises()
    expect(confirmButton().disabled).toBe(true)
    await type("jane_doe")
    expect(confirmButton().disabled).toBe(true)
    await type("Jane_Doe ")
    expect(confirmButton().disabled).toBe(true)
    await type("Jane_Doe")
    expect(confirmButton().disabled).toBe(false)
    expect(api.eraseUser).not.toHaveBeenCalled()
  })

  it("sends the username and trimmed reason, then closes and emits erased", async () => {
    api.eraseUser.mockResolvedValue({} as never)
    const w = open()
    await flushPromises()
    await type("Jane_Doe")
    const reason = document.body.querySelector<HTMLTextAreaElement>("textarea")!
    reason.value = "  spam  "
    reason.dispatchEvent(new Event("input"))
    await flushPromises()
    confirmButton().click()
    await flushPromises()
    expect(api.eraseUser).toHaveBeenCalledWith(7, { confirm_username: "Jane_Doe", reason: "spam" })
    expect(w.emitted("erased")?.[0]).toEqual([target])
    expect(w.emitted("update:modelValue")?.at(-1)).toEqual([false])
  })

  it.each([
    ["CONFIRMATION_MISMATCH", 400, "The username you typed doesn't match.", false],
    ["USER_NOT_REMOVED", 409, "Delete this account first. Only deleted accounts can be erased.", true],
    ["SELF_ERASURE", 403, "You can't erase your own account.", false],
    ["USER_NOT_FOUND", 404, "That item no longer exists. The list will refresh.", true],
  ])("maps %s to a localized inline error", async (code, status, message, stale) => {
    api.eraseUser.mockRejectedValue({ response: { status, data: { code, error: "x" } } })
    const w = open()
    await flushPromises()
    await type("Jane_Doe")
    confirmButton().click()
    await flushPromises()
    expect(document.body.querySelector('[role="alert"]')!.textContent).toContain(message)
    expect(!!w.emitted("stale")).toBe(stale)
    expect(w.emitted("erased")).toBeUndefined()
  })

  it("has every new key in both locales", () => {
    expectKeysInBothLocales([
      "admin.users.statusDeleted",
      "admin.users.actionErase",
      "admin.users.deletedOn",
      "admin.users.badgeDeleted",
      "admin.userModal.eraseTitle",
      "admin.userModal.eraseDesc",
      "admin.userModal.eraseWarning",
      "admin.userModal.eraseConfirmLabel",
      "admin.userModal.eraseConfirmHint",
      "admin.userModal.eraseConfirm",
      "admin.userModal.eraseReasonLabel",
      "admin.userModal.eraseReasonNote",
      "admin.userModal.erased",
      "admin.errors.selfErase",
      "admin.errors.notRemoved",
      "admin.errors.confirmMismatch",
      "admin.auditLogs.action.USER_ERASED",
    ])
  })
})
