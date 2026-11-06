// Desktop-style slide between PLANZI and REMOVY. The apps are on different
// origins, so the source app plays the exit and the target app plays the enter.
// REMOVY is the right desktop: it leaves toward the right and arrives from the right.
// Keep in sync with cooked-app/frontend/src/utils/platformTransition.ts.
const SIDE = 1

const EXIT_MS = 480
const ENTER_MS = 680
const EASING = "cubic-bezier(0.32, 0.72, 0, 1)"
const BACKDROP = "#050505"
// Must match the inline arrival script in index.html.
const ARRIVING_CLASS = "platform-arriving"

function reducedMotion(): boolean {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches
}

function setStage(active: boolean) {
  const { style } = document.documentElement
  style.background = active ? BACKDROP : ""
  style.overflowX = active ? "hidden" : ""
}

/** Slides the current page away. Resolve, then navigate to the other platform. */
export async function playPlatformExit(): Promise<void> {
  if (reducedMotion()) return
  setStage(true)
  const animation = document.body.animate(
    [
      { transform: "translate3d(0, 0, 0)", filter: "blur(0)", opacity: 1 },
      { transform: `translate3d(${SIDE * 100}vw, 0, 0)`, filter: "blur(20px)", opacity: 0 },
    ],
    { duration: EXIT_MS, easing: EASING, fill: "forwards" },
  )
  // Restore the page if the browser brings it back from the back/forward cache.
  window.addEventListener(
    "pageshow",
    (event) => {
      if (!event.persisted) return
      animation.cancel()
      setStage(false)
    },
    { once: true },
  )
  await animation.finished
}

let entered = false

/** Slides the page in after arriving from the other platform. Runs at most once. */
export async function playPlatformEnter(): Promise<void> {
  if (entered) return
  entered = true
  const root = document.documentElement
  if (reducedMotion()) {
    root.classList.remove(ARRIVING_CLASS)
    return
  }
  setStage(true)
  const animation = document.body.animate(
    [
      { transform: `translate3d(${SIDE * 100}vw, 0, 0)`, filter: "blur(20px)", opacity: 0 },
      { transform: "translate3d(0, 0, 0)", filter: "blur(0)", opacity: 1 },
    ],
    { duration: ENTER_MS, easing: EASING },
  )
  root.classList.remove(ARRIVING_CLASS)
  try {
    await animation.finished
  } finally {
    setStage(false)
  }
}
