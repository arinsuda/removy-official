import { nextTick } from "vue";
import { i18n } from "./index";
import { LOCALE_STORAGE_KEY, LEGACY_LOCALE_STORAGE_KEY, DEFAULT_LOCALE, type SupportedLocale, isSupportedLocale } from "./types";
import api from "@/api";

const pendingLocaleLoads = new Map<SupportedLocale, Promise<void>>();

export async function setLocale(locale: SupportedLocale): Promise<boolean> {
  if (!isSupportedLocale(locale)) return false;

  try {
    await loadLocaleMessages(locale);
    (i18n.global.locale as unknown as { value: string }).value = locale;

    // Persist
    localStorage.setItem(LOCALE_STORAGE_KEY, locale);

    // Set document lang
    document.documentElement.setAttribute("lang", locale);

    // Set Accept-Language header
    api.defaults.headers.common["Accept-Language"] = locale;

    // Update title
    updateDocumentTitle();

    await nextTick();
    return true;
  } catch (err) {
    console.error(`Failed to switch locale to ${locale}`, err);
    return false;
  }
}

async function loadLocaleMessages(locale: SupportedLocale): Promise<void> {
  if (i18n.global.availableLocales.includes(locale)) {
    return;
  }

  let promise = pendingLocaleLoads.get(locale);
  if (!promise) {
    promise = (async () => {
      try {
        const messages = await import(`./locales/${locale}/index.ts`);
        i18n.global.setLocaleMessage(locale, messages.default);
      } finally {
        pendingLocaleLoads.delete(locale);
      }
    })();
    pendingLocaleLoads.set(locale, promise);
  }
  return promise;
}

export function getSavedLocale(): SupportedLocale {
  const saved = localStorage.getItem(LOCALE_STORAGE_KEY) ?? localStorage.getItem(LEGACY_LOCALE_STORAGE_KEY);
  if (isSupportedLocale(saved)) {
    return saved;
  }
  return DEFAULT_LOCALE;
}

export function updateDocumentTitle() {
  if (import.meta.env.MODE === "test") return;
  import("@/router").then(({ default: router }) => {
    const currentRoute = router.currentRoute.value;
    if (currentRoute && currentRoute.meta && currentRoute.meta.titleKey) {
      const titleKey = currentRoute.meta.titleKey as string;
      document.title = `${i18n.global.t(titleKey)} | POPCORN`;
    } else if (currentRoute && currentRoute.meta && typeof currentRoute.meta.title === "string") {
      document.title = `${currentRoute.meta.title} | POPCORN`;
    } else if (currentRoute && currentRoute.name) {
      const routeName = String(currentRoute.name);
      document.title = `${routeName.toUpperCase()} | POPCORN`;
    } else {
      document.title = "POPCORN";
    }
  });
}
