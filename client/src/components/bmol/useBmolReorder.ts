import { computed, inject, provide, reactive, ref, type ComputedRef, type InjectionKey } from "vue"
import { useI18n } from "vue-i18n"
import { moveTarget, type RankGroup } from "./bmolRanks"

const DRAG_THRESHOLD = 6

export interface BmolReorder {
  grabbedId: number | null
  targetRank: number | null
  /** Pointer offset of the dragged card, for the lifted look. */
  dx: number
  dy: number
  dragging: boolean
  onHandleKeydown(e: KeyboardEvent, id: number, rank: number): void
  onHandlePointerDown(e: PointerEvent, id: number, rank: number): void
  onHandlePointerMove(e: PointerEvent): void
  onHandlePointerUp(e: PointerEvent): void
  cancel(): void
}

const KEY: InjectionKey<BmolReorder> = Symbol("bmol-reorder")

/** Owned by the board; cards read it through useBmolReorderState. */
export function provideBmolReorder(
  groups: ComputedRef<RankGroup[]>,
  onDrop: (id: number, rank: number) => void,
  titleOf: (id: number) => string,
) {
  const { t } = useI18n()
  const announcement = ref("")
  let origin = { x: 0, y: 0, rank: 0, mode: "" as "" | "pointer" | "keyboard", pending: false }

  const state = reactive({ grabbedId: null as number | null, targetRank: null as number | null, dx: 0, dy: 0, dragging: false })
  const rankLabel = (rank: number) => t("library.bmol.drag.rank", { rank })

  function reset() {
    state.grabbedId = null
    state.targetRank = null
    state.dx = 0
    state.dy = 0
    state.dragging = false
    origin = { x: 0, y: 0, rank: 0, mode: "", pending: false }
  }

  function commit() {
    const id = state.grabbedId
    const target = state.targetRank
    const from = origin.rank
    reset()
    if (id === null || target === null || target === from) {
      announcement.value = t("library.bmol.drag.cancelled")
      return
    }
    announcement.value = t("library.bmol.drag.dropped", { title: titleOf(id), rank: rankLabel(target) })
    onDrop(id, target)
  }

  function cancel() {
    if (state.grabbedId === null) return
    reset()
    announcement.value = t("library.bmol.drag.cancelled")
  }

  function onHandleKeydown(e: KeyboardEvent, id: number, rank: number) {
    if (state.grabbedId === null) {
      if (e.key !== " " && e.key !== "Enter") return
      e.preventDefault()
      Object.assign(origin, { rank, mode: "keyboard" })
      state.grabbedId = id
      state.targetRank = rank
      announcement.value = t("library.bmol.drag.grabbed", { title: titleOf(id), rank: rankLabel(rank) })
      return
    }
    if (state.grabbedId !== id) return
    if (e.key === "ArrowUp" || e.key === "ArrowDown") {
      e.preventDefault()
      const next = moveTarget(groups.value, state.targetRank ?? rank, e.key === "ArrowUp" ? "up" : "down")
      if (next === null) return
      state.targetRank = next
      announcement.value = t("library.bmol.drag.moved", { title: titleOf(id), rank: rankLabel(next) })
    } else if (e.key === "Enter" || e.key === " ") {
      e.preventDefault()
      commit()
    } else if (e.key === "Escape") {
      e.preventDefault()
      e.stopPropagation()
      cancel()
    }
  }

  function onHandlePointerDown(e: PointerEvent, id: number, rank: number) {
    if (e.button !== 0 || state.grabbedId !== null) return
    ;(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
    origin = { x: e.clientX, y: e.clientY, rank, mode: "pointer", pending: true }
    state.grabbedId = id
    state.targetRank = rank
  }

  function onHandlePointerMove(e: PointerEvent) {
    if (origin.mode !== "pointer" || state.grabbedId === null) return
    const dx = e.clientX - origin.x
    const dy = e.clientY - origin.y
    if (!state.dragging && Math.hypot(dx, dy) < DRAG_THRESHOLD) return
    if (!state.dragging) announcement.value = t("library.bmol.drag.grabbed", { title: titleOf(state.grabbedId), rank: rankLabel(origin.rank) })
    state.dragging = true
    state.dx = dx
    state.dy = dy
    const zone = document.elementFromPoint(e.clientX, e.clientY)?.closest<HTMLElement>("[data-bmol-rank]")
    const rank = zone ? Number(zone.dataset.bmolRank) : NaN
    state.targetRank = Number.isFinite(rank) ? rank : origin.rank
  }

  function onHandlePointerUp(e: PointerEvent) {
    if (origin.mode !== "pointer") return
    ;(e.currentTarget as HTMLElement).releasePointerCapture?.(e.pointerId)
    if (state.dragging) commit()
    else reset()
  }

  // Plain object with getters: reads stay reactive because they go through `state`.
  const api: BmolReorder = {
    get grabbedId() { return state.grabbedId },
    get targetRank() { return state.targetRank },
    get dx() { return state.dx },
    get dy() { return state.dy },
    get dragging() { return state.dragging },
    onHandleKeydown,
    onHandlePointerDown,
    onHandlePointerMove,
    onHandlePointerUp,
    cancel,
  }
  provide(KEY, api)
  return { reorder: api, announcement, cancel, active: computed(() => state.grabbedId !== null) }
}

export function useBmolReorderState(): BmolReorder | null {
  return inject(KEY, null)
}
