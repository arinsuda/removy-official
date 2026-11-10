/**
 * Mobile-first breakpoints in px. Single source for JS; the CSS side is `--breakpoint-*`
 * in `assets/styles/main.css` and the two must match (see breakpoints.spec.ts).
 */
export const BREAKPOINTS = { xs: 480, sm: 640, md: 768, lg: 1024, xl: 1280 } as const

export type BreakpointName = keyof typeof BREAKPOINTS

/** Distance below a breakpoint for `max-width` queries, so they never overlap `min-width`. */
const MAX_WIDTH_EPSILON = 0.02

export const minWidthQuery = (bp: BreakpointName): string => `(min-width: ${BREAKPOINTS[bp]}px)`
export const maxWidthQuery = (bp: BreakpointName): string => `(max-width: ${BREAKPOINTS[bp] - MAX_WIDTH_EPSILON}px)`

export const isAtLeast = (bp: BreakpointName): boolean => window.innerWidth >= BREAKPOINTS[bp]
