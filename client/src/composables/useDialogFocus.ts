import { nextTick, onBeforeUnmount, watch, type Ref } from "vue"

const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

export function focusablesIn(root: HTMLElement): HTMLElement[] {
  return Array.from(root.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(
    (el) => el.tabIndex >= 0 && !el.matches(":disabled") && el.getClientRects().length > 0,
  )
}

/**
 * Focus management for a dialog panel: moves focus in when `active` turns on,
 * returns it to the opener when it turns off (only when `isTop` is true), and
 * wraps Tab inside the panel. Bind `onKeydown` to the panel (and call it for
 * document Tab when focus lands outside the panel).
 */
export function useDialogFocus(
  panel: Ref<HTMLElement | null>,
  active: Ref<boolean>,
  initialFocus?: Ref<string | HTMLElement | undefined>,
  isTop: () => boolean = () => true,
) {
  let opener: HTMLElement | null = null
  let observer: MutationObserver | null = null

  function resolveInitial(root: HTMLElement): HTMLElement | null {
    const target = initialFocus?.value
    if (typeof target === "string") return root.querySelector<HTMLElement>(target)
    return target ?? null
  }

  function stopObserver() {
    observer?.disconnect()
    observer = null
  }

  function restoreFocus() {
    stopObserver()
    if (isTop() && opener?.isConnected) opener.focus()
    opener = null
  }

  watch(
    active,
    async (isActive) => {
      if (isActive) {
        opener = document.activeElement instanceof HTMLElement ? document.activeElement : null
        await nextTick()
        const root = panel.value
        if (!root) return
        const target = resolveInitial(root) ?? focusablesIn(root)[0] ?? root
        target.focus()
        stopObserver()
        observer = new MutationObserver(() => {
          const r = panel.value
          if (!r || !active.value || !isTop()) return
          const cur = document.activeElement
          if (cur === document.body || (cur instanceof HTMLElement && r.contains(cur) && cur.matches(":disabled"))) {
            r.focus()
          }
        })
        observer.observe(root, { attributes: true, subtree: true, attributeFilter: ["disabled"] })
      } else {
        restoreFocus()
      }
    },
    { immediate: true, flush: "post" },
  )

  onBeforeUnmount(() => {
    if (active.value) restoreFocus()
    else stopObserver()
  })

  function onKeydown(event: KeyboardEvent) {
    if (event.key !== "Tab" || !panel.value) return
    const items = focusablesIn(panel.value)
    if (items.length === 0) {
      event.preventDefault()
      panel.value.focus()
      return
    }
    const first = items[0]!
    const last = items[items.length - 1]!
    const current = document.activeElement
    if (!panel.value.contains(current)) {
      event.preventDefault()
      ;(event.shiftKey ? last : first).focus()
      return
    }
    if (event.shiftKey && (current === first || current === panel.value)) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && (current === last || current === panel.value)) {
      event.preventDefault()
      first.focus()
    }
  }

  return { onKeydown }
}
