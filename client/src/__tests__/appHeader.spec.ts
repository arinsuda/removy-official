import { afterEach, beforeEach, describe, expect, it } from "vitest"
import { BREAKPOINTS } from "../constants/breakpoints"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import { i18n } from "../i18n"
import AppHeader from "../components/layout/AppHeader.vue"
import AccountMenu from "../components/layout/AccountMenu.vue"
import HeaderNav from "../components/layout/HeaderNav.vue"
import MobileNavSheet from "../components/layout/MobileNavSheet.vue"
import { formatBadge } from "../components/layout/navItems"
import { useAuthStore } from "../stores/auth"
import { useNotificationStore } from "../stores/notification"
import type { AuthUser } from "../types/auth"

const user = { id: 7, username: "arin", display_name: "Arin S", avatar_url: null, level: 3, role: "user" } as AuthUser

const stub = { template: "<div />" }
function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/", name: "home", component: stub },
      { path: "/movies", component: stub },
      { path: "/tv", component: stub },
      { path: "/upcoming", name: "upcoming", component: stub },
      { path: "/about-us", name: "aboutUs", component: stub },
      { path: "/feed", name: "feed", component: stub },
      { path: "/users/:id", component: stub },
      { path: "/users/:id/library", component: stub },
      { path: "/users/:id/achievements", component: stub },
      { path: "/login", name: "login", component: stub },
      { path: "/register", name: "register", component: stub },
    ],
  })
}

describe("formatBadge", () => {
  it("hides 0, shows counts, caps at 99+", () => {
    expect(formatBadge(0)).toBeNull()
    expect(formatBadge(5)).toBe("5")
    expect(formatBadge(99)).toBe("99")
    expect(formatBadge(100)).toBe("99+")
  })
})

describe("HeaderNav", () => {
  it("marks only the current route with aria-current", async () => {
    const router = makeRouter()
    await router.push("/movies")
    const w = mount(HeaderNav, { global: { plugins: [i18n, router] } })
    const current = w.findAll("[aria-current='page']")
    expect(current).toHaveLength(1)
    expect(current[0]!.attributes("href")).toBe("/movies")
    w.unmount()
  })
})

describe("AccountMenu", () => {
  let w: VueWrapper
  const menuItems = () => Array.from(document.querySelectorAll<HTMLElement>('[role="menuitem"]'))

  beforeEach(async () => {
    const router = makeRouter()
    await router.push("/")
    w = mount(AccountMenu, { props: { user, isAdmin: false }, global: { plugins: [i18n, router] }, attachTo: document.body })
  })
  afterEach(() => w.unmount())

  it("opens with the trigger, exposes aria state and focuses the first item", async () => {
    const trigger = w.find("button[aria-haspopup='menu']")
    expect(trigger.attributes("aria-expanded")).toBe("false")
    await trigger.trigger("click")
    await flushPromises()
    expect(trigger.attributes("aria-expanded")).toBe("true")
    expect(document.activeElement).toBe(menuItems()[0])
    expect(w.text()).toContain("@arin")
  })

  it("moves focus with Arrow, Home and End keys and wraps", async () => {
    await w.find("button[aria-haspopup='menu']").trigger("click")
    await flushPromises()
    const items = menuItems()
    const menu = w.find("[role='menu']")
    await menu.trigger("keydown", { key: "ArrowDown" })
    expect(document.activeElement).toBe(items[1])
    await menu.trigger("keydown", { key: "End" })
    expect(document.activeElement).toBe(items[items.length - 1])
    await menu.trigger("keydown", { key: "ArrowDown" })
    expect(document.activeElement).toBe(items[0])
    await menu.trigger("keydown", { key: "ArrowUp" })
    expect(document.activeElement).toBe(items[items.length - 1])
    await menu.trigger("keydown", { key: "Home" })
    expect(document.activeElement).toBe(items[0])
  })

  it("closes on Escape and returns focus to the trigger", async () => {
    const trigger = w.find("button[aria-haspopup='menu']")
    await trigger.trigger("click")
    await flushPromises()
    await w.find("[role='menu']").trigger("keydown", { key: "Escape" })
    await flushPromises()
    expect(w.find("[role='menu']").exists()).toBe(false)
    expect(document.activeElement).toBe(trigger.element)
  })

  it("closes on outside click", async () => {
    await w.find("button[aria-haspopup='menu']").trigger("click")
    await flushPromises()
    document.body.click()
    await flushPromises()
    expect(w.find("[role='menu']").exists()).toBe(false)
  })

  it("opens from the trigger with ArrowUp onto the last item", async () => {
    await w.find("button[aria-haspopup='menu']").trigger("keydown", { key: "ArrowUp" })
    await flushPromises()
    const items = menuItems()
    expect(document.activeElement).toBe(items[items.length - 1])
  })

  it("emits logout and closes", async () => {
    await w.find("button[aria-haspopup='menu']").trigger("click")
    await flushPromises()
    await w.find(".am-item--danger").trigger("click")
    expect(w.emitted("logout")).toHaveLength(1)
    expect(w.find("[role='menu']").exists()).toBe(false)
  })
})

describe("AppHeader", () => {
  let w: VueWrapper
  beforeEach(() => setActivePinia(createPinia()))
  afterEach(() => w.unmount())

  async function mountHeader() {
    const router = makeRouter()
    await router.push("/")
    w = mount(AppHeader, { global: { plugins: [i18n, router] }, attachTo: document.body })
  }

  it("opens the nav sheet from the burger and closes it once the viewport reaches lg", async () => {
    const original = window.innerWidth
    await mountHeader()
    const burger = w.get(".hdr-burger")
    expect(burger.attributes("aria-expanded")).toBe("false")
    await burger.trigger("click")
    await flushPromises()
    expect(burger.attributes("aria-expanded")).toBe("true")

    window.innerWidth = BREAKPOINTS.lg - 1
    window.dispatchEvent(new Event("resize"))
    await flushPromises()
    expect(burger.attributes("aria-expanded")).toBe("true")

    window.innerWidth = BREAKPOINTS.lg
    window.dispatchEvent(new Event("resize"))
    await flushPromises()
    expect(burger.attributes("aria-expanded")).toBe("false")
    expect(document.querySelector("[role='dialog']")).toBeNull()
    window.innerWidth = original
  })

  it("renders login and sign up for guests, without an account menu", async () => {
    await mountHeader()
    expect(w.find(".hdr-guest").exists()).toBe(true)
    expect(w.find(".hdr-guest a[href='/login']").exists()).toBe(true)
    expect(w.find(".hdr-guest a[href='/register']").exists()).toBe(true)
    expect(w.find("button[aria-haspopup='menu']").exists()).toBe(false)
    expect(w.find(".hdr-bell").exists()).toBe(false)
  })

  it("renders the account menu and a bell for signed-in users, with a capped badge", async () => {
    setActivePinia(createPinia())
    useAuthStore().user = user
    const noti = useNotificationStore()
    await mountHeader()
    expect(w.find(".hdr-guest").exists()).toBe(false)
    expect(w.find("button[aria-haspopup='menu']").exists()).toBe(true)
    expect(w.find(".hdr-badge").exists()).toBe(false)
    noti.unreadCount = 120
    await flushPromises()
    expect(w.find(".hdr-badge").text()).toBe("99+")
    noti.unreadCount = 4
    await flushPromises()
    expect(w.find(".hdr-badge").text()).toBe("4")
  })
})

describe("MobileNavSheet", () => {
  it("opens as a dialog, focuses close, locks scroll and restores focus on close", async () => {
    const router = makeRouter()
    await router.push("/")
    const opener = document.createElement("button")
    document.body.appendChild(opener)
    opener.focus()
    const w = mount(MobileNavSheet, {
      props: { modelValue: false, user: null, isAdmin: false },
      global: { plugins: [i18n, router] },
      attachTo: document.body,
    })
    expect(document.querySelector("[role='dialog']")).toBeNull()

    await w.setProps({ modelValue: true })
    await flushPromises()
    expect(document.querySelector("[role='dialog']")).not.toBeNull()
    expect(document.activeElement?.classList.contains("mns-close")).toBe(true)
    expect(document.body.style.overflow).toBe("hidden")
    expect(document.querySelector("a[href='/login']")).not.toBeNull()
    expect(document.querySelector("[aria-current='page']")?.getAttribute("href")).toBe("/")

    const panel = document.querySelector<HTMLElement>("[role='dialog']")!
    panel.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }))
    await flushPromises()
    expect(w.emitted("update:modelValue")?.[0]).toEqual([false])

    await w.setProps({ modelValue: false })
    await flushPromises()
    expect(document.body.style.overflow).toBe("")
    expect(document.activeElement).toBe(opener)
    w.unmount()
    opener.remove()
  })
})
