/**
 * @vitest-environment happy-dom
 */

import { describe, it, expect, beforeEach} from "vitest";
import { i18n, getSavedLocale, setLocale, useLocale, formatRelativeTime } from "../i18n";
import { LOCALE_STORAGE_KEY } from "../i18n/types";

describe("Locale Management and Switcher", () => {
  beforeEach(() => {
    localStorage.clear();
    (i18n.global.locale as unknown as { value: string }).value = "th";
    document.documentElement.removeAttribute("lang");
  });

  it("bootstraps with the default Thai locale", () => {
    expect(getSavedLocale()).toBe("th");
  });

  it("respects a valid saved English locale", () => {
    localStorage.setItem(LOCALE_STORAGE_KEY, "en");
    expect(getSavedLocale()).toBe("en");
  });

  it("falls back to default if saved locale is invalid", () => {
    localStorage.setItem(LOCALE_STORAGE_KEY, "invalid_locale");
    expect(getSavedLocale()).toBe("th");
  });

  it("updates HTML lang attribute and persists locale after setLocale()", async () => {
    const success = await setLocale("en");
    expect(success).toBe(true);
    expect(i18n.global.locale.value).toBe("en");
    expect(localStorage.getItem(LOCALE_STORAGE_KEY)).toBe("en");
    expect(document.documentElement.getAttribute("lang")).toBe("en");
  });

  it("switching toggleLocale() works dynamically", async () => {
    const { toggleLocale, locale } = useLocale();
    expect(locale.value).toBe("th");

    await toggleLocale();
    expect(locale.value).toBe("en");
    expect(document.documentElement.getAttribute("lang")).toBe("en");

    await toggleLocale();
    expect(locale.value).toBe("th");
    expect(document.documentElement.getAttribute("lang")).toBe("th");
  });

  it("formats relative time correctly in Thai and English", () => {
    const now = new Date().getTime();
    const tenSecondsAgo = now - 10000;
    const tenMinutesAgo = now - 10 * 60000;

    expect(formatRelativeTime(tenSecondsAgo, "th")).toBe("10 วินาทีที่แล้ว");
    expect(formatRelativeTime(tenSecondsAgo, "en")).toBe("10 seconds ago");

    expect(formatRelativeTime(tenMinutesAgo, "th")).toBe("10 นาทีที่แล้ว");
    expect(formatRelativeTime(tenMinutesAgo, "en")).toBe("10 minutes ago");
  });

  it("resource integrity contains identical keys for errors", () => {
    const thErrors = i18n.global.getLocaleMessage("th").errors;
    const enErrors = i18n.global.getLocaleMessage("en").errors;
    
    expect(thErrors).toBeDefined();
    expect(enErrors).toBeDefined();
    
    const thKeys = Object.keys(thErrors).sort();
    const enKeys = Object.keys(enErrors).sort();
    
    expect(thKeys).toEqual(enKeys);
  });
});
