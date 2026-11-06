import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import { createMemoryHistory, createRouter } from "vue-router"
import NotFoundView from "@/views/NotFoundView.vue"
import { createI18n } from "vue-i18n"
import en from "@/i18n/locales/en/common.json"

const i18n = createI18n({ legacy: false, locale: "en", messages: { en: { common: { notFound: en.notFound } } } })

describe("NotFoundView", () => {
  it("renders a heading and a link back home", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: "/", component: { template: "<div />" } }],
    })
    await router.push("/")
    const wrapper = mount(NotFoundView, { global: { plugins: [i18n, router] } })
    expect(wrapper.get("h1").text()).toBe(en.notFound.heading)
    const link = wrapper.get("a")
    expect(link.attributes("href")).toBe("/")
    expect(link.text()).toBe(en.notFound.home)
  })
})
