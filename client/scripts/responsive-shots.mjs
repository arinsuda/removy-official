// Responsive screenshot + horizontal-overflow check. Needs a running stack (Vite + API).
// Usage: npm run shots [-- --montage]
// Env: RESPONSIVE_BASE_URL, RESPONSIVE_USER, RESPONSIVE_PASSWORD (authed routes are skipped when unset),
//      RESPONSIVE_CHROME_PATH (optional executable path; default is the system Chrome channel),
//      RESPONSIVE_OUT_DIR (default .shots/latest).
import { chromium } from "playwright-core"
import { mkdirSync, readdirSync, writeFileSync } from "node:fs"
import path from "node:path"
import { fileURLToPath } from "node:url"

const BASE_URL = (process.env.RESPONSIVE_BASE_URL || "http://localhost:3000").replace(/\/$/, "")
const USER = process.env.RESPONSIVE_USER
const PASSWORD = process.env.RESPONSIVE_PASSWORD
const CHROME_PATH = process.env.RESPONSIVE_CHROME_PATH
const OUT_DIR = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
  process.env.RESPONSIVE_OUT_DIR || ".shots/latest",
)

const VIEWPORTS = [
  { name: "phone-390", width: 390, height: 844, touch: true },
  { name: "ipad-768", width: 768, height: 1024, touch: true },
  { name: "ipad-1024", width: 1024, height: 768, touch: true },
  { name: "ipad-1180", width: 1180, height: 820, touch: true },
  { name: "desk-1440", width: 1440, height: 900, touch: false },
]
const GUEST_ROUTES = ["/login", "/register", "/forgot-password", "/movies/603"]
const AUTH_ROUTES = [
  "/", "/feed", "/notifications", "/search?q=batman", "/movies/603", "/tv/1399", "/upcoming",
  "/about-us", "/donate", "/settings", "/actors/287", "/onboarding", "/nope-404",
]
const SETTLE_MS = 2500
const NAV_TIMEOUT_MS = 20000
const LOGIN_TIMEOUT_MS = 15000
const OVERFLOW_TOLERANCE_PX = 1
const MAX_OFFENDERS = 6
const MONTAGE_COLUMNS = 4
const MONTAGE_THUMB_WIDTH = 300
const MONTAGE_GAP = 6

/** Runs in the page. Elements inside an overflow-x auto/scroll container are intentional scrollers and ignored. */
function measureOverflow({ vw, tolerance, maxOffenders }) {
  const inScroller = (el) => {
    for (let p = el.parentElement; p && p !== document.body; p = p.parentElement) {
      const ox = getComputedStyle(p).overflowX
      if (ox === "auto" || ox === "scroll" || ox === "hidden" || ox === "clip") return true
    }
    return false
  }
  const offenders = []
  for (const el of document.querySelectorAll("body *")) {
    const r = el.getBoundingClientRect()
    if (r.width === 0 || r.height === 0) continue
    const s = getComputedStyle(el)
    if (s.position === "fixed" || s.visibility === "hidden") continue
    if (r.right <= vw + tolerance && r.left >= -tolerance) continue
    if (inScroller(el)) continue
    const cls = typeof el.className === "string" ? el.className.split(" ")[0] : ""
    offenders.push(`${cls || el.tagName.toLowerCase()}(${Math.round(r.left)}..${Math.round(r.right)})`)
  }
  const scrollWidth = document.documentElement.scrollWidth
  return { scrollWidth, overflowX: scrollWidth > vw + tolerance, offenders: [...new Set(offenders)].slice(0, maxOffenders) }
}

async function login(ctx) {
  const page = await ctx.newPage()
  await page.goto(`${BASE_URL}/login`, { waitUntil: "domcontentloaded" })
  await page.locator("input").first().fill(USER)
  await page.locator('input[type="password"]').fill(PASSWORD)
  await page.locator('button[type="submit"]').click()
  await page.waitForURL((u) => !u.pathname.startsWith("/login"), { timeout: LOGIN_TIMEOUT_MS })
  await page.close()
}

const slug = (route) => route.replace(/[^a-z0-9]+/gi, "_").replace(/^_|_$/g, "") || "home"

async function montage(browser) {
  for (const vp of VIEWPORTS) {
    const files = readdirSync(OUT_DIR).filter((f) => f.startsWith(`${vp.name}__`) && f.endsWith(".png"))
    if (!files.length) continue
    const figures = files
      .map((f) => `<figure style="margin:0;background:#fff"><img src="${f}" style="width:${MONTAGE_THUMB_WIDTH}px;display:block"><figcaption>${f.replace(`${vp.name}__`, "").replace(".png", "")}</figcaption></figure>`)
      .join("")
    const html = `<body style="margin:0;background:#888;font:11px sans-serif"><div style="display:grid;grid-template-columns:repeat(${MONTAGE_COLUMNS},${MONTAGE_THUMB_WIDTH}px);gap:${MONTAGE_GAP}px;padding:${MONTAGE_GAP}px">${figures}</div></body>`
    const htmlFile = path.join(OUT_DIR, `montage-${vp.name}.html`)
    writeFileSync(htmlFile, html)
    const page = await browser.newPage({ viewport: { width: MONTAGE_COLUMNS * (MONTAGE_THUMB_WIDTH + MONTAGE_GAP) + MONTAGE_GAP, height: 800 } })
    await page.goto(`file://${htmlFile}`)
    await page.waitForLoadState("load")
    await page.screenshot({ path: path.join(OUT_DIR, `montage-${vp.name}.png`), fullPage: true })
    await page.close()
  }
}

const hasAuth = Boolean(USER && PASSWORD)
if (!hasAuth) console.log("RESPONSIVE_USER / RESPONSIVE_PASSWORD not set: skipping authed routes")
mkdirSync(OUT_DIR, { recursive: true })
const browser = await chromium.launch(CHROME_PATH ? { executablePath: CHROME_PATH, headless: true } : { channel: "chrome", headless: true })
const report = []

for (const vp of VIEWPORTS) {
  const groups = [{ label: "guest", routes: GUEST_ROUTES, auth: false }]
  if (hasAuth) groups.push({ label: "auth", routes: AUTH_ROUTES, auth: true })
  for (const { label, routes, auth } of groups) {
    const ctx = await browser.newContext({ viewport: { width: vp.width, height: vp.height }, isMobile: vp.touch, hasTouch: vp.touch, deviceScaleFactor: 1 })
    try {
      if (auth) await login(ctx)
    } catch (e) {
      report.push({ vp: vp.name, label, route: "(login)", error: String(e).slice(0, 120) })
      await ctx.close()
      continue
    }
    const page = await ctx.newPage()
    for (const route of routes) {
      try {
        await page.goto(`${BASE_URL}${route}`, { waitUntil: "domcontentloaded", timeout: NAV_TIMEOUT_MS })
        await page.waitForTimeout(SETTLE_MS)
        const ov = await page.evaluate(measureOverflow, { vw: vp.width, tolerance: OVERFLOW_TOLERANCE_PX, maxOffenders: MAX_OFFENDERS })
        const file = `${vp.name}__${label}__${slug(route)}.png`
        await page.screenshot({ path: path.join(OUT_DIR, file) })
        report.push({ vp: vp.name, label, route, ...ov, file })
      } catch (e) {
        report.push({ vp: vp.name, label, route, error: String(e).slice(0, 120) })
      }
    }
    await ctx.close()
  }
}

writeFileSync(path.join(OUT_DIR, "report.json"), JSON.stringify(report, null, 1))
if (process.argv.includes("--montage")) await montage(browser)
await browser.close()

let failed = 0
for (const r of report) {
  if (r.error) console.log(`ERROR ${r.vp} ${r.label} ${r.route}: ${r.error}`)
  else if (r.overflowX) {
    failed++
    console.log(`OVERFLOW ${r.vp} ${r.label} ${r.route} scrollWidth=${r.scrollWidth} ${r.offenders.join(" ")}`)
  } else if (r.offenders.length) console.log(`note ${r.vp} ${r.label} ${r.route}: clipped offenders ${r.offenders.join(" ")}`)
}
console.log(`${report.length} pages, ${failed} with document-level horizontal overflow. Output: ${OUT_DIR}`)
process.exit(failed ? 1 : 0)
