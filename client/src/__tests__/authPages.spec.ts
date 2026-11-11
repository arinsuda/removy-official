import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import LoginView from "@/views/auth/LoginView.vue"
import RegisterView from "@/views/auth/RegisterView.vue"
import PasswordStrength from "@/components/auth/PasswordStrength.vue"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", () => ({ authApi: { getGoogleConfig: vi.fn(), getGoogleLoginUrl: vi.fn(), resendVerification: vi.fn() } }))
vi.mock("gsap", () => ({
  gsap: {
    timeline: () => ({
      from: vi.fn(function (this: unknown) {
        return this
      }),
    }),
    fromTo: vi.fn(),
    to: vi.fn((_el: unknown, vars?: { onComplete?: () => void }) => {
      vars?.onComplete?.()
      return Promise.resolve()
    }),
  },
}))

const i18n = useRealI18n()

function mountView(view: typeof LoginView | typeof RegisterView) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: "/:p(.*)*", component: { template: "<div />" } }] })
  return { w: mount(view, { global: { plugins: [pinia, i18n, router] } }), pinia }
}

async function mountAt(view: typeof LoginView | typeof RegisterView, url: string) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/", name: "home", component: { template: "<div />" } },
      { path: "/login", name: "login", component: LoginView },
      { path: "/register", name: "register", component: RegisterView },
      { path: "/movies/:id", name: "movie-detail", component: { template: "<div />" } },
      { path: "/:p(.*)*", component: { template: "<div />" } },
    ],
  })
  await router.push(url)
  await router.isReady()
  return { w: mount(view, { global: { plugins: [pinia, i18n, router] } }), pinia, router }
}

describe("auth pages", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    i18n.global.locale.value = "en"
  })

  it("renders the login shell with labelled fields, forgot and register links", () => {
    const { w } = mountView(LoginView)
    expect(w.get("h1").text()).toBe("Log In")
    expect(w.findAll("label").map((l) => l.text())).toEqual(["Username or email", "Password"])
    expect(w.findAll("a").map((a) => a.attributes("href"))).toEqual(expect.arrayContaining(["/forgot-password", "/register"]))
    expect(w.text()).toContain("Continue with Google")
  })

  it("renders Thai login copy", () => {
    i18n.global.locale.value = "th"
    const { w } = mountView(LoginView)
    expect(w.get("h1").text()).toBe("เข้าสู่ระบบ")
  })

  it("shows a localized message for empty login without calling the store", async () => {
    const { w } = mountView(LoginView)
    const login = vi.spyOn(useAuthStore(), "login")
    await w.get("form").trigger("submit")
    expect(w.get('[role="alert"]').text()).toContain("Please fill in all fields")
    expect(login).not.toHaveBeenCalled()
  })

  it("shows field errors on register and does not submit", async () => {
    const { w } = mountView(RegisterView)
    const register = vi.spyOn(useAuthStore(), "register")
    await w.get("form").trigger("submit")
    await flushPromises()
    const errors = w.findAll(".field__error").map((e) => e.text())
    expect(errors).toEqual(["Username is required", "Email is required", "Password is required", "Please confirm your password"])
    expect(register).not.toHaveBeenCalled()
    const user = w.findAll("input")[0]!
    expect(user.attributes("aria-invalid")).toBe("true")
  })

  it("scores password strength with a localized label", async () => {
    const w = mount(PasswordStrength, { props: { password: "Abcdefg1!xyz" }, global: { plugins: [i18n] } })
    expect(w.text()).toBe("Strong")
    await w.setProps({ password: "abc" })
    expect(w.text()).toBe("Weak")
  })

  it("redirects after login only to validated same-origin paths and forwards safe query across login/register links", async () => {
    const { w, router } = await mountAt(LoginView, "/login?redirect=%2Fmovies%2F603%3Ftab%3Dreviews")
    const registerUrl = new URL(w.get('a[href^="/register"]').attributes("href") ?? "", "http://localhost")
    expect(registerUrl.pathname).toBe("/register")
    expect(registerUrl.searchParams.get("redirect")).toBe("/movies/603?tab=reviews")

    const authStore = useAuthStore()
    vi.spyOn(authStore, "login").mockResolvedValue({ id: 1 } as never)
    const inputs = w.findAll("input")
    await inputs[0]!.setValue("user")
    await inputs[1]!.setValue("secret")
    await w.get("form").trigger("submit")
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe("/movies/603?tab=reviews")
  })

  it("accepts return_url as a fallback and sanitizes it", async () => {
    const ok = await mountAt(LoginView, "/login?return_url=%2Fmovies%2F603")
    const url = new URL(ok.w.get('a[href^="/register"]').attributes("href") ?? "", "http://localhost")
    expect(url.searchParams.get("redirect")).toBe("/movies/603")
    const bad = await mountAt(LoginView, "/login?return_url=//evil.com")
    expect(bad.w.get('a[href^="/register"]').attributes("href")).toBe("/register")
  })

  it("drops unsafe protocol-relative redirects on login and preserves safe links on register", async () => {
    const { w, router } = await mountAt(LoginView, "/login?redirect=//evil.com")
    expect(w.get('a[href^="/register"]').attributes("href")).toBe("/register")

    const authStore = useAuthStore()
    vi.spyOn(authStore, "login").mockResolvedValue({ id: 1 } as never)
    const inputs = w.findAll("input")
    await inputs[0]!.setValue("user")
    await inputs[1]!.setValue("secret")
    await w.get("form").trigger("submit")
    await flushPromises()
    expect(router.currentRoute.value.name).toBe("home")

    const reg = await mountAt(RegisterView, "/register?redirect=%2Factors%2F100")
    const loginUrl = new URL(reg.w.get('a[href^="/login"]').attributes("href") ?? "", "http://localhost")
    expect(loginUrl.pathname).toBe("/login")
    expect(loginUrl.searchParams.get("redirect")).toBe("/actors/100")
  })
})
