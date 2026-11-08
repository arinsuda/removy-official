import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import AppTabs from "@/components/ui/AppTabs.vue"

const tabs = [
  { key: "a", label: "A" },
  { key: "b", label: "B" },
  { key: "c", label: "C" },
]

const build = (modelValue = "a") =>
  mount(AppTabs, {
    props: {
      tabs,
      modelValue,
      label: "Demo",
      "onUpdate:modelValue": (v: string) => wrapper.setProps({ modelValue: v }),
    },
    attachTo: document.body,
  })
let wrapper: ReturnType<typeof build>

describe("AppTabs", () => {
  it("exposes tab roles, selection and roving tabindex", () => {
    wrapper = build("b")
    const btns = wrapper.findAll('[role="tab"]')
    expect(wrapper.find('[role="tablist"]').attributes("aria-label")).toBe("Demo")
    expect(btns.map((b) => b.attributes("aria-selected"))).toEqual(["false", "true", "false"])
    expect(btns.map((b) => b.attributes("tabindex"))).toEqual(["-1", "0", "-1"])
    expect(btns[1]!.attributes("aria-controls")).toContain("-panel-b")
    wrapper.unmount()
  })

  it("moves with arrow, Home and End keys, wrapping around", async () => {
    wrapper = build("a")
    const list = wrapper.find('[role="tablist"]')
    await list.trigger("keydown", { key: "ArrowLeft" })
    expect(wrapper.props("modelValue")).toBe("c")
    await list.trigger("keydown", { key: "ArrowRight" })
    expect(wrapper.props("modelValue")).toBe("a")
    await list.trigger("keydown", { key: "End" })
    expect(wrapper.props("modelValue")).toBe("c")
    await list.trigger("keydown", { key: "Home" })
    expect(wrapper.props("modelValue")).toBe("a")
    wrapper.unmount()
  })

  it("selects on click and can drop aria-controls", async () => {
    wrapper = build("a")
    await wrapper.setProps({ controls: false })
    const btns = wrapper.findAll('[role="tab"]')
    expect(btns[0]!.attributes("aria-controls")).toBeUndefined()
    await btns[2]!.trigger("click")
    expect(wrapper.props("modelValue")).toBe("c")
    wrapper.unmount()
  })
})
