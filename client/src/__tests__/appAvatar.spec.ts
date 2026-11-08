import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import AppAvatar from "@/components/ui/AppAvatar.vue"

describe("AppAvatar", () => {
  it("renders the image when src is set", () => {
    const w = mount(AppAvatar, { props: { src: "https://x/a.png", name: "Ann Lee" } })
    expect(w.find("img").exists()).toBe(true)
    expect(w.attributes("aria-label")).toBe("Ann Lee")
    expect(w.attributes("data-avatar-fallback")).toBeUndefined()
  })

  it("shows initials when src is missing", () => {
    const w = mount(AppAvatar, { props: { name: "Ann Lee" } })
    expect(w.find("img").exists()).toBe(false)
    expect(w.text()).toBe("AL")
    expect(w.attributes("data-avatar-fallback")).toBeDefined()
  })

  it("falls back to initials when the image fails to load", async () => {
    const w = mount(AppAvatar, { props: { src: "https://x/bad.png", name: "bob" } })
    await w.find("img").trigger("error")
    expect(w.find("img").exists()).toBe(false)
    expect(w.text()).toBe("BO")
  })

  it("is hidden from assistive tech in decorative mode", () => {
    const w = mount(AppAvatar, { props: { name: "Ann", decorative: true } })
    expect(w.attributes("aria-hidden")).toBe("true")
    expect(w.attributes("role")).toBeUndefined()
    expect(w.attributes("aria-label")).toBeUndefined()
  })
})
