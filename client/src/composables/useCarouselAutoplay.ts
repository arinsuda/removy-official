import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue"

const AUTOPLAY_MS = 6000

/**
 * Index state plus autoplay for a slideshow. Autoplay stops while hovered,
 * focused, user-paused, in a hidden tab, or when the user prefers reduced motion.
 */
export function useCarouselAutoplay(length: () => number) {
  const index = ref(0)
  const hovering = ref(false)
  const focused = ref(false)
  const userPaused = ref(false)
  const tabHidden = ref(false)
  const reduceMotion = ref(false)

  const playing = computed(
    () => length() > 1 && !reduceMotion.value && !userPaused.value && !hovering.value && !focused.value && !tabHidden.value,
  )

  let timer: ReturnType<typeof setInterval> | null = null
  function stop() {
    if (timer) clearInterval(timer)
    timer = null
  }
  function schedule() {
    stop()
    if (!playing.value) return
    timer = setInterval(() => (index.value = (index.value + 1) % length()), AUTOPLAY_MS)
  }
  // Restarting on every index change gives a manual pick a full interval.
  watch([playing, index, length], schedule)
  watch(length, () => (index.value = 0))

  let motionQuery: MediaQueryList | null = null
  const syncMotion = () => (reduceMotion.value = motionQuery?.matches ?? false)
  const syncVisibility = () => (tabHidden.value = document.hidden)

  onMounted(() => {
    motionQuery = window.matchMedia("(prefers-reduced-motion: reduce)")
    syncMotion()
    syncVisibility()
    motionQuery.addEventListener("change", syncMotion)
    document.addEventListener("visibilitychange", syncVisibility)
    schedule()
  })
  onBeforeUnmount(() => {
    stop()
    motionQuery?.removeEventListener("change", syncMotion)
    document.removeEventListener("visibilitychange", syncVisibility)
  })

  return { index, userPaused, hovering, focused, playing }
}
