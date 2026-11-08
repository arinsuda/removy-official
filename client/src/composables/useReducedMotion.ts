import { onBeforeUnmount, ref } from "vue"

/** Tracks `prefers-reduced-motion: reduce`; false where matchMedia is missing. */
export function useReducedMotion() {
  const query = typeof window !== "undefined" ? window.matchMedia?.("(prefers-reduced-motion: reduce)") : undefined
  const reduced = ref(query?.matches ?? false)
  const onChange = (e: MediaQueryListEvent) => (reduced.value = e.matches)
  query?.addEventListener?.("change", onChange)
  onBeforeUnmount(() => query?.removeEventListener?.("change", onChange))
  return reduced
}
