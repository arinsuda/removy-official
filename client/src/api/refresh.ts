const LOCK_NAME = "popcorn-auth-refresh";

// A refresh POST that hangs must not hold the lock (and every other tab) forever, and a
// tab waiting for the lock gives up after LOCK_WAIT_MS. Both failures carry no HTTP
// status, so isSessionRejected() is false and the user stays signed in.
export const REFRESH_POST_TIMEOUT_MS = 15000;
export const LOCK_WAIT_MS = 20000;

// Only a rejected session may log the user out. A 5xx (the server could not check the
// session) or a network error says nothing about the refresh token, so it must not.
// A 403 XHR_HEADER_REQUIRED is a missing-header bug, not a dead session.
export const XHR_HEADER_REQUIRED_CODE = "XHR_HEADER_REQUIRED";

export function isSessionRejected(error: unknown): boolean {
  const response = (error as { response?: { status?: number; data?: { code?: string } } })?.response;
  if (response?.data?.code === XHR_HEADER_REQUIRED_CODE) return false;
  const status = response?.status;
  return status === 400 || status === 401 || status === 403;
}

// Runs refresh() once across tabs. Refresh tokens rotate and work once, so a second tab
// that refreshes with the token the first tab just spent would log the user out. After
// acquiring the lock we look again: if the access token is no longer the one the failed
// request used, another tab already refreshed and there is nothing left to do.
export async function runExclusiveRefresh(
  staleAccessToken: string | null,
  refresh: () => Promise<void>,
): Promise<"refreshed" | "already-refreshed"> {
  return withRefreshLock(async () => {
    const current = localStorage.getItem("access_token");
    if (current && current !== staleAccessToken) return "already-refreshed" as const;
    await refresh();
    return "refreshed" as const;
  });
}

// Runs fn under the cross-tab refresh lock (or directly when the Locks API is missing).
// Read the refresh token inside fn: a tab that waited must use the token the winner stored.
export function withRefreshLock<T>(fn: () => Promise<T>): Promise<T> {
  if (typeof navigator !== "undefined" && navigator.locks) {
    // AbortSignal.timeout is missing in older browsers (Safari < 16): wait without a limit there.
    const options: LockOptions =
      typeof AbortSignal.timeout === "function" ? { signal: AbortSignal.timeout(LOCK_WAIT_MS) } : {};
    return navigator.locks.request(LOCK_NAME, options, fn);
  }
  return fn();
}
