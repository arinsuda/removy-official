import { readdirSync, readFileSync } from "node:fs"
import { join, relative } from "node:path"

const SRC = join(import.meta.dirname, "..", "..")
const LETTER = /[A-Za-z฀-๿]/
const MARK = "hardcode-ok"

// Mirrors src/constants/breakpoints.ts (breakpoints.spec.ts keeps that file and main.css in sync).
const BREAKPOINT_PX = [480, 640, 768, 1024, 1280]
const MAX_EPSILON = 0.02
const MEDIA_WIDTH = /\b(min|max)-width:\s*(\d+(?:\.\d+)?)px/g
const TAILWIND_ARBITRARY = /\b(min|max)-\[(\d+(?:\.\d+)?)px\]:/g
const PLAIN_VH = /(?<![\w.-])\d+(?:\.\d+)?vh\b/
const DYNAMIC_VH = /\d(?:dvh|svh)\b/

const LINE_RULES: { name: string; re: RegExp }[] = [
  { name: "color", re: /#[0-9a-fA-F]{3,8}\b|\b(?:rgba?|hsla?)\(/ },
  // Values 0-10 are local stacking inside one component; anything above must use the --z-* scale.
  { name: "z-index", re: /(?:z-index|zIndex)\s*:\s*(?:1[1-9]|[2-9]\d|\d{3,})\b|\bz-(?:1[1-9]|[2-9]\d|\d{3,}|\[)/ },
  { name: "locale", re: /["'`](?:th-TH|en-US)["'`]/ },
]

function files(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) {
      if (entry.name !== "__tests__") files(full, out)
    } else if (/\.(vue|ts)$/.test(entry.name) && !/\.(spec|test)\.ts$/.test(entry.name)) out.push(full)
  }
  return out
}

/** Lines that carry (or follow) an exception marker are dropped before scanning. */
function stripExempt(source: string): string {
  const lines = source.split("\n")
  return lines.map((line, i) => (line.includes(MARK) || lines[i - 1]?.includes(MARK) ? "" : line)).join("\n")
}

const THAI = /[\u0E00-\u0E7F]/
const STRING_LITERAL = /(["'`])((?:\\.|(?!\1)[^\\])*)\1/g
/** Calls whose first argument is shown to the user as-is. */
const UI_SINK = /\b(?:addToast|showToast|toast\.\w+|alert|confirm|prompt)\(\s*(["'`])((?:\\.|(?!\1)[^\\])*)\1/

function scriptOf(source: string, isVue: boolean): { code: string; offset: number } {
  if (!isVue) return { code: source, offset: 0 }
  const m = /<script\b[^>]*>([\s\S]*?)<\/script>/.exec(source)
  if (!m) return { code: "", offset: 0 }
  return { code: m[1]!, offset: source.slice(0, m.index).split("\n").length }
}

/** User-facing copy written into code: any Thai string literal, or a literal passed straight to a toast/dialog. */
function scanScript(source: string, isVue: boolean): string[] {
  const { code, offset } = scriptOf(source, isVue)
  const found: string[] = []
  code.split("\n").forEach((raw, i) => {
    const line = raw.replace(/\/\/.*$/, "")
    const trimmed = line.trim()
    if (trimmed.startsWith("*") || trimmed.startsWith("/*")) return
    const sink = UI_SINK.exec(line)
    if (sink && LETTER.test(sink[2]!)) found.push(`script-text @${i + offset}: ${trimmed}`)
    else if ([...line.matchAll(STRING_LITERAL)].some((m) => THAI.test(m[2]!))) found.push(`script-text @${i + offset}: ${trimmed}`)
  })
  return found
}

function templateOf(source: string): string {
  const start = source.indexOf("<template>")
  const end = source.lastIndexOf("</template>")
  return start === -1 || end === -1 ? "" : source.slice(start + "<template>".length, end)
}

/** Responsive-foundation rules: breakpoint tokens only, dynamic viewport units instead of plain `vh`. */
export function scanLayout(source: string): string[] {
  const lines = stripExempt(source).split("\n")
  const found: string[] = []
  lines.forEach((line, i) => {
    if (line.includes("@media")) {
      for (const m of line.matchAll(MEDIA_WIDTH)) {
        const px = Number(m[2])
        const ok = BREAKPOINT_PX.some((bp) => (m[1] === "min" ? px === bp : Math.abs(px - (bp - MAX_EPSILON)) < 1e-6))
        if (!ok) found.push(`breakpoint @${i + 1}: ${line.trim()}`)
      }
    }
    for (const m of line.matchAll(TAILWIND_ARBITRARY)) {
      if (!BREAKPOINT_PX.includes(Number(m[2]))) found.push(`breakpoint @${i + 1}: ${m[0]}`)
    }
    // A plain `vh` is allowed only as the fallback written right before a dvh/svh declaration.
    if (PLAIN_VH.test(line) && !DYNAMIC_VH.test(line) && !DYNAMIC_VH.test(lines[i + 1] ?? "")) {
      found.push(`viewport-unit @${i + 1}: ${line.trim()}`)
    }
  })
  return found
}

export function scanSource(source: string, isVue: boolean): string[] {
  const text = stripExempt(source)
  const found: string[] = scanLayout(source)
  text.split("\n").forEach((line, i) => {
    for (const rule of LINE_RULES) if (rule.re.test(line)) found.push(`${rule.name} @${i + 1}: ${line.trim()}`)
  })
  found.push(...scanScript(text, isVue))
  if (!isVue) return found

  const tpl = templateOf(text).replace(/<!--[\s\S]*?-->/g, "")
  for (const m of tpl.matchAll(/(?<![:\w@#-])(aria-label|title|alt|placeholder)="([^"]*)"/g)) {
    if (LETTER.test(m[2]!.replace(/\{\{[\s\S]*?\}\}/g, ""))) found.push(`attr ${m[1]}: ${m[2]}`)
  }
  // Quoted attribute values may contain ">" (arrows, comparisons); blank them before reading text nodes.
  const bare = tpl.replace(/"[^"]*"/g, '""').replace(/\{\{[\s\S]*?\}\}/g, "")
  for (const m of bare.matchAll(/>([^<>]+)</g)) {
    if (LETTER.test(m[1]!)) found.push(`text: ${m[1]!.trim()}`)
  }
  return found
}

/** Every vue/ts/css file under src (stylesheets included), for the layout rules only. */
export function scanLayoutTree(): string[] {
  const all = (dir: string, out: string[] = []): string[] => {
    for (const e of readdirSync(dir, { withFileTypes: true })) {
      const full = join(dir, e.name)
      if (e.isDirectory()) {
        if (e.name !== "__tests__") all(full, out)
      } else if (/\.(vue|ts|css)$/.test(e.name)) out.push(full)
    }
    return out
  }
  return all(SRC).flatMap((f) => scanLayout(readFileSync(f, "utf8")).map((v) => `${relative(SRC, f)} ${v}`))
}

export function scanDir(dir: string): string[] {
  return files(join(SRC, dir)).flatMap((f) => scanSource(readFileSync(f, "utf8"), f.endsWith(".vue")).map((v) => `${relative(SRC, f)} ${v}`))
}

