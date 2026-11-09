import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import { Lock } from "lucide-vue-next"
import AuthField from "@/components/auth/AuthField.vue"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()
const global = { plugins: [i18n] }

const base = { label: "Password", icon: Lock, modelValue: "" }

describe("AuthField", () => {
  it("links the label and forwards attributes to the input", () => {
    const wrapper = mount(AuthField, { global, props: base, attrs: { autocomplete: "username" } })
    const input = wrapper.get("input")
    expect(wrapper.get("label").attributes("for")).toBe(input.attributes("id"))
    expect(input.attributes("autocomplete")).toBe("username")
  })

  it("reveals a password only through the toggle", async () => {
    const wrapper = mount(AuthField, { global, props: { ...base, type: "password" } })
    expect(wrapper.get("input").attributes("type")).toBe("password")
    await wrapper.get("button").trigger("click")
    expect(wrapper.get("input").attributes("type")).toBe("text")
  })

  it("exposes the error to assistive technology", () => {
    const wrapper = mount(AuthField, { global, props: { ...base, error: "Required" } })
    const input = wrapper.get("input")
    expect(input.attributes("aria-invalid")).toBe("true")
    expect(wrapper.get(`#${input.attributes("aria-describedby")}`).text()).toBe("Required")
  })
})
