import { describe, expect, it } from "vitest"
import { mount } from "@vue/test-utils"
import { createPinia } from "pinia"
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
    global: { plugins: [createPinia(), i18n], stubs: { RouterLink: { template: "<a><slot /></a>" } } },
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

describe("FeedItem manga row", () => {
  const mangaItem = (media: NonNullable<FeedItemResponse["media"]>): FeedItemResponse => ({
    id: 2,
    type: "watched_added",
    actor: actor(null),
    media,
    message: "",
    visibility: "default",
    viewer_liked: false,
    created_at: new Date().toISOString(),
  })

  it("shows the start year and kind under the type label", () => {
    const wrapper = mountItem(
      mangaItem({ id: 7, media_type: "manga", title: "Solo", poster_url: "https://s4.anilist.co/x.png", start_year: 2018, kind: "manhwa" }),
    )
    expect(wrapper.text()).toContain("2018 · มันฮวา")
  })

  it("uses a book icon instead of the film placeholder when there is no poster", () => {
    const wrapper = mountItem(mangaItem({ id: 7, media_type: "manga", title: "Solo", poster_url: null, start_year: null, kind: "manga" }))
    const placeholder = wrapper.get(".feed-item__poster--empty").html()
    expect(placeholder).toContain("lucide-book-open")
    expect(placeholder).not.toContain("lucide-film")
  })

  it("keeps the film placeholder and skips the year line for movies", () => {
    const wrapper = mountItem(mangaItem({ id: 9, media_type: "movie", title: "Heat", poster_url: null }))
    expect(wrapper.get(".feed-item__poster--empty").html()).toContain("lucide-film")
    expect(wrapper.findAll(".feed-item__media-type")).toHaveLength(1)
  })
})
