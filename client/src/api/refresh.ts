const LOCK_NAME = "popcorn-auth-refresh";

// Only a rejected session may log the user out. A 5xx (the server could not check the
// session) or a network error says nothing about the refresh token, so it must not.
export function isSessionRejected(error: unknown): boolean {
  const status = (error as { response?: { status?: number } })?.response?.status;
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
  const run = async () => {
    const current = localStorage.getItem("access_token");
    if (current && current !== staleAccessToken) return "already-refreshed" as const;
    await refresh();
    return "refreshed" as const;
  };
  if (typeof navigator !== "undefined" && navigator.locks) {
    return navigator.locks.request(LOCK_NAME, run);
  }
  return run();
}
