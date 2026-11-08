import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import FeedItem from "@/components/feed/FeedItem.vue"
import { i18n } from "@/i18n"
import type { FeedItemResponse } from "@/types/feed"

const actor = (avatar_url: string | null) => ({ id: 1, username: "ann", display_name: null, avatar_url })

const build = (avatar: string | null, targetAvatar: string | null = null): FeedItemResponse => ({
  id: 1,
  type: "user_followed",
  actor: actor(avatar),
  target_user: { ...actor(targetAvatar), id: 2, username: "bob" },
  message: "",
  visibility: "default",
  created_at: new Date().toISOString(),
})

const mountItem = (item: FeedItemResponse) =>
  mount(FeedItem, {
    props: { item },
    global: { plugins: [i18n], stubs: { RouterLink: { template: "<a><slot /></a>" } } },
  })

describe("FeedItem avatars", () => {
  it("renders a initials instead of a broken image when avatar is missing", () => {
    const wrapper = mountItem(build(null, null))
    expect(wrapper.findAll("[data-avatar-fallback]")).toHaveLength(2)
    expect(wrapper.findAll("img")).toHaveLength(0)
  })

  it("renders the image when avatar exists", () => {
    const wrapper = mountItem(build("https://x/a.png", "https://x/b.png"))
    expect(wrapper.findAll("img")).toHaveLength(2)
    expect(wrapper.findAll("[data-avatar-fallback]")).toHaveLength(0)
  })
})
