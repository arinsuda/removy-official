export interface ApiErrorLike {
  response?: { status?: number; data?: { code?: string; error?: string } }
}

/** Narrow a caught value to the axios-style error shape the API returns. */
export function asApiError(err: unknown): ApiErrorLike {
  return (err ?? {}) as ApiErrorLike
}
