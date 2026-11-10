import { vi } from "vitest"

/** Extra time on top of a component debounce for mocked requests to resolve and render. */
export const SPEC_WAIT_SLACK_MS = 2000

/** Wait options for an action gated by a debounce of `debounceMs`. */
export const waitAfterDebounce = (debounceMs: number) => ({ timeout: debounceMs + SPEC_WAIT_SLACK_MS })

/**
 * Resolve once `selector` matches an enabled element. Asserting on the outcome (instead of
 * advancing fake timers) avoids racing Vue's scheduler against the debounce timer.
 */
export function waitForEnabled<T extends HTMLElement = HTMLButtonElement>(selector: string, debounceMs = 0): Promise<T> {
  return vi.waitFor(() => {
    const el = document.body.querySelector<T>(selector)
    if (!el || (el as unknown as HTMLButtonElement).disabled) throw new Error(`${selector} not ready`)
    return el
  }, waitAfterDebounce(debounceMs))
}
