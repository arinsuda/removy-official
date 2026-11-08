import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import { BellOff } from "lucide-vue-next"
import AppCard from "@/components/ui/AppCard.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import { i18n } from "@/i18n"

describe("AppCard", () => {
  it("renders header and footer only when slots are given", () => {
    const bare = mount(AppCard, { slots: { default: "body" } })
    expect(bare.find("header").exists()).toBe(false)
    expect(bare.find("footer").exists()).toBe(false)
    const full = mount(AppCard, { slots: { default: "body", header: "H", footer: "F" } })
    expect(full.get("header").text()).toBe("H")
    expect(full.get("footer").text()).toBe("F")
  })
})

describe("AppEmptyState", () => {
  it("renders title, description, icon and action slot", () => {
    const w = mount(AppEmptyState, {
      props: { title: "Nothing", description: "Add some", icon: BellOff },
      slots: { action: "<button>Go</button>" },
    })
    expect(w.text()).toContain("Nothing")
    expect(w.text()).toContain("Add some")
    expect(w.find("svg").exists()).toBe(true)
    expect(w.find("button").text()).toBe("Go")
  })

  it("omits optional parts", () => {
    const w = mount(AppEmptyState, { props: { title: "Nothing" } })
    expect(w.find("svg").exists()).toBe(false)
    expect(w.find("button").exists()).toBe(false)
  })
})

describe("AppErrorState", () => {
  it("emits retry and uses the locale default label", async () => {
    const w = mount(AppErrorState, { props: { title: "Oops" }, global: { plugins: [i18n] } })
    expect(w.attributes("role")).toBe("alert")
    await w.find("button").trigger("click")
    expect(w.emitted("retry")).toHaveLength(1)
    expect(w.find("button").text().length).toBeGreaterThan(0)
  })

  it("accepts a custom retry label", () => {
    const w = mount(AppErrorState, { props: { title: "Oops", retryLabel: "Again" }, global: { plugins: [i18n] } })
    expect(w.find("button").text()).toBe("Again")
  })
})
