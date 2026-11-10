type AdminErrorShape = { response?: { status?: number; data?: { error?: string; code?: string } } }

const CODE_TO_KEY: Record<string, string> = {
  FINAL_ADMIN: "admin.errors.finalAdmin",
  INACTIVE_USER_ROLE_CHANGE: "admin.errors.inactiveRole",
  SELF_DEACTIVATION: "admin.errors.selfDeactivate",
  SELF_DELETION: "admin.errors.selfDelete",
  SELF_ERASURE: "admin.errors.selfErase",
  USER_NOT_REMOVED: "admin.errors.notRemoved",
  CONFIRMATION_MISMATCH: "admin.errors.confirmMismatch",
  ALREADY_IN_STATUS: "admin.errors.alreadyStatus",
  INVALID_ROLE: "admin.errors.invalidRole",
  INVALID_INPUT: "admin.errors.invalidInput",
  USER_NOT_FOUND: "admin.errors.notFound",
  REVIEW_NOT_FOUND: "admin.errors.notFound",
  FORBIDDEN: "admin.errors.forbidden",
  UNAUTHORIZED: "admin.errors.forbidden",
  INTERNAL_ERROR: "admin.errors.generic",
}

/**
 * Maps an admin API failure to an `admin.errors.*` i18n key by machine-readable `code`,
 * falling back to HTTP status for uncoded 404/403 responses and `admin.errors.generic`.
 */
export function adminErrorKey(err: unknown): string {
  const res = (err as AdminErrorShape)?.response
  const code = res?.data?.code
  if (code && Object.hasOwn(CODE_TO_KEY, code)) return CODE_TO_KEY[code]!

  const status = res?.status
  if (status === 404) return "admin.errors.notFound"
  if (status === 403 || status === 401) return "admin.errors.forbidden"
  return "admin.errors.generic"
}
