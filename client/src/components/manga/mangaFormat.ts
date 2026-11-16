/** Turns AniList's HTML-flavoured synopsis into plain text. Only line breaks survive. */
export function plainDescription(raw: string): string {
  return raw
    .replace(/<br\s*\/?>/gi, "\n")
    .replace(/<[^>]*>/g, "")
    .replace(/\n{3,}/g, "\n\n")
    .trim()
}

/** The number itself, or null when the API has no value (null or 0 means unknown). */
export function knownNumber(value: number | null): number | null {
  return value ? value : null
}

/** Status keys the UI translates. The API status is free text, so anything else is "other". */
export function statusKey(status: string): "releasing" | "finished" | "other" {
  const s = status.toLowerCase()
  return s === "releasing" || s === "finished" ? s : "other"
}

/** Only http(s) links are rendered, so a stored scheme like javascript: never reaches an href. */
export function isWebUrl(url: string): boolean {
  return /^https?:\/\//i.test(url)
}
