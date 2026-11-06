export const SUPPORTED_LOCALES = ["th", "en"] as const;
export type SupportedLocale = typeof SUPPORTED_LOCALES[number];

export const DEFAULT_LOCALE: SupportedLocale = "th";
export const LOCALE_STORAGE_KEY = "popcorn-locale";
// Read-only fallback so existing users keep their language after the rename.
export const LEGACY_LOCALE_STORAGE_KEY = "removy-locale";

export function isSupportedLocale(value: unknown): value is SupportedLocale {
  return typeof value === "string" && (SUPPORTED_LOCALES as readonly string[]).includes(value);
}
