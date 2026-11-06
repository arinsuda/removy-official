import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import AppButton from "@/components/ui/AppButton.vue"

describe("AppButton", () => {
  it("renders a native button that defaults to type=button", () => {
    const wrapper = mount(AppButton, { slots: { default: "Go" } })
    expect(wrapper.element.tagName).toBe("BUTTON")
    expect(wrapper.attributes("type")).toBe("button")
    expect(wrapper.text()).toBe("Go")
  })

  it("is disabled and busy while loading, and hides its label", () => {
    const wrapper = mount(AppButton, {
      props: { loading: true },
      slots: { default: "Go" },
    })
    expect(wrapper.attributes("disabled")).toBeDefined()
    expect(wrapper.attributes("aria-busy")).toBe("true")
    expect(wrapper.text()).toBe("")
  })

  it("renders an anchor without button-only attributes when given href", () => {
    const wrapper = mount(AppButton, { props: { href: "https://example.com" } })
    expect(wrapper.element.tagName).toBe("A")
    expect(wrapper.attributes("type")).toBeUndefined()
    expect(wrapper.attributes("disabled")).toBeUndefined()
  })
})
