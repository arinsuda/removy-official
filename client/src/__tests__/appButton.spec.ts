import { describe, expect, it, vi } from "vitest"
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

  it.each(["ghost", "danger", "outline", "primary"] as const)("applies the %s variant class", (variant) => {
    const wrapper = mount(AppButton, { props: { variant } })
    expect(wrapper.classes()).toContain(`app-btn--${variant}`)
  })

  it("applies the small size class", () => {
    expect(mount(AppButton, { props: { size: "sm" } }).classes()).toContain("app-btn--sm")
  })

  it("warns in dev when an icon-only button has no accessible name", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {})
    mount(AppButton, { props: { iconOnly: true } })
    expect(warn).toHaveBeenCalledOnce()
    warn.mockClear()
    const wrapper = mount(AppButton, { props: { iconOnly: true }, attrs: { "aria-label": "Close" } })
    expect(warn).not.toHaveBeenCalled()
    expect(wrapper.classes()).toContain("app-btn--icon")
    warn.mockRestore()
  })

  it("marks a disabled link inert and blocks its click", async () => {
    const onClick = vi.fn()
    const wrapper = mount(AppButton, {
      props: { href: "https://example.com", disabled: true },
      attrs: { onClick },
    })
    expect(wrapper.attributes("aria-disabled")).toBe("true")
    expect(wrapper.attributes("tabindex")).toBe("-1")
    await wrapper.trigger("click")
    expect(onClick).not.toHaveBeenCalled()
  })

  it("lets an enabled link click through", async () => {
    const onClick = vi.fn()
    const wrapper = mount(AppButton, { props: { href: "https://example.com" }, attrs: { onClick } })
    expect(wrapper.attributes("aria-disabled")).toBeUndefined()
    expect(wrapper.attributes("tabindex")).toBeUndefined()
    await wrapper.trigger("click")
    expect(onClick).toHaveBeenCalledOnce()
  })
})
