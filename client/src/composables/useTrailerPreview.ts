import { ref, computed } from "vue"
import type { Video } from "@/types"

export type TrailerPriority = "trailer" | "teaser" | "clip" | "other"

export interface ResolvedTrailer {
  key: string
  name: string
  type: string
  priority: TrailerPriority
  embedUrl: string
}

export interface PopupPosition {
  top: number
  left: number
  width: number
  transformOrigin: "top left" | "top center" | "top right"
}

const TYPE_SCORE: Record<string, number> = {
  Trailer: 0,
  Teaser: 1,
  Clip: 2,
  Featurette: 3,
  "Behind the Scenes": 4,
  Bloopers: 5,
}

const PRIORITY_MAP: Record<string, TrailerPriority> = {
  Trailer: "trailer",
  Teaser: "teaser",
  Clip: "clip",
}

export function buildEmbedUrl(key: string): string {
  const p = new URLSearchParams({
    autoplay: "1",
    mute: "0",
    controls: "0",
    playsinline: "1",
    rel: "0",
    loop: "1",
    playlist: key,
    modestbranding: "1",
    iv_load_policy: "3",
    disablekb: "1",
    fs: "0",
    enablejsapi: "1",
  })
  return `https://www.youtube.com/embed/${key}?${p.toString()}`
}

/**
 * คืน "รายการ" trailer ที่เรียงลำดับความเหมาะสมไว้แล้ว (ดีที่สุดก่อน)
 * ใช้แทน resolveTrailer เดิมเมื่อจะรองรับ retry เมื่อ key แรกเล่นไม่ได้
 */
export function resolveTrailerCandidates(
  videos: Video[] | undefined | null,
): ResolvedTrailer[] {
  if (!videos?.length) return []

  const yt = videos.filter(v => v.site === "YouTube")
  if (!yt.length) return []

  const sorted = [...yt].sort((a, b) => {
    if (a.official !== b.official) return a.official ? -1 : 1
    const aS = TYPE_SCORE[a.type] ?? 99
    const bS = TYPE_SCORE[b.type] ?? 99
    return aS - bS
  })

  const seen = new Set<string>()
  const candidates: ResolvedTrailer[] = []
  for (const v of sorted) {
    if (seen.has(v.key)) continue
    seen.add(v.key)
    candidates.push({
      key: v.key,
      name: v.name,
      type: v.type,
      priority: PRIORITY_MAP[v.type] ?? "other",
      embedUrl: buildEmbedUrl(v.key),
    })
  }
  return candidates
}

/** @deprecated ใช้ resolveTrailerCandidates แทน เก็บไว้ให้ของเดิมที่ยังเรียกอยู่ไม่พัง */
export function resolveTrailer(
  videos: Video[] | undefined | null,
): ResolvedTrailer | null {
  return resolveTrailerCandidates(videos)[0] ?? null
}

declare global {
  interface Window {
    YT: any
    onYouTubeIframeAPIReady?: () => void
  }
}

let ytApiPromise: Promise<void> | null = null

function loadYoutubeApi(): Promise<void> {
  if (window.YT?.Player) return Promise.resolve()
  if (ytApiPromise) return ytApiPromise

  ytApiPromise = new Promise(resolve => {
    const prevCallback = window.onYouTubeIframeAPIReady
    window.onYouTubeIframeAPIReady = () => {
      prevCallback?.()
      resolve()
    }
    if (!document.querySelector('script[src*="youtube.com/iframe_api"]')) {
      const tag = document.createElement("script")
      tag.src = "https://www.youtube.com/iframe_api"
      document.head.appendChild(tag)
    }
  })
  return ytApiPromise
}

interface Options {
  /** How long after hover before iframe is injected into DOM (default: 500ms) */
  mountDelay?: number
}

export function useTrailerPreview(options: Options = {}) {
  const { mountDelay = 500 } = options

  const isIframeMounted = ref(false)
  const isIframeLoaded = ref(false)

  const trailerCandidates = ref<ResolvedTrailer[]>([])
  const candidateIndex = ref(0)
  const trailerUnavailable = ref(false)

  let timer: ReturnType<typeof setTimeout> | null = null
  let ytPlayer: any = null

  const currentTrailer = computed<ResolvedTrailer | null>(
    () => trailerCandidates.value[candidateIndex.value] ?? null,
  )

  /** ตั้ง candidate list ใหม่ (เรียกครั้งเดียวตอน fetch videos สำเร็จ) */
  function setCandidates(candidates: ResolvedTrailer[]) {
    trailerCandidates.value = candidates
    candidateIndex.value = 0
    trailerUnavailable.value = candidates.length === 0
  }

  function scheduleMount() {
    if (!currentTrailer.value) return
    cancelMount()
    timer = setTimeout(() => {
      isIframeMounted.value = true
    }, mountDelay)
  }

  function cancelMount() {
    if (timer !== null) {
      clearTimeout(timer)
      timer = null
    }
  }

  function destroyPlayer() {
    try {
      ytPlayer?.destroy?.()
    } catch {}
    ytPlayer = null
  }

  function unmount() {
    cancelMount()
    destroyPlayer()
    isIframeMounted.value = false
    isIframeLoaded.value = false
  }

  function onIframeLoad() {
    isIframeLoaded.value = true
  }

  /**
   * เรียกใน onMounted ของ element ที่จะฝัง player (div ที่มี :id="elementId")
   * แทนการใช้ <iframe :src="..."> ตรงๆ เพราะต้องดัก onError จริงจาก YouTube
   */
  function attachPlayer(elementId: string) {
    const trailer = currentTrailer.value
    if (!trailer) return

    loadYoutubeApi().then(() => {
      if (!isIframeMounted.value || currentTrailer.value?.key !== trailer.key)
        return

      ytPlayer = new window.YT.Player(elementId, {
        width: "100%",
        height: "100%",
        videoId: trailer.key,
        playerVars: {
          autoplay: 1,
          mute: 0,
          controls: 0,
          playsinline: 1,
          rel: 0,
          modestbranding: 1,
          disablekb: 1,
          fs: 0,
        },
        events: {
          onReady: (e: any) => {
            const iframe = e.target?.getIframe?.()
            if (iframe) {
              iframe.style.position = "absolute"
              iframe.style.inset = "-1px"
              iframe.style.width = "calc(100% + 2px)"
              iframe.style.height = "calc(100% + 2px)"
              iframe.style.border = "none"
            }
            onIframeLoad()
          },
          onError: (e: any) => {
            onPlayerError(e?.data)
          },
        },
      })
    })
  }

  /** ลอง candidate ตัวถัดไปเมื่อวิดีโอปัจจุบันเล่นไม่ได้จริง */
  function onPlayerError(_code?: number) {
    destroyPlayer()
    isIframeLoaded.value = false
    isIframeMounted.value = false
    candidateIndex.value++

    if (candidateIndex.value >= trailerCandidates.value.length) {
      trailerUnavailable.value = true
      return
    }

    isIframeMounted.value = true
  }

  const showSkeleton = computed(
    () => isIframeMounted.value && !isIframeLoaded.value,
  )
  const showFallback = computed(
    () => !isIframeMounted.value || trailerUnavailable.value,
  )

  return {
    isIframeMounted,
    isIframeLoaded,
    showSkeleton,
    showFallback,
    currentTrailer,
    trailerUnavailable,
    setCandidates,
    scheduleMount,
    cancelMount,
    unmount,
    onIframeLoad,
    attachPlayer,
    onPlayerError,
  }
}

const POPUP_WIDTH = 300
const POPUP_OFFSET_Y = -8

export function usePopupPosition() {
  const position = ref<PopupPosition | null>(null)

  function calculate(cardEl: HTMLElement): void {
    const rect = cardEl.getBoundingClientRect()
    const vw = window.innerWidth
    const scrollY = window.scrollY

    const cardCenterX = rect.left + rect.width / 2
    let left = cardCenterX - POPUP_WIDTH / 2

    type Side = "left" | "center" | "right"
    let side: Side = "center"

    if (left < 12) {
      left = 12
      side = "left"
    } else if (left + POPUP_WIDTH > vw - 12) {
      left = vw - POPUP_WIDTH - 12
      side = "right"
    }

    const originMap: Record<Side, PopupPosition["transformOrigin"]> = {
      left: "top left",
      center: "top center",
      right: "top right",
    }

    position.value = {
      top: rect.top + scrollY + POPUP_OFFSET_Y,
      left,
      width: POPUP_WIDTH,
      transformOrigin: originMap[side],
    }
  }

  function clear() {
    position.value = null
  }

  return { position, calculate, clear }
}
