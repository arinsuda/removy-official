/** Maps a failed handoff/link request to a navigation.link.handoff.* i18n key. */
export function handoffErrorKey(err: unknown): string {
  const e = err as { response?: { status?: number } } | undefined
  if (!e?.response) return "navigation.link.handoff.network"
  switch (e.response.status) {
    case 429:
      return "navigation.link.handoff.tooMany"
    case 502:
      return "navigation.link.handoff.emailFailed"
    case 503:
      return "navigation.link.handoff.disabled"
    default:
      return "navigation.link.handoff.wrong"
  }
}
