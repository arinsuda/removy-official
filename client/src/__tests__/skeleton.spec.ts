import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import SkeletonTableRows from "@/components/ui/skeleton/SkeletonTableRows.vue"

const global = { mocks: { $t: (key: string) => key } }

describe("skeleton components", () => {
  it("announces loading once and hides the placeholder blocks", () => {
    const wrapper = mount(SkeletonList, { props: { count: 3 }, global })
    expect(wrapper.findAll('[role="status"]')).toHaveLength(1)
    expect(wrapper.find(".skeleton-block").attributes("aria-hidden")).toBe("true")
    expect(wrapper.findAll(".skeleton-list__row")).toHaveLength(3)
  })

  it("renders one cell per column so table layout stays stable", () => {
    const wrapper = mount(
      { components: { SkeletonTableRows }, template: '<table><tbody><SkeletonTableRows :columns="8" :rows="2" /></tbody></table>' },
    )
    expect(wrapper.findAll("tr")).toHaveLength(2)
    expect(wrapper.findAll("td")).toHaveLength(16)
  })
})
