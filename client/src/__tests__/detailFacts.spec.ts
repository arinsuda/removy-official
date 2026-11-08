import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import DetailFacts from "@/components/detail/DetailFacts.vue"
import { i18n } from "@/i18n"

const mountFacts = (homepage?: string) => mount(DetailFacts, { props: { facts: [], homepage }, global: { plugins: [i18n] } })

describe("DetailFacts homepage", () => {
  it("renders http(s) links with the bare host", () => {
    const a = mountFacts("https://www.example.com/film").get("a")
    expect(a.attributes("href")).toBe("https://www.example.com/film")
    expect(a.text()).toBe("example.com")
  })

  it.each(["javascript:alert(1)", "data:text/html,x", "not a url", ""])("omits %j", hp => {
    expect(mountFacts(hp).find("a").exists()).toBe(false)
  })
})
