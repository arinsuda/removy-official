export interface HighlightPart {
  text: string
  match: boolean
}

const escapeRe = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")

/** Split text into parts, flagging case-insensitive matches of any query word. */
export function highlightParts(text: string, query: string): HighlightPart[] {
  const words = query.trim().split(/\s+/).filter(Boolean)
  if (!text || words.length === 0) return [{ text, match: false }]
  // The capture group makes split() keep the matches at odd indexes.
  const re = new RegExp(`(${words.map(escapeRe).join("|")})`, "gi")
  return text
    .split(re)
    .map((part, i) => ({ text: part, match: i % 2 === 1 }))
    .filter((p) => p.text !== "")
}
