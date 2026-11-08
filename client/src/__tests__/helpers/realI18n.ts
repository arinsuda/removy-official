import { afterEach } from "vitest"
import { i18n, DEFAULT_LOCALE } from "@/i18n"

// The i18n instance is a singleton shared by every test in a file, so state
// changed here is restored in afterEach.
const { global: g } = i18n

/** Use the real locale bundles; any missing key throws instead of rendering the key. */
export function useRealI18n() {
  g.setMissingHandler((locale, key) => {
    throw new Error(`Missing i18n key "${key}" for locale "${String(locale)}"`)
  })
  afterEach(() => {
    g.locale.value = DEFAULT_LOCALE
    g.setMissingHandler(null)
  })
  return i18n
}

/** Assert every key resolves in both the en and th bundles. */
export function expectKeysInBothLocales(keys: string[]) {
  const missing = keys.flatMap((key) =>
    (["en", "th"] as const).filter((l) => !g.te(key, l)).map((l) => `${l}:${key}`),
  )
  if (missing.length) throw new Error(`Keys missing from locale bundles: ${missing.join(", ")}`)
}
