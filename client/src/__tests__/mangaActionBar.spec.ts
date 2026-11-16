import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import MangaActionBar from "@/components/manga/MangaActionBar.vue"
import { useRealI18n } from "./helpers/realI18n"

const i18n = useRealI18n()
const wrappers: VueWrapper[] = []
beforeEach(() => {
  i18n.global.locale.value = "en"
})
afterEach(() => wrappers.splice(0).forEach((w) => w.unmount()))

/** Stub of useMediaActivity's return value; the composable itself has its own specs. */
function activityStub(overrides: Partial<{ signedIn: boolean; watchlisted: boolean; watched: boolean; liked: boolean }> = {}) {
  const signedIn = overrides.signedIn ?? true
  return {
    isWatchlisted: { value: overrides.watchlisted ?? false },
    isWatched: { value: overrides.watched ?? false },
    isLiked: { value: overrides.liked ?? false },
    pending: { value: new Set<string>() },
    requireLogin: vi.fn(() => signedIn),
    toggleWatchlist: vi.fn(async () => {}),
    toggleWatched: vi.fn(async () => {}),
    toggleLike: vi.fn(async () => {}),
  }
}

function render(activity: ReturnType<typeof activityStub>, hasOwnReview = false, userId: number | null = 1) {
  setActivePinia(createPinia())
  const w = mount(MangaActionBar, {
    props: { manga: { id: 7 } as never, userId, hasOwnReview, activity: activity as never },
    global: { plugins: [i18n] },
  })
  wrappers.push(w)
  return w
}

const button = (w: VueWrapper, test: string) => w.get(`[data-test="${test}"]`)

describe("MangaActionBar", () => {
  it("labels the library actions as Want to read and Read", () => {
    const w = render(activityStub())
    expect(button(w, "manga-action-want").text()).toBe("Want to read")
    expect(button(w, "manga-action-read").text()).toBe("Read")
  })

  it("shows Edit rating once the viewer has a review", () => {
    expect(button(render(activityStub()), "manga-action-rate").text()).toBe("Rate")
    expect(button(render(activityStub(), true), "manga-action-rate").text()).toBe("Edit rating")
  })

  it("exposes pressed state for the toggles", () => {
    const w = render(activityStub({ watchlisted: true, watched: true, liked: true }))
    expect(button(w, "manga-action-want").attributes("aria-pressed")).toBe("true")
    expect(button(w, "manga-action-read").attributes("aria-pressed")).toBe("true")
    expect(button(w, "manga-action-like").attributes("aria-pressed")).toBe("true")
    expect(button(w, "manga-action-like").text()).toBe("Remove like")
  })

  it("toggles the library and like state through the activity", async () => {
    const activity = activityStub()
    const w = render(activity)
    await button(w, "manga-action-want").trigger("click")
    await button(w, "manga-action-read").trigger("click")
    await button(w, "manga-action-like").trigger("click")
    expect(activity.toggleWatchlist).toHaveBeenCalledTimes(1)
    expect(activity.toggleWatched).toHaveBeenCalledTimes(1)
    expect(activity.toggleLike).toHaveBeenCalledTimes(1)
  })

  it("emits rate for a signed-in viewer", async () => {
    const w = render(activityStub())
    await button(w, "manga-action-rate").trigger("click")
    expect(w.emitted("rate")).toHaveLength(1)
  })

  it("asks a guest to sign in and does not open the BMOL sheet", async () => {
    const activity = activityStub({ signedIn: false })
    const w = render(activity, false, null)
    await button(w, "manga-action-rate").trigger("click")
    await button(w, "manga-action-bmol").trigger("click")
    expect(activity.requireLogin).toHaveBeenCalledWith("rate")
    expect(activity.requireLogin).toHaveBeenCalledWith("bmol")
    expect(w.emitted("rate")).toBeUndefined()
    expect(w.findComponent({ name: "MangaBmolSheet" }).exists()).toBe(false)
  })
})
