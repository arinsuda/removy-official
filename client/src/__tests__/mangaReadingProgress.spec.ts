import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createPinia, setActivePinia } from "pinia"
import { AxiosError, AxiosHeaders } from "axios"
import MangaReadingProgress from "@/components/manga/MangaReadingProgress.vue"
import { readingProgressApi } from "@/api/endpoints/readingProgress"
import { MANGA_PROGRESS_QUERY_KEY, MANGA_PROGRESS_SAVE_DEBOUNCE_MS } from "@/components/manga/mangaConfig"
import { parseProgressField, progressMax, stepProgress } from "@/components/manga/mangaProgress"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import type { MangaItem } from "@/types/manga"
import type { ReadingProgress } from "@/types/readingProgress"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/readingProgress", () => ({
  readingProgressApi: { get: vi.fn(), put: vi.fn(), remove: vi.fn() },
}))

const i18n = useRealI18n()
const api = vi.mocked(readingProgressApi)
const wrappers: VueWrapper[] = []

const MANGA_ID = 7

const record = (chapter: number, volume: number): ReadingProgress => ({
  media_id: MANGA_ID,
  media_type: "manga",
  chapter,
  volume,
  updated_at: "2026-11-01T00:00:00Z",
})

/** A 4xx the way the API sends it, so the error mapping reads the real `code`. */
const apiError = (code: string) =>
  new AxiosError("rejected", "ERR_BAD_REQUEST", undefined, undefined, {
    status: 400,
    statusText: "Bad Request",
    headers: {},
    config: { headers: new AxiosHeaders() },
    data: { code, error: "server text the client must not show" },
  })

/** Stub of useMediaActivity; the composable has its own specs. */
function activityStub(watched = false) {
  return {
    isWatched: { value: watched },
    pending: { value: new Set<string>() },
    toggleWatched: vi.fn(async () => {}),
  }
}

/** Flushes promises and the zero-delay timers TanStack Query uses to notify observers. */
async function settle() {
  await flushPromises()
  await vi.advanceTimersByTimeAsync(0)
  await flushPromises()
}

function render(
  options: {
    progress?: ReadingProgress | null
    chapters?: number | null
    volumes?: number | null
    watched?: boolean
    userId?: number
    loadError?: boolean
    queryClient?: QueryClient
  } = {},
) {
  const pinia = createPinia()
  setActivePinia(pinia)
  useAuthStore().user = { id: options.userId ?? 1 } as never
  if (options.loadError) api.get.mockRejectedValue(apiError("UPSTREAM_UNAVAILABLE"))
  else api.get.mockResolvedValue({ data: { progress: options.progress ?? null } } as never)
  const queryClient = options.queryClient ?? new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const activity = activityStub(options.watched)
  const manga = { id: MANGA_ID, chapters: options.chapters ?? null, volumes: options.volumes ?? null } as MangaItem
  const w = mount(MangaReadingProgress, {
    props: { manga, activity: activity as never },
    global: {
      plugins: [
        i18n,
        pinia,
        [VueQueryPlugin, { queryClient }],
      ],
    },
  })
  wrappers.push(w)
  return { w, activity, queryClient }
}

const stepperValue = (w: VueWrapper, index: number) => w.findAll(".manga-progress-stepper__value")[index].text()
const button = (w: VueWrapper, selector: string) => w.get(selector)
const toastMessages = () => useUIStore().toasts.map((toast) => toast.message)

describe("MangaReadingProgress", () => {
  beforeEach(() => {
    // Only timers the progress code uses are faked; promises keep flowing.
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] })
    vi.clearAllMocks()
    i18n.global.locale.value = "en"
    api.put.mockImplementation(async (_id, input) => ({ data: { progress: record(input.chapter, input.volume) } }) as never)
    api.remove.mockResolvedValue({ data: undefined } as never)
  })

  afterEach(() => {
    wrappers.splice(0).forEach((w) => w.unmount())
    vi.useRealTimers()
  })

  describe("limits", () => {
    it("caps each field at AniList's total when it is known, otherwise at the server cap", () => {
      expect(progressMax("chapter", { chapters: 120, volumes: null })).toBe(120)
      expect(progressMax("chapter", { chapters: null, volumes: null })).toBe(9999)
      expect(progressMax("volume", { chapters: null, volumes: 5000 })).toBe(999)
      expect(progressMax("volume", { chapters: null, volumes: 5 })).toBe(5)
    })

    it("stops steps at the limits and drops a step that would reach (0, 0)", () => {
      const totals = { chapters: 120, volumes: null }
      expect(stepProgress({ chapter: 120, volume: 1 }, "chapter", 1, totals)).toBeNull()
      expect(stepProgress({ chapter: 150, volume: 1 }, "chapter", 1, totals)).toBeNull()
    expect(stepProgress({ chapter: 3, volume: 2 }, "chapter", -5, totals)).toEqual({ chapter: 0, volume: 2 })
      expect(stepProgress({ chapter: 1, volume: 0 }, "chapter", -1, totals)).toBeNull()
    })

    it("accepts only whole numbers within the range", () => {
      expect(parseProgressField(" 7 ", 120)).toBe(7)
      expect(parseProgressField("121", 120)).toBeNull()
      expect(parseProgressField("1.5", 120)).toBeNull()
      expect(parseProgressField("", 120)).toBeNull()
    })

    it("disables the increase button at AniList's chapter total", async () => {
      const { w } = render({ progress: record(3, 0), chapters: 3 })
      await settle()
      expect(button(w, '[aria-label="Increase Chapter"]').attributes("disabled")).toBeDefined()
      expect(button(w, '[aria-label="Decrease Chapter"]').attributes("disabled")).toBeUndefined()
    })

    it("disables the decrease button when a step would save (0, 0)", async () => {
      const { w } = render({ progress: record(1, 0) })
      await settle()
      expect(button(w, '[aria-label="Decrease Chapter"]').attributes("disabled")).toBeDefined()
    })
  })

  describe("saving", () => {
    it("shows each step at once and saves once after the pause", async () => {
      const { w } = render({ progress: record(5, 0) })
      await settle()
      expect(stepperValue(w, 0)).toBe("5")

      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await settle()
      expect(stepperValue(w, 0)).toBe("7")
      expect(api.put).not.toHaveBeenCalled()

      await vi.advanceTimersByTimeAsync(MANGA_PROGRESS_SAVE_DEBOUNCE_MS)
      expect(api.put).toHaveBeenCalledTimes(1)
      expect(api.put).toHaveBeenCalledWith(MANGA_ID, { chapter: 7, volume: 0 })
    })

    it("rolls back to the saved position and toasts the mapped message when the save is rejected", async () => {
      api.put.mockRejectedValue(apiError("PROGRESS_OUT_OF_RANGE"))
      const { w } = render({ progress: record(5, 0) })
      await settle()

      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await settle()
      expect(stepperValue(w, 0)).toBe("6")

      await vi.advanceTimersByTimeAsync(MANGA_PROGRESS_SAVE_DEBOUNCE_MS)
      await settle()
      expect(stepperValue(w, 0)).toBe("5")
      expect(toastMessages()).toContain("That chapter or volume is above the allowed maximum.")
      expect(toastMessages().join(" ")).not.toContain("server text")
    })

    it("falls back to the generic message for a code it does not know", async () => {
      api.put.mockRejectedValue(apiError("SOMETHING_NEW"))
      const { w } = render({ progress: record(5, 0) })
      await settle()

      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await vi.advanceTimersByTimeAsync(MANGA_PROGRESS_SAVE_DEBOUNCE_MS)
      await settle()
      expect(toastMessages()).toContain("Could not save your reading progress. Please try again.")
    })

    it("clears the saved position with DELETE and shows the not-started state", async () => {
      const { w } = render({ progress: record(5, 2) })
      await settle()

      await button(w, '[data-test="manga-progress-clear"]').trigger("click")
      await settle()
      expect(api.remove).toHaveBeenCalledWith(MANGA_ID)
      expect(w.text()).toContain("Not started yet")
      expect(toastMessages()).toContain("Reading progress cleared")
    })
  })

  describe("rollback and pending saves", () => {
    it("rolls back to the server's value after an older response when a newer save then fails", async () => {
      let resolveFirst: (value: unknown) => void = () => {}
      api.put
        .mockImplementationOnce(() => new Promise((resolve) => (resolveFirst = resolve)) as never)
        .mockRejectedValueOnce(apiError("INVALID_INPUT"))
      const { w } = render({ progress: record(5, 0) })
      await settle()

      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await vi.advanceTimersByTimeAsync(MANGA_PROGRESS_SAVE_DEBOUNCE_MS)
      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await vi.advanceTimersByTimeAsync(MANGA_PROGRESS_SAVE_DEBOUNCE_MS)
      resolveFirst({ data: { progress: record(6, 0) } })
      await settle()
      expect(stepperValue(w, 0)).toBe("6")
    })

    it("does not drop a pending step when the edit dialog saves the same values", async () => {
      const { w } = render({ progress: record(5, 0) })
      await settle()
      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await button(w, '[data-test="manga-progress-edit"]').trigger("click")
      await settle()
      document.body
        .querySelector<HTMLFormElement>('[data-test="manga-progress-edit-form"]')
        ?.dispatchEvent(new Event("submit", { cancelable: true }))
      await settle()
      await vi.advanceTimersByTimeAsync(MANGA_PROGRESS_SAVE_DEBOUNCE_MS)
      expect(api.put).toHaveBeenCalledWith(MANGA_ID, { chapter: 6, volume: 0 })
    })
  })

  describe("viewer and load errors", () => {
    it("keys the cached progress by viewer so another account never sees it", async () => {
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
      render({ progress: record(5, 0), userId: 1, queryClient })
      await settle()
      expect(queryClient.getQueryData([MANGA_PROGRESS_QUERY_KEY, 1, "manga", MANGA_ID])).toEqual(record(5, 0))
      expect(queryClient.getQueryData([MANGA_PROGRESS_QUERY_KEY, 2, "manga", MANGA_ID])).toBeUndefined()

      const second = render({ progress: null, userId: 2, queryClient })
      await settle()
      expect(stepperValue(second.w, 0)).toBe("0")
    })

    it("shows the load failure message, not the save failure one", async () => {
      const { w } = render({ loadError: true })
      await settle()
      expect(w.text()).toContain("Could not load your reading progress. Please try again.")
      expect(w.text()).not.toContain("Could not save")
    })
  })

  describe("edit dialog", () => {
    it("saves both fields at once and refuses an out-of-range value", async () => {
      const { w } = render({ progress: record(5, 0) })
      await settle()

      await button(w, '[data-test="manga-progress-edit"]').trigger("click")
      await settle()
      const inputs = () => Array.from(document.body.querySelectorAll<HTMLInputElement>('input[type="number"]'))
      const form = () => document.body.querySelector<HTMLFormElement>('[data-test="manga-progress-edit-form"]')
      const saveButton = () => form()?.querySelector<HTMLButtonElement>('button[type="submit"]')

      inputs()[0].value = "9"
      inputs()[0].dispatchEvent(new Event("input"))
      inputs()[1].value = "2"
      inputs()[1].dispatchEvent(new Event("input"))
      await settle()
      expect(saveButton()?.disabled).toBe(false)
      form()?.dispatchEvent(new Event("submit", { cancelable: true }))
      await settle()
      expect(api.put).toHaveBeenCalledWith(MANGA_ID, { chapter: 9, volume: 2 })

      await button(w, '[data-test="manga-progress-edit"]').trigger("click")
      await settle()
      inputs()[1].value = "5000"
      inputs()[1].dispatchEvent(new Event("input"))
      await settle()
      expect(saveButton()?.disabled).toBe(true)
      expect(api.put).toHaveBeenCalledTimes(1)
    })

    it("accepts a value above AniList's total up to the server cap", async () => {
      const { w } = render({ progress: record(5, 0), chapters: 120 })
      await settle()
      await button(w, '[data-test="manga-progress-edit"]').trigger("click")
      await settle()
      const input = document.body.querySelector<HTMLInputElement>('input[type="number"]')!
      input.value = "150"
      input.dispatchEvent(new Event("input"))
      await settle()
      document.body
        .querySelector<HTMLFormElement>('[data-test="manga-progress-edit-form"]')
        ?.dispatchEvent(new Event("submit", { cancelable: true }))
      await settle()
      expect(api.put).toHaveBeenCalledWith(MANGA_ID, { chapter: 150, volume: 0 })
    })
  })

  describe("mark as read", () => {
    it("offers the prompt only at the final chapter and never marks the title on its own", async () => {
      const { w, activity } = render({ progress: record(2, 0), chapters: 3 })
      await settle()
      expect(w.find('[data-test="manga-progress-mark-read"]').exists()).toBe(false)

      await button(w, '[aria-label="Increase Chapter"]').trigger("click")
      await settle()
      expect(w.find('[data-test="manga-progress-mark-read"]').exists()).toBe(true)
      expect(activity.toggleWatched).not.toHaveBeenCalled()

      await button(w, '[data-test="manga-progress-mark-read"]').trigger("click")
      expect(activity.toggleWatched).toHaveBeenCalledTimes(1)
    })

    it("hides the prompt once the title is already marked as read", async () => {
      const { w } = render({ progress: record(3, 0), chapters: 3, watched: true })
      await settle()
      expect(w.find('[data-test="manga-progress-mark-read"]').exists()).toBe(false)
    })

    it("never offers it when AniList does not know the chapter count", async () => {
      const { w } = render({ progress: record(500, 0), chapters: null })
      await settle()
      expect(w.find('[data-test="manga-progress-mark-read"]').exists()).toBe(false)
      expect(w.find('[role="progressbar"]').exists()).toBe(false)
    })
  })
})
