import { afterEach, beforeEach, describe, expect, it, vi, type Mock } from "vitest"
import { mount } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import ShareImageModal from "@/components/share/ShareImageModal.vue"
import { writeAdultAck } from "@/components/manga/mangaAdult"
import { i18n } from "@/i18n"
import { useAuthStore } from "@/stores/auth"
import type { ShareMediaContext } from "@/types/share"

const COVER = "https://s4.anilist.co/file/anilistcdn/media/manga/cover/medium/bx1.png"
const USER_ID = 1
const ADULT_DOB = "1990-01-01"
const MINOR_DOB = "2015-01-01"

const adultManga: ShareMediaContext = {
  id: 7,
  mediaType: "manga",
  title: "Adult title",
  posterPath: null,
  posterUrl: COVER,
  releaseYear: "2015",
  genres: [],
  genreNames: ["Drama"],
  voteAverage: 0,
  popcornRating: null,
  isAdult: true,
}

function stubImageFetch(): Mock {
  const fetchMock = vi.fn().mockResolvedValue(new Response(new Blob([new Uint8Array([1, 2])], { type: "image/jpeg" })))
  vi.stubGlobal("fetch", fetchMock)
  return fetchMock
}

function mountModal(media: ShareMediaContext) {
  return mount(ShareImageModal, {
    props: { modelValue: true, shareType: "poster", media },
    global: { plugins: [i18n] },
    attachTo: document.body,
  })
}

const fetchedUrls = (fetchMock: Mock) => fetchMock.mock.calls.map((c) => String(c[0]))

describe("adult manga covers in share images", () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ""
  })

  function signIn(dateOfBirth: string) {
    useAuthStore().user = { id: USER_ID, date_of_birth: dateOfBirth } as never
  }

  it("does not fetch an adult cover before the viewer confirms 18+", async () => {
    signIn(ADULT_DOB)
    const fetchMock = stubImageFetch()
    const w = mountModal(adultManga)
    await vi.waitFor(() => expect(document.body.querySelector("[data-template=poster]")).not.toBeNull())
    expect(fetchMock).not.toHaveBeenCalled()
    w.unmount()
  })

  it("fetches the cover through the proxy once an eligible viewer has confirmed", async () => {
    signIn(ADULT_DOB)
    writeAdultAck(USER_ID)
    const fetchMock = stubImageFetch()
    const w = mountModal(adultManga)
    await vi.waitFor(() => expect(fetchMock).toHaveBeenCalled())
    expect(fetchedUrls(fetchMock)).toEqual([`/api/share/proxy-image?url=${encodeURIComponent(COVER)}`])
    w.unmount()
  })

  it("never fetches an adult cover for a viewer under 18, even with a stored confirmation", async () => {
    signIn(MINOR_DOB)
    writeAdultAck(USER_ID)
    const fetchMock = stubImageFetch()
    const w = mountModal(adultManga)
    await vi.waitFor(() => expect(document.body.querySelector("[data-template=poster]")).not.toBeNull())
    expect(fetchMock).not.toHaveBeenCalled()
    w.unmount()
  })
})
