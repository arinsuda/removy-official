import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import { createMemoryHistory, createRouter } from "vue-router"
import NotFoundView from "@/views/NotFoundView.vue"
import en from "@/i18n/locales/en/common.json"
import { expectKeysInBothLocales, useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()

describe("NotFoundView", () => {
  it("has its keys in both locales", () => {
    expectKeysInBothLocales(["common.notFound.heading", "common.notFound.message", "common.notFound.home"])
  })

  it("renders a heading and a link back home", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: "/", component: { template: "<div />" } }],
    })
    i18n.global.locale.value = "en"
    await router.push("/")
    const wrapper = mount(NotFoundView, { global: { plugins: [i18n, router] } })
    expect(wrapper.get("h1").text()).toBe(en.notFound.heading)
    const link = wrapper.get("a")
    expect(link.attributes("href")).toBe("/")
    expect(link.text()).toBe(en.notFound.home)
  })
})
