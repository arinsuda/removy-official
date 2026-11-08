import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import LoginView from "@/views/auth/LoginView.vue"
import { authApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", () => ({
  authApi: { resendVerification: vi.fn(), getGoogleConfig: vi.fn(), getGoogleLoginUrl: vi.fn() },
}))

const i18n = useRealI18n()

async function submit(identifier: string, err: unknown) {
  const pinia = createPinia()
  setActivePinia(pinia)
  vi.spyOn(useAuthStore(), "login").mockRejectedValue(err)
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: "/", component: { template: "<div />" } }] })
  const w = mount(LoginView, { global: { plugins: [pinia, i18n, router] } })
  const inputs = w.findAll("input")
  await inputs[0].setValue(identifier)
  await inputs[1].setValue("pw")
  await w.find("form").trigger("submit")
  await flushPromises()
  return w
}

const forbidden = { response: { status: 403, data: { error: "please verify your email first" } } }

describe("LoginView unverified email", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    i18n.global.locale.value = "en"
  })

  it("explains the 403 and resends to the typed email", async () => {
    const w = await submit("a@b.co", forbidden)
    expect(w.text()).toContain("Please verify your email first.")
    const btn = w.findAll("button").find((b) => b.text() === "Resend verification email")!
    await btn.trigger("click")
    await flushPromises()
    expect(authApi.resendVerification).toHaveBeenCalledWith("a@b.co")
    expect(useUIStore().toasts.at(-1)?.message).toBe("If the email exists, we've sent a link")
  })

  it("asks for an email when a username was typed", async () => {
    const w = await submit("someuser", forbidden)
    const btn = w.findAll("button").find((b) => b.text() === "Resend verification email")!
    expect(btn.attributes("disabled")).toBeDefined()
    await w.findAll("input")[2].setValue("me@b.co")
    await btn.trigger("click")
    await flushPromises()
    expect(authApi.resendVerification).toHaveBeenCalledWith("me@b.co")
  })

  it("shows a cooldown message on 429", async () => {
    vi.mocked(authApi.resendVerification).mockRejectedValue({ response: { status: 429 } })
    const w = await submit("a@b.co", forbidden)
    await w.findAll("button").find((b) => b.text() === "Resend verification email")!.trigger("click")
    await flushPromises()
    expect(w.text()).toContain("Too many requests")
  })

  it("does not offer resend for other errors", async () => {
    const w = await submit("a@b.co", { response: { status: 401 } })
    expect(w.text()).not.toContain("Resend verification email")
  })
})
