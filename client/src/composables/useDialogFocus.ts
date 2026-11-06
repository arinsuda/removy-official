import { nextTick, onBeforeUnmount, watch, type Ref } from "vue"

const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

function focusablesIn(root: HTMLElement): HTMLElement[] {
  return Array.from(root.querySelectorAll<HTMLElement>(FOCUSABLE))
}

/**
 * Focus management for a dialog panel: moves focus in when `active` turns on,
 * returns it to the opener when it turns off, and wraps Tab inside the panel.
 * Bind `onKeydown` to the panel only; focus leaving the panel is not forced back.
 */
export function useDialogFocus(
  panel: Ref<HTMLElement | null>,
  active: Ref<boolean>,
  initialFocus?: Ref<string | HTMLElement | undefined>,
) {
  let opener: HTMLElement | null = null

  function resolveInitial(root: HTMLElement): HTMLElement | null {
    const target = initialFocus?.value
    if (typeof target === "string") return root.querySelector<HTMLElement>(target)
    return target ?? null
  }

  function restoreFocus() {
    if (opener?.isConnected) opener.focus()
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
      } else {
        restoreFocus()
      }
    },
    { immediate: true, flush: "post" },
  )

  onBeforeUnmount(() => {
    if (active.value) restoreFocus()
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
    if (event.shiftKey && (current === first || current === panel.value)) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && current === last) {
      event.preventDefault()
      first.focus()
    }
  }

  return { onKeydown }
}
