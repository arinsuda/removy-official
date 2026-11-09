/** Logs only the HTTP status of a failed request; response bodies and URLs can carry user data. */
export function logApiError(label: string, err: unknown): void {
  const status = (err as { response?: { status?: number } } | null)?.response?.status
  console.error(label, status ?? "network")
}
