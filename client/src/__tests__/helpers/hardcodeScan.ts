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

