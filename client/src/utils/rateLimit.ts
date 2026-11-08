/** True when an axios error is the API's 429 / RATE_LIMITED response. */
export function isRateLimited(err: unknown): boolean {
  const res = (err as { response?: { status?: number; data?: { code?: string } } })?.response
  return res?.status === 429 || res?.data?.code === "RATE_LIMITED"
}
