import { reactive } from "vue"
import { LIST_DRAG_THRESHOLD } from "./config"

/** New order after moving the element at `from` to index `to`. */
export function moveIndex<T>(arr: T[], from: number, to: number): T[] {
  if (from === to || from < 0 || to < 0 || from >= arr.length || to >= arr.length) return arr
  const next = [...arr]
  next.splice(to, 0, next.splice(from, 1)[0]!)
  return next
}

/**
 * Pointer drag for a grid of cells carrying data-media-key (adapted from useBmolReorder):
 * press a handle, move past a small threshold, release over another cell. A plain click
 * on the handle does nothing; the up/down buttons cover keyboards.
 */
export function useListReorder(keys: () => string[], onMove: (from: number, to: number) => void) {
  const state = reactive({ key: null as string | null, target: null as string | null, dragging: false })
  let origin = { x: 0, y: 0 }

  const reset = () => Object.assign(state, { key: null, target: null, dragging: false })

  function onPointerDown(e: PointerEvent, key: string) {
    if (e.button !== 0 || state.key !== null) return
    ;(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
    origin = { x: e.clientX, y: e.clientY }
    Object.assign(state, { key, target: key, dragging: false })
  }

  function onPointerMove(e: PointerEvent) {
    if (state.key === null) return
    if (!state.dragging && Math.hypot(e.clientX - origin.x, e.clientY - origin.y) < LIST_DRAG_THRESHOLD) return
    state.dragging = true
    const cell = document.elementFromPoint(e.clientX, e.clientY)?.closest<HTMLElement>("[data-media-key]")
    state.target = cell?.dataset.mediaKey ?? state.key
  }

  function onPointerUp(e: PointerEvent) {
    if (state.key === null) return
    ;(e.currentTarget as HTMLElement).releasePointerCapture?.(e.pointerId)
    const { key, target, dragging } = state
    reset()
    if (!dragging || target === null || target === key) return
    const order = keys()
    onMove(order.indexOf(key), order.indexOf(target))
  }

  const cellState = (key: string) => (state.dragging && state.key === key ? "dragging" : state.dragging && state.target === key ? "target" : undefined)

  return { onPointerDown, onPointerMove, onPointerUp, cancel: reset, cellState, dragging: () => state.dragging }
}
