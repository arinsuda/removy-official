import { ref } from "vue"
import { useAuthStore } from "@/stores/auth"

export type AuthGateReason =
  | "rate"
  | "review"
  | "like"
  | "helpful"
  | "watchlist"
  | "watchLog"
  | "list"
  | "bmol"
  | "follow"
  | "comment"
  | "shareFeed"
  | "generic"

const isOpen = ref(false)
const reason = ref<AuthGateReason>("generic")

/**
 * Shared guest write gate.
 * Call `requireAuth(reason)` before any signed-in action reachable from public views;
 * guests see the single shared AuthPromptModal and receive `false`.
 */
export function useAuthGate() {
  function isSignedIn(): boolean {
    const auth = useAuthStore()
    return auth.isLoggedIn
  }

  function requireAuth(nextReason: AuthGateReason = "generic"): boolean {
    if (isSignedIn()) return true
    reason.value = nextReason
    isOpen.value = true
    return false
  }

  function close(): void {
    isOpen.value = false
  }

  return {
    isOpen,
    reason,
    requireAuth,
    close,
  }
}
