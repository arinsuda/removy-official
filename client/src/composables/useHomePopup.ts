import { inject, onBeforeUnmount, onMounted, provide, ref, type InjectionKey, type Ref } from "vue"
import { movieApi } from "@/api/api"
import type { Video } from "@/types"
import type { HomeMedia } from "@/components/home/homeMedia"
import { resolveTrailerCandidates, useTrailerPreview } from "@/composables/useTrailerPreview"

type TrailerState = ReturnType<typeof useTrailerPreview>

export interface PopupAnchor {
  top: number
  left: number
}

export interface HomePopup {
  /** Key of the card whose popup is open, or null. */
  openKey: Ref<string | null>
  anchor: Ref<PopupAnchor>
  isDesktop: Ref<boolean>
  getState(item: HomeMedia): TrailerState
  /** Hover enter. */
  show(item: HomeMedia, key: string, el: HTMLElement): void
  /** Keyboard focus; ignored unless the browser reports :focus-visible. */
  focus(item: HomeMedia, key: string, el: HTMLElement): void
  /** Pointer or focus left the card. */
  leave(key: string): void
  popupEnter(): void
  popupLeave(key: string): void
  close(): void
}

const POPUP_KEY: InjectionKey<HomePopup> = Symbol("home-popup")
const SHOW_DELAY = 220
const LEAVE_DELAY = 80
const POPUP_WIDTH = 320
const EDGE = 8

/** Call once in the home view; cards read it with `useHomePopup()`. */
export function provideHomePopup(): HomePopup {
  const openKey = ref<string | null>(null)
  const anchor = ref<PopupAnchor>({ top: 0, left: 0 })
  const isDesktop = ref(true)

  const states = new Map<string, TrailerState>()
  const videoCache = new Map<string, Video[]>()
  let hoveredKey: string | null = null
  let insidePopup = false
  let showTimer: ReturnType<typeof setTimeout> | null = null
  let leaveTimer: ReturnType<typeof setTimeout> | null = null
  let media: MediaQueryList | null = null

  const syncViewport = () => {
    isDesktop.value = media?.matches ?? true
    if (!isDesktop.value) close()
  }

  function getState(item: HomeMedia): TrailerState {
    let state = states.get(item.key)
    if (!state) {
      state = useTrailerPreview({ mountDelay: 500 })
      states.set(item.key, state)
    }
    return state
  }

  async function loadTrailer(item: HomeMedia) {
    if (videoCache.has(item.key)) return
    videoCache.set(item.key, [])
    const state = getState(item)
    try {
      const res = item.type === "tv" ? await movieApi.getTVVideos(item.id) : await movieApi.getVideos(item.id)
      const videos = res.data?.results ?? []
      videoCache.set(item.key, videos)
      const candidates = resolveTrailerCandidates(videos)
      state.setCandidates(candidates)
      if (hoveredKey === item.key && candidates.length > 0) state.scheduleMount()
    } catch {
      state.setCandidates([])
    }
  }

  function place(el: HTMLElement) {
    const rect = el.getBoundingClientRect()
    const maxLeft = Math.max(EDGE, window.innerWidth - POPUP_WIDTH - EDGE)
    const left = rect.left + rect.width / 2 - POPUP_WIDTH / 2
    anchor.value = {
      top: Math.max(EDGE, rect.top - 8),
      left: Math.min(Math.max(EDGE, left), maxLeft),
    }
  }

  function clearTimers() {
    if (showTimer) clearTimeout(showTimer)
    if (leaveTimer) clearTimeout(leaveTimer)
    showTimer = leaveTimer = null
  }

  function show(item: HomeMedia, key: string, el: HTMLElement) {
    if (!isDesktop.value) return
    clearTimers()
    insidePopup = false
    hoveredKey = item.key
    void loadTrailer(item)
    showTimer = setTimeout(() => {
      place(el)
      openKey.value = key
      if (getState(item).currentTrailer.value) getState(item).scheduleMount()
    }, SHOW_DELAY)
  }

  function focus(item: HomeMedia, key: string, el: HTMLElement) {
    if (el.querySelector(":focus-visible")) show(item, key, el)
  }

  function closeIf(key: string) {
    if (openKey.value === key) close()
  }

  function leave(key: string) {
    if (!isDesktop.value) return
    if (showTimer) clearTimeout(showTimer)
    leaveTimer = setTimeout(() => {
      if (!insidePopup) closeIf(key)
      if (!openKey.value) hoveredKey = null
    }, LEAVE_DELAY)
  }

  function close() {
    clearTimers()
    insidePopup = false
    hoveredKey = null
    const key = openKey.value
    openKey.value = null
    if (key) {
      // Popup keys are `${rail}:${media key}`; trailer state is per media key.
      states.get(key.slice(key.indexOf(":") + 1))?.unmount()
    }
  }

  const onKeydown = (e: KeyboardEvent) => {
    if (e.key === "Escape" && openKey.value) close()
  }
  const onScroll = () => {
    if (openKey.value) close()
  }

  onMounted(() => {
    media = window.matchMedia("(min-width: 768px)")
    syncViewport()
    media.addEventListener("change", syncViewport)
    document.addEventListener("keydown", onKeydown)
    window.addEventListener("scroll", onScroll, { passive: true })
  })

  onBeforeUnmount(() => {
    media?.removeEventListener("change", syncViewport)
    document.removeEventListener("keydown", onKeydown)
    window.removeEventListener("scroll", onScroll)
    clearTimers()
    states.forEach((s) => s.unmount())
    states.clear()
  })

  const api: HomePopup = {
    openKey,
    anchor,
    isDesktop,
    getState,
    show,
    focus,
    leave,
    popupEnter: () => {
      insidePopup = true
      if (leaveTimer) clearTimeout(leaveTimer)
    },
    popupLeave: (key) => {
      insidePopup = false
      closeIf(key)
    },
    close,
  }
  provide(POPUP_KEY, api)
  return api
}

export function useHomePopup(): HomePopup {
  const popup = inject(POPUP_KEY)
  if (!popup) throw new Error("useHomePopup requires provideHomePopup in an ancestor")
  return popup
}
