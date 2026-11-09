/** Maps a failed DELETE /users/:id reply to a `settings.danger.errors.*` i18n key. */
export function deleteErrorKey(status: number | undefined, code: string | undefined): string {
  if (code === "WRONG_PASSWORD") return "settings.danger.errors.wrongPassword"
  if (code === "CONFIRM_MISMATCH") return "settings.danger.errors.confirmMismatch"
  if (status === 429) return "settings.danger.errors.tooMany"
  return "settings.danger.errors.generic"
}
