type AdminErrorShape = { response?: { status?: number; data?: { error?: string; code?: string } } }

/**
 * Maps an admin API failure to an `admin.errors.*` i18n key. The admin API answers with
 * plain `{ "error": "..." }` strings, so known protections are matched by status plus
 * message, or by an explicit `code` if the server adds one.
 */
export function adminErrorKey(err: unknown): string {
  const res = (err as AdminErrorShape)?.response
  const status = res?.status
  const code = res?.data?.code
  const msg = (res?.data?.error ?? "").toLowerCase()

  if (code === "FINAL_ADMIN" || msg.includes("final active admin")) return "admin.errors.finalAdmin"
  if (msg.includes("inactive user")) return "admin.errors.inactiveRole"
  if (msg.includes("deactivate your own")) return "admin.errors.selfDeactivate"
  if (msg.includes("delete your own")) return "admin.errors.selfDelete"
  if (msg.includes("already in requested status")) return "admin.errors.alreadyStatus"
  if (msg.includes("invalid role")) return "admin.errors.invalidRole"
  if (status === 404) return "admin.errors.notFound"
  if (status === 403) return "admin.errors.forbidden"
  return "admin.errors.generic"
}
