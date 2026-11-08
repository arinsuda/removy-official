import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import AppInput from "@/components/ui/AppInput.vue"

const base = { label: "Email", modelValue: "" }

describe("AppInput", () => {
  it("links the label and forwards attributes to the input", () => {
    const wrapper = mount(AppInput, { props: base, attrs: { autocomplete: "email" } })
    const input = wrapper.get("input")
    expect(wrapper.get("label").attributes("for")).toBe(input.attributes("id"))
    expect(input.attributes("autocomplete")).toBe("email")
    expect(input.attributes("data-focus-custom")).toBeUndefined()
  })

  it("describes the input by its hint", () => {
    const wrapper = mount(AppInput, { props: { ...base, hint: "We never share it" } })
    const input = wrapper.get("input")
    expect(wrapper.get(`#${input.attributes("aria-describedby")}`).text()).toBe("We never share it")
    expect(input.attributes("aria-invalid")).toBeUndefined()
  })

  it("prefers the error over the hint and marks the input invalid", () => {
    const wrapper = mount(AppInput, { props: { ...base, hint: "Hint", error: "Required" } })
    const input = wrapper.get("input")
    expect(input.attributes("aria-invalid")).toBe("true")
    expect(wrapper.get(`#${input.attributes("aria-describedby")}`).text()).toBe("Required")
    expect(wrapper.text()).not.toContain("Hint")
  })

  it("supports disabled and v-model", async () => {
    const wrapper = mount(AppInput, { props: { ...base, disabled: true } })
    expect(wrapper.get("input").attributes("disabled")).toBeDefined()
    await wrapper.setProps({ disabled: false })
    await wrapper.get("input").setValue("a@b.co")
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["a@b.co"])
  })
})
