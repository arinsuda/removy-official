/**
 * Input-capability queries. Hover UI (popups, hover reveals) is gated on a precise hovering
 * pointer, never on viewport width: an iPad is wide but has no hover.
 */
export const HOVER_FINE_QUERY = "(hover: hover) and (pointer: fine)"
export const COARSE_POINTER_QUERY = "(pointer: coarse)"

/** Whether the primary input can hover; false where matchMedia is missing (SSR, old test envs). */
export const canHover = (): boolean =>
  typeof window !== "undefined" && (window.matchMedia?.(HOVER_FINE_QUERY).matches ?? false)
