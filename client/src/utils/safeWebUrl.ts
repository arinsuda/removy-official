/** Returns the URL string only when it is a plain http(s) URL; otherwise "". */
export function safeWebUrl(raw: unknown): string {
  if (typeof raw !== "string") return ""
  try {
    const u = new URL(raw)
    return u.protocol === "http:" || u.protocol === "https:" ? raw : ""
  } catch {
    return ""
  }
}
