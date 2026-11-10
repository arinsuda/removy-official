import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createMemoryHistory, createRouter, RouterView } from "vue-router"
import { h } from "vue"
import SettingsView from "@/views/settings/SettingsView.vue"
import { useAuthStore } from "@/stores/auth"
import { useRealI18n } from "./helpers/realI18n"

const m = vi.hoisted(() => ({ getProfile: vi.fn(), updateProfile: vi.fn(), updateFavoriteGenres: vi.fn() }))
vi.mock("@/api/api", () => ({
  userApi: m,
  authApi: { getGoogleStatus: vi.fn().mockResolvedValue({ data: { enabled: true, google_connected: false, can_disconnect: false } }) },
}))
vi.mock("@/api/endpoints/feed", () => ({ feedApi: { getSettings: vi.fn().mockResolvedValue({ data: {} }), updateSettings: vi.fn() } }))
vi.mock("@/api/endpoints/link", () => ({ linkApi: { status: vi.fn().mockResolvedValue({ data: { links: [] } }), unlink: vi.fn() } }))

const i18n = useRealI18n()
const profile = {
  id: 1, username: "alice", email: "a@x.test", verified_email_at: "2024-01-01T00:00:00Z", display_name: "Alice", bio: "hello",
  avatar_url: null, date_of_birth: null, gender: "", favorite_genres: "[28]", is_private: false, level: 1,
}

async function mountSettings() {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: 1, favorite_genres: "[28]" } as never
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/settings", component: SettingsView },
      { path: "/elsewhere", component: { template: "<div />" } },
    ],
  })
  await router.push("/settings")
  await router.isReady()
  const wrapper = mount({ render: () => h(RouterView) }, {
    global: {
      plugins: [i18n, pinia, router, [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]],
      stubs: { SettingsEmailSection: true, SettingsPasswordSection: true, SettingsConnectionsSection: true, SettingsPreferencesSection: true },
    },
    attachTo: document.body,
  })
  await flushPromises()
  return { wrapper, router }
}

beforeEach(() => {
  vi.clearAllMocks()
  document.body.innerHTML = ""
  m.getProfile.mockResolvedValue({ data: { user: profile } })
  m.updateProfile.mockResolvedValue({ data: { user: { ...profile, bio: "" } } })
})

const t = (k: string) => i18n.global.t(k)

describe("SettingsView", () => {
  it("shows the save bar only after an edit, and discarding hides it", async () => {
    const { wrapper } = await mountSettings()
    expect(wrapper.find('[role="region"]').exists()).toBe(false)
    await wrapper.find("textarea").setValue("changed")
    expect(wrapper.find('[role="region"]').exists()).toBe(true)
    const discard = wrapper.findAll("button").find(b => b.text() === t("settings.saveBar.discard"))!
    await discard.trigger("click")
    expect(wrapper.find('[role="region"]').exists()).toBe(false)
  })

  it("keeps the username read-only and clears the bio by sending an empty key", async () => {
    const { wrapper } = await mountSettings()
    const username = wrapper.findAll("input").find(i => (i.element as HTMLInputElement).value === "@alice")!
    expect(username.attributes("disabled")).toBeDefined()
    await wrapper.find("textarea").setValue("")
    await wrapper.findAll("button").find(b => b.text() === t("settings.saveBar.save"))!.trigger("click")
    await flushPromises()
    const body = m.updateProfile.mock.calls[0]![1] as FormData
    expect(body.get("bio")).toBe("")
    expect(body.has("username")).toBe(false)
  })

  it("blocks leaving with unsaved changes unless confirmed in AppModal", async () => {
    const { wrapper, router } = await mountSettings()
    await wrapper.find("textarea").setValue("changed")

    const nav1 = router.push("/elsewhere")
    await flushPromises()
    const dialog1 = document.body.querySelector('[role="dialog"]')!
    expect(dialog1).not.toBeNull()
    expect(dialog1.textContent).toContain(t("settings.unsavedConfirm"))
    const cancelBtn = Array.from(dialog1.querySelectorAll("button")).find(b => b.textContent === t("common.cancel"))!
    cancelBtn.click()
    await nav1
    await flushPromises()
    expect(router.currentRoute.value.path).toBe("/settings")

    const nav2 = router.push("/elsewhere")
    await flushPromises()
    const dialog2 = document.body.querySelector('[role="dialog"]')!
    const discardBtn = Array.from(dialog2.querySelectorAll("button")).find(b => b.textContent === t("settings.saveBar.discard"))!
    discardBtn.click()
    await nav2
    await flushPromises()
    expect(router.currentRoute.value.path).toBe("/elsewhere")
  })
})
