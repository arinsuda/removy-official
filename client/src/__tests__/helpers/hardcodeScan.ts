import { readdirSync, readFileSync } from "node:fs"
import { join, relative } from "node:path"

const SRC = join(import.meta.dirname, "..", "..")
const LETTER = /[A-Za-z฀-๿]/
const MARK = "hardcode-ok"

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

export function scanSource(source: string, isVue: boolean): string[] {
  const text = stripExempt(source)
  const found: string[] = []
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

export function scanDir(dir: string): string[] {
  return files(join(SRC, dir)).flatMap((f) => scanSource(readFileSync(f, "utf8"), f.endsWith(".vue")).map((v) => `${relative(SRC, f)} ${v}`))
}

