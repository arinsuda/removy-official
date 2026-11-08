import { afterEach, describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import AboutUsView from "@/views/aboutus/AboutUsView.vue"
import AboutCreator from "@/components/about/AboutCreator.vue"
import { useAuthStore } from "@/stores/auth"
import en from "@/i18n/locales/en/about.json"
import th from "@/i18n/locales/th/about.json"
import { expectKeysInBothLocales, useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()

function setup(loggedIn: boolean) {
  const pinia = createPinia()
  setActivePinia(pinia)
  if (loggedIn) useAuthStore().user = { id: 7 } as never
  const stub = { template: "<div />" }
  const router = createRouter({
    history: createMemoryHistory(),
    routes: ["home", "register", "donate", "movies", "feed"].map((name) => ({ path: `/${name}`, name, component: stub })).concat([
      { path: "/users/:userId/library", name: "user-library", component: stub },
      { path: "/users/:userId/achievements", name: "user-achievements", component: stub },
    ]),
  })
  return { plugins: [pinia, i18n, router] }
}

describe("About page", () => {
  afterEach(() => i18n.global.locale.value = "th")

  it("has its keys in both locales", () => {
    expectKeysInBothLocales(["about.hero.cta.home", "about.hero.cta.register", "about.creator.avatarAlt", "about.creator.links.email", "about.builtWith"])
    expect(Object.keys(en)).toEqual(Object.keys(th))
  })

  it("sends guests to register and signed-in users home", () => {
    i18n.global.locale.value = "en"
    const guest = mount(AboutUsView, { global: setup(false) })
    const guestCta = guest.get(".about-hero a")
    expect(guestCta.attributes("href")).toBe("/register")
    expect(guestCta.text()).toContain(en.hero.cta.register)

    const member = mount(AboutUsView, { global: setup(true) })
    const memberCta = member.get(".about-hero a")
    expect(memberCta.attributes("href")).toBe("/home")
    expect(memberCta.text()).toContain(en.hero.cta.home)
    expect(member.findAll(".feature").map((a) => a.attributes("href"))).toContain("/users/7/library")
  })

  it("opens external creator links safely and labels the avatar in the active locale", () => {
    i18n.global.locale.value = "th"
    const w = mount(AboutCreator, { global: setup(false) })
    for (const a of w.findAll('a[target="_blank"]')) expect(a.attributes("rel")).toBe("noopener noreferrer")
    expect(w.findAll('a[target="_blank"]')).toHaveLength(2)
    expect(w.get('[role="img"]').attributes("aria-label")).toBe(th.creator.avatarAlt)
    expect(w.text()).toContain(th.creator.links.email)
  })

  it("drops the fake stats", () => {
    const w = mount(AboutUsView, { global: setup(false) })
    expect(w.text()).not.toMatch(/10K\+|5K\+|∞/)
  })
})
