import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { nextTick, type Component } from "vue"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { useAuthStore } from "@/stores/auth"
import UserLikes from "@/components/profile/UserLikes.vue"
import UserWatched from "@/components/profile/UserWatched.vue"
import UserWatchlist from "@/components/profile/UserWatchlist.vue"
import { useRealI18n } from "./helpers/realI18n"

const { libraryItems } = vi.hoisted(() => ({
  libraryItems: [1, 2, 3].map((id) => ({
    id: id * 10,
    created_at: "2024-01-01",
    media: {
      id,
      title: `Title ${id}`,
      media_type: "movie",
      poster_url: "",
      genres: [],
      vote_average: 7,
    },
    watch_count: 0,
  })),
}))

vi.mock("@/api/api", () => ({
  libraryApi: {
    getVisibleUserLibrary: vi.fn().mockResolvedValue({ data: { items: libraryItems } }),
    removeItem: vi.fn().mockResolvedValue({}),
  },
}))
vi.mock("@/api/endpoints/media", () => ({
  mediaApi: {
    getLikeByUserId: vi.fn().mockResolvedValue({
      data: { likes: libraryItems.map((i) => ({ created_at: i.created_at, media: i.media })) },
    }),
    unlikeMedia: vi.fn().mockResolvedValue({}),
  },
}))
vi.mock("@/api/endpoints/watchLog", () => ({
  watchLogApi: { getMyWatchLogs: vi.fn().mockResolvedValue({ data: {} }) },
}))
vi.mock("vue-router", () => ({ useRouter: () => ({ push: vi.fn() }) }))

const i18n = useRealI18n()
const origGetClientRects = HTMLElement.prototype.getClientRects
const wrappers: VueWrapper[] = []

beforeEach(() => {
  HTMLElement.prototype.getClientRects = function (this: HTMLElement) {
    return (this.isConnected ? [{ width: 1, height: 1 }] : []) as unknown as DOMRectList
  }
})
afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
  HTMLElement.prototype.getClientRects = origGetClientRects
})

const cases: [string, Component, string][] = [
  ["UserLikes", UserLikes, "likes"],
  ["UserWatched", UserWatched, "watched"],
  ["UserWatchlist", UserWatchlist, "watchlist"],
]

describe.each(cases)("%s remove confirm", (_name, comp, listType) => {
  function mountAs(viewerId: number) {
    const pinia = createPinia()
    setActivePinia(pinia)
    useAuthStore().user = { id: viewerId } as never
    return mount(comp, {
      props: { userId: 1, listType },
      global: { plugins: [i18n, pinia] },
      attachTo: document.body,
    })
  }

  it("hides remove buttons from visitors", async () => {
    const wrapper = mountAs(2)
    wrappers.push(wrapper)
    await flushPromises()
    expect(wrapper.findAll(".rm-btn").length).toBe(0)
  })

  it("opens exactly one dialog for one remove click", async () => {
    const wrapper = mountAs(1)
    wrappers.push(wrapper)
    await flushPromises()

    const removeButtons = wrapper.findAll(".rm-btn")
    expect(removeButtons.length).toBe(3)
    await removeButtons[1].trigger("click")
    await nextTick()
    await nextTick()

    expect(document.body.querySelectorAll('[role="alertdialog"]').length).toBe(1)
    expect(document.body.textContent).toContain("Title 2")
  })
})
