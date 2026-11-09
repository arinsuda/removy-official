import { LIST_ERROR_CODES } from "@/views/library/config"

const PREFIX = "library.customLists.errors"

export function httpStatus(err: unknown): number | undefined {
  return (err as { response?: { status?: number } } | null)?.response?.status
}

/** i18n key for a failed list request: the server's code when known, else a generic message. */
export function listErrorKey(err: unknown): string {
  const code = (err as { response?: { data?: { code?: string } } } | null)?.response?.data?.code
  const known = LIST_ERROR_CODES.find((c) => c === code)
  return `${PREFIX}.${known ?? "generic"}`
}
