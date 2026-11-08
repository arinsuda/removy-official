import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import HandoffView from "@/views/auth/HandoffView.vue"
import { linkApi } from "@/api/endpoints/link"
import { safeWebUrl } from "@/utils/safeWebUrl"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/link", () => ({
  linkApi: { remote: vi.fn(), redeem: vi.fn(), confirm: vi.fn(), sendCode: vi.fn() },
}))
vi.mock("@/utils/platformTransition", () => ({ playPlatformEnter: vi.fn() }))

const i18n = useRealI18n()
i18n.global.locale.value = "en"

async function render(hash: string, web: string | Error) {
  vi.mocked(linkApi.remote).mockImplementation(() =>
    web instanceof Error ? Promise.reject(web) : Promise.resolve({ data: { web_url: web } } as never),
  )
  window.location.hash = hash
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: "/", name: "home", component: { template: "<div />" } }, { path: "/login", component: { template: "<div />" } }],
  })
  const w = mount(HandoffView, { global: { plugins: [createPinia(), i18n, router] } })
  await flushPromises()
  return w
}

const confirmRequired = {
  response: { status: 409, data: { code: "LINK_CONFIRM_REQUIRED", pending_id: "p", email: "a@b.c", methods: ["local_password"] } },
}

describe("HandoffView back to PLANZI", () => {
  beforeEach(() => vi.clearAllMocks())

  it("shows the link in the confirm state without the token", async () => {
    vi.mocked(linkApi.redeem).mockRejectedValue(confirmRequired)
    const w = await render("#t=SECRET", "https://planzi.example")
    const a = w.get("a.handoff-back")
    expect(a.attributes("href")).toBe("https://planzi.example")
    expect(a.attributes("href")).not.toContain("SECRET")
    expect(a.attributes("rel")).toContain("noopener")
    expect(a.text()).toContain("Back to PLANZI")
  })

  it("shows the link in the error state", async () => {
    const w = await render("", "https://planzi.example")
    expect(w.get("a.handoff-back").attributes("href")).toBe("https://planzi.example")
  })

  it("hides the link for invalid or absent urls", async () => {
    expect((await render("", "javascript:alert(1)")).find("a.handoff-back").exists()).toBe(false)
    expect((await render("", new Error("404"))).find("a.handoff-back").exists()).toBe(false)
  })

  it("only accepts http(s)", () => {
    expect(safeWebUrl("http://x.test")).toBe("http://x.test")
    expect(safeWebUrl("ftp://x.test")).toBe("")
    expect(safeWebUrl("nope")).toBe("")
    expect(safeWebUrl(undefined)).toBe("")
  })
})
