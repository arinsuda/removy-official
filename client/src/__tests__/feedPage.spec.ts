/**
 * @vitest-environment happy-dom
 */
import { beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import FeedItem from "@/components/feed/FeedItem.vue"
import FeedList from "@/components/feed/FeedList.vue"
import FeedFilters from "@/components/feed/FeedFilters.vue"
import { i18n } from "@/i18n"
import type { FeedItemResponse } from "@/types/feed"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/api", () => ({ reviewApi: { likeReview: vi.fn(), unlikeReview: vi.fn() } }))
import { reviewApi } from "@/api/api"

useRealI18n()

const stub = { template: "<div />" }
const router = () =>
  createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/", component: stub },
      { name: "user-profile", path: "/users/:userId", component: stub },
      { name: "movie-detail", path: "/movies/:id", component: stub },
      { name: "tv-detail", path: "/tv/:id", component: stub },
      { name: "search-results", path: "/search", component: stub },
      { name: "movies", path: "/movies", component: stub },
      { name: "settings", path: "/settings", component: stub },
    ],
  })

const item = (over: Partial<FeedItemResponse> = {}): FeedItemResponse => ({
  id: 1,
  type: "review_created",
  actor: { id: 2, username: "mali", display_name: "Mali Film", avatar_url: null },
  media: { id: 603, media_type: "movie", title: "Inception", poster_url: null },
  review_id: 55,
  message: "A slow-burn masterpiece.",
  visibility: "default",
  created_at: new Date().toISOString(),
  ...over,
})

const mountItem = (it: FeedItemResponse, canManage = false) =>
  mount(FeedItem, { props: { item: it, canManage }, global: { plugins: [createPinia(), router(), i18n] } })

beforeEach(() => {
  i18n.global.locale.value = "en"
  setActivePinia(createPinia())
  vi.clearAllMocks()
})

describe("FeedItem", () => {
  it("renders a localized sentence with linked actor and title", () => {
    const w = mountItem(item())
    expect(w.find(".feed-item__sentence").text()).toBe("Mali Film reviewed Inception")
    expect(w.find('a[href="/movies/603"]').exists()).toBe(true)
    expect(w.find('a[href="/users/2"]').exists()).toBe(true)
    expect(w.find(".feed-item__excerpt").text()).toBe("A slow-burn masterpiece.")
  })

  it.each([
    ["watchlist_added", "Mali Film added Inception to their watchlist"],
    ["rewatched", "Mali Film rewatched Inception"],
    ["media_liked", "Mali Film liked Inception"],
  ] as const)("renders %s", (type, text) => {
    expect(mountItem(item({ type, message: "" })).find(".feed-item__sentence").text()).toBe(text)
  })

  it("renders achievements and follows", () => {
    expect(mountItem(item({ type: "achievement_unlocked", media: undefined, message: "Critic I" })).find(".feed-item__sentence").text()).toBe(
      "Mali Film unlocked the achievement Critic I",
    )
    const follow = mountItem(item({ type: "user_followed", media: undefined, message: "", target_user: { id: 9, username: "bob", display_name: null, avatar_url: null } }))
    expect(follow.find(".feed-item__sentence").text()).toBe("Mali Film started following bob")
  })

  it("likes and unlikes a review inline, rolling back when the call fails", async () => {
    vi.mocked(reviewApi.likeReview).mockResolvedValue({} as never)
    vi.mocked(reviewApi.unlikeReview).mockResolvedValue({} as never)
    const w = mountItem(item())
    const like = () => w.find(".feed-item__like")
    await like().trigger("click")
    await flushPromises()
    expect(reviewApi.likeReview).toHaveBeenCalledWith(55)
    expect(like().attributes("aria-pressed")).toBe("true")
    await like().trigger("click")
    await flushPromises()
    expect(reviewApi.unlikeReview).toHaveBeenCalledWith(55)

    vi.spyOn(console, "error").mockImplementation(() => {})
    vi.mocked(reviewApi.likeReview).mockRejectedValue({ response: { status: 500 } })
    await like().trigger("click")
    await flushPromises()
    expect(like().attributes("aria-pressed")).toBe("false")
  })

  it("hides the like button on your own activity and shows owner actions instead", async () => {
    const w = mountItem(item(), true)
    expect(w.find(".feed-item__like").exists()).toBe(false)
    await w.find(".feed-menu__trigger").trigger("click")
    const radios = w.findAll('[role="menuitemradio"]')
    expect(radios).toHaveLength(4)
    await radios[1]!.trigger("click")
    expect(w.emitted("update-visibility")?.[0]).toEqual([1, "public"])
  })

  it("asks for confirmation before deleting", async () => {
    const w = mountItem(item(), true)
    await w.find(".feed-menu__trigger").trigger("click")
    await w.find(".feed-menu__item--danger").trigger("click")
    expect(w.emitted("delete-activity")).toBeUndefined()
    const confirm = document.body.querySelector<HTMLButtonElement>(".app-btn--danger")
    expect(confirm).not.toBeNull()
    confirm!.click()
    await flushPromises()
    expect(w.emitted("delete-activity")?.[0]).toEqual([1])
    w.unmount()
  })
})

describe("FeedList", () => {
  const base = { items: [], loading: false, loadingMore: false, error: null, hasMore: false, isEmpty: false }
  const mountList = (props: Record<string, unknown>) => mount(FeedList, { props: { ...base, ...props }, global: { plugins: [createPinia(), router(), i18n] } })

  it("shows skeletons while loading", () => {
    expect(mountList({ loading: true }).find('[aria-busy="true"]').exists()).toBe(true)
  })

  it("shows an error with retry", async () => {
    const w = mountList({ error: "load_failed" })
    expect(w.text()).toContain("Could not load the feed")
    await w.find('[role="alert"] button').trigger("click")
    expect(w.emitted("retry")).toHaveLength(1)
  })

  it("shows follow suggestions when the feed is empty, a filter message when filtered", () => {
    const empty = mountList({ isEmpty: true })
    expect(empty.text()).toContain("Your feed is quiet")
    expect(empty.find('a[href="/search?q="]').exists()).toBe(true)
    const filtered = mountList({ isEmpty: true, filtered: true })
    expect(filtered.text()).toContain("No activity in this category yet")
    expect(filtered.find('a[href="/search?q="]').exists()).toBe(false)
  })

  it("renders items, marks only the viewer's own as manageable, and shows the end note", () => {
    const w = mountList({ items: [item({ id: 1 }), item({ id: 2, actor: { id: 7, username: "me", display_name: null, avatar_url: null } })], currentUserId: 7 })
    expect(w.findAll("article")).toHaveLength(2)
    expect(w.findAll(".feed-menu")).toHaveLength(1)
    expect(w.text()).toContain("You are all caught up")
  })
})

describe("FeedFilters", () => {
  it("lists the server categories and emits the chosen one", async () => {
    const w = mount(FeedFilters, { props: { modelValue: "" }, global: { plugins: [i18n] } })
    const tabs = w.findAll('[role="tab"]')
    expect(tabs.map((t) => t.text())).toEqual(["All", "Reviews", "Lists", "Social"])
    await tabs[1]!.trigger("click")
    expect(w.emitted("update:modelValue")?.[0]).toEqual(["reviews"])
  })
})
