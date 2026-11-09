import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import ForgotPasswordView from "@/views/auth/ForgotPasswordView.vue"
import ResetPasswordView from "@/views/auth/ResetPasswordView.vue"
import VerifyEmailView from "@/views/auth/VerifyEmailView.vue"
import CheckEmailView from "@/views/auth/CheckEmailView.vue"
import HandoffView from "@/views/auth/HandoffView.vue"
import { authApi } from "@/api/api"
import { linkApi } from "@/api/endpoints/link"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", () => ({
  authApi: { forgotPassword: vi.fn(), resetPassword: vi.fn(), verifyEmail: vi.fn(), resendVerification: vi.fn() },
}))
vi.mock("@/api/endpoints/link", () => ({
  linkApi: { remote: vi.fn(), redeem: vi.fn(), confirm: vi.fn(), sendCode: vi.fn() },
}))
vi.mock("@/utils/platformTransition", () => ({ playPlatformEnter: vi.fn() }))

const i18n = useRealI18n()

async function mountAt(view: object, path: string) {
  setActivePinia(createPinia())
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: "/:p(.*)*", name: "home", component: { template: "<div />" } }],
  })
  await router.push(path)
  const w = mount(view, { global: { plugins: [createPinia(), i18n, router] } })
  await flushPromises()
  return w
}

const api = vi.mocked(authApi)

beforeEach(() => {
  vi.clearAllMocks()
  i18n.global.locale.value = "en"
})

describe("forgot password", () => {
  it("renders inside the auth shell and blocks invalid emails", async () => {
    const w = await mountAt(ForgotPasswordView, "/forgot-password")
    expect(w.get("h1").text()).toBe("Forgot Password")
    expect(w.find(".auth-brand").exists()).toBe(true)
    await w.get("input").setValue("nope")
    await w.get("form").trigger("submit")
    expect(w.get(".field__error").text()).toBe("Invalid email address format")
    expect(api.forgotPassword).not.toHaveBeenCalled()
  })

  it("sends the link and shows the sent state with a resend cooldown", async () => {
    api.forgotPassword.mockResolvedValue({} as never)
    const w = await mountAt(ForgotPasswordView, "/forgot-password")
    await w.get("input").setValue("a@b.co")
    await w.get("form").trigger("submit")
    await flushPromises()
    expect(api.forgotPassword).toHaveBeenCalledWith({ email: "a@b.co" })
    expect(w.get("h1").text()).toBe("Check your email")
    expect(w.text()).toContain("a@b.co")
    expect(w.text()).toContain("Resend in 60s")
    w.unmount()
  })

  it("renders Thai copy", async () => {
    i18n.global.locale.value = "th"
    const w = await mountAt(ForgotPasswordView, "/forgot-password")
    expect(w.get("h1").text()).toBe("ลืมรหัสผ่าน")
  })
})

describe("reset password", () => {
  it("shows the expired state without a token", async () => {
    const w = await mountAt(ResetPasswordView, "/reset-password")
    expect(w.get("h1").text()).toBe("Link expired")
    expect(w.findAll("a").map((a) => a.attributes("href"))).toContain("/forgot-password")
  })

  it("validates mismatch and submits the token and uid from the query", async () => {
    api.resetPassword.mockResolvedValue({} as never)
    const w = await mountAt(ResetPasswordView, "/reset-password?token=T&uid=7")
    const [a, b] = w.findAll("input")
    await a!.setValue("Passw0rdOK")
    await b!.setValue("Different1")
    expect(w.get(".field__error").text()).toBe("Passwords do not match")
    expect(w.get('button[type="submit"]').attributes("disabled")).toBeDefined()
    await b!.setValue("Passw0rdOK")
    await w.get("form").trigger("submit")
    await flushPromises()
    expect(api.resetPassword).toHaveBeenCalledWith({
      token: "T",
      user_id: 7,
      new_password: "Passw0rdOK",
      confirm_password: "Passw0rdOK",
    })
    expect(w.get("h1").text()).toBe("Password updated")
  })

  it("falls back to the expired state when the server rejects the token", async () => {
    api.resetPassword.mockRejectedValue({ response: { status: 410 } })
    const w = await mountAt(ResetPasswordView, "/reset-password?token=T&uid=7")
    const [a, b] = w.findAll("input")
    await a!.setValue("Passw0rdOK")
    await b!.setValue("Passw0rdOK")
    await w.get("form").trigger("submit")
    await flushPromises()
    expect(w.get("h1").text()).toBe("Link expired")
  })
})

describe("verify email", () => {
  it("does not verify on open and asks for the password", async () => {
    const w = await mountAt(VerifyEmailView, "/verify-email?token=T")
    expect(api.verifyEmail).not.toHaveBeenCalled()
    expect(w.findAll("label").map((l) => l.text())).toEqual(["Password"])
  })

  it("keeps the form and token after a wrong password", async () => {
    api.verifyEmail.mockRejectedValue({ response: { status: 400, data: { code: "WRONG_PASSWORD" } } })
    const w = await mountAt(VerifyEmailView, "/verify-email?token=T")
    await w.get("input").setValue("nope")
    await w.get("form").trigger("submit")
    await flushPromises()
    expect(api.verifyEmail).toHaveBeenCalledWith("T", { password: "nope" })
    expect(w.get('[role="alert"]').text()).toContain("That password is not correct")
    expect(w.find("form").exists()).toBe(true)
  })

  it("requires an explicit confirmation before skipping the password", async () => {
    api.verifyEmail.mockResolvedValue({ data: { password_reset_required: true } } as never)
    const w = await mountAt(VerifyEmailView, "/verify-email?token=T")
    await w.get("button.auth-link").trigger("click")
    expect(api.verifyEmail).not.toHaveBeenCalled()
    expect(w.get('[role="alertdialog"]').text()).toContain("cannot be undone")
    const confirm = w.findAll("button").find((b) => b.text().includes("remove my old sign-in methods"))!
    await confirm.trigger("click")
    await flushPromises()
    expect(api.verifyEmail).toHaveBeenCalledWith("T", { skipPassword: true })
    expect(w.get("h1").text()).toBe("Email verified")
    expect(w.text()).toContain("Set a new password")
  })

  it("shows the attempts-exceeded message with a resend option", async () => {
    api.verifyEmail.mockRejectedValue({ response: { status: 400, data: { code: "VERIFY_ATTEMPTS_EXCEEDED" } } })
    const w = await mountAt(VerifyEmailView, "/verify-email?token=T")
    await w.get("input").setValue("x")
    await w.get("form").trigger("submit")
    await flushPromises()
    expect(w.get("h1").text()).toBe("Link expired")
    expect(w.text()).toContain("Too many wrong passwords")
    expect(w.text()).toContain("Resend link")
  })

  it("shows the expired state when the token is missing", async () => {
    const w = await mountAt(VerifyEmailView, "/verify-email")
    expect(w.get("h1").text()).toBe("Link expired")
  })
})

describe("check email", () => {
  it("renders and resends to the known address", async () => {
    api.resendVerification.mockResolvedValue({} as never)
    const w = await mountAt(CheckEmailView, "/check-email")
    expect(w.get("h1").text()).toBe("Check Your Email")
    vi.stubGlobal("prompt", vi.fn().mockReturnValue("me@x.co"))
    await w.get("button").trigger("click")
    await flushPromises()
    expect(api.resendVerification).toHaveBeenCalledWith("me@x.co")
    expect(w.get("button").text()).toBe("Sent successfully")
  })
})

describe("handoff", () => {
  it("renders the confirm state in the shell with labelled inputs and clears the token", async () => {
    vi.mocked(linkApi.remote).mockResolvedValue({ data: { web_url: "https://planzi.example" } } as never)
    vi.mocked(linkApi.redeem).mockRejectedValue({
      response: {
        status: 409,
        data: { code: "LINK_CONFIRM_REQUIRED", pending_id: "p", email: "a@b.c", methods: ["local_password", "email_code"] },
      },
    })
    window.location.hash = "#t=SECRET"
    const w = await mountAt(HandoffView, "/")
    expect(w.get("h1").text()).toBe("Confirm it's you")
    expect(w.find(".auth-brand").exists()).toBe(true)
    expect(w.findAll("label").map((l) => l.text())).toEqual(["POPCORN password"])
    expect(window.location.hash).toBe("")
    expect(linkApi.redeem).toHaveBeenCalledWith("SECRET")
    expect(w.text()).toContain("Send code to my email")
  })

  it("renders the failed state with a back-to-login button", async () => {
    vi.mocked(linkApi.remote).mockRejectedValue(new Error("x"))
    window.location.hash = ""
    const w = await mountAt(HandoffView, "/")
    expect(w.get("h1").text()).toBe("Link problem")
    expect(w.get('[role="alert"]').text()).toContain("invalid or has expired")
    expect(w.findAll("a").map((a) => a.attributes("href"))).toContain("/login")
  })
})
