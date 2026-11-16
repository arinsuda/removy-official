import { beforeEach, describe, expect, it } from "vitest"
import { i18n } from "@/i18n"
import { formatNotification } from "@/utils/notificationText"
import { getNotificationActionUrl, type AppNotification, type NotificationType } from "@/types/notification"
import { expectKeysInBothLocales, useRealI18n } from "./helpers/realI18n"

useRealI18n()

const ALL_TYPES: NotificationType[] = [
  "welcome",
  "followed_you",
  "follow_requested",
  "follow_accepted",
  "following_reviewed",
  "review_liked",
  "following_liked_review",
  "review_marked_helpful",
  "following_marked_helpful",
  "review_commented",
  "following_commented",
  "following_added_watched",
  "following_added_watchlist",
  "following_liked_media",
  "movie_now_playing",
  "achievement_unlocked",
  "email_verified",
  "password_changed",
  "event_results",
]

const g = i18n.global
const fmt = (n: AppNotification) => formatNotification(n, g)
const make = (over: Partial<AppNotification>): AppNotification => ({
  id: 1,
  type: "review_liked",
  category: "social",
  message: "",
  is_read: false,
  created_at: new Date().toISOString(),
  actor: { id: 3, username: "kenji", display_name: "Kenji" },
  params: { media_id: 603, media_type: "movie", media_title: "The Matrix" },
  ...over,
})

beforeEach(() => {
  g.locale.value = "en"
})

describe("formatNotification", () => {
  it("has an en and th text for every type, plus the fallbacks", () => {
    expectKeysInBothLocales([
      ...ALL_TYPES.map((t) => `notifications.types.${t}`),
      "notifications.someone",
      "notifications.untitled",
      "notifications.titles.event_results",
    ])
  })

  it("builds media text from the actor and the stored title", () => {
    expect(fmt(make({}))).toBe('Kenji liked your review of "The Matrix"')
    g.locale.value = "th"
    expect(fmt(make({}))).toContain("The Matrix")
    expect(fmt(make({}))).toContain("Kenji")
  })

  it("uses the actor username when there is no display name", () => {
    expect(fmt(make({ type: "followed_you", params: null, actor: { id: 3, username: "kenji" } }))).toBe("kenji started following you")
  })

  it("translates achievements by key and falls back to the stored name", () => {
    const base = { type: "achievement_unlocked" as const, category: "achievement" as const, actor: null }
    expect(fmt(make({ ...base, params: { action_type: "watched_count", target_count: 10, exp: 1500, name: "Movie Night" } }))).toBe(
      "Achievement unlocked: Movie Night (+1,500 EXP)",
    )
    expect(fmt(make({ ...base, params: { action_type: "nope", target_count: 1, exp: 5, name: "Server Name" } }))).toBe(
      "Achievement unlocked: Server Name (+5 EXP)",
    )
  })

  it("renders param-free types", () => {
    expect(fmt(make({ type: "email_verified", params: null, actor: null }))).toBe("Your email has been verified")
  })

  it("falls back to someone when the actor is gone", () => {
    expect(fmt(make({ actor: null }))).toBe('Someone liked your review of "The Matrix"')
  })

  it("falls back to untitled when a new row has no title", () => {
    expect(fmt(make({ params: { media_id: 1, media_type: "movie" } }))).toBe('Kenji liked your review of "a title"')
  })

  it("shows a legacy message when a row has no params", () => {
    expect(fmt(make({ params: null, message: "kenji liked your review of \"Dune\"" }))).toBe('kenji liked your review of "Dune"')
  })

  it("translates a legacy Thai welcome into the current language", () => {
    const legacy = make({ type: "welcome", category: "system", actor: null, params: null, message: "ยินดีต้อนรับสู่ POPCORN, kenji! เริ่มสร้าง Portfolio" })
    expect(fmt(legacy)).toBe("Welcome to POPCORN! Start building your movie watching portfolio.")
  })

  it("falls back to the title, then the generic title, for unknown types", () => {
    expect(fmt(make({ type: "email_verified_x" as NotificationType, params: null }))).toBe(g.t("notifications.title"))
    expect(fmt(make({ type: "review_liked", params: null, message: "" }))).toBe(g.t("notifications.titles.review_liked"))
  })
})

describe("event_results notifications", () => {
  const results = (params: AppNotification["params"]) =>
    make({ type: "event_results", category: "system", actor: null, target_ref: "event", target_id: 7, params })

  it("names the round from the params, in the current language", () => {
    const n = results({ event_type: "vote_monthly", media_type: "movie", period: "2026-11" })
    expect(fmt(n)).toContain("Popular Movie of the Month - November 2026")
    g.locale.value = "th"
    expect(fmt(n)).toContain("ประกาศผล")
  })

  it("falls back to the title when the params are malformed", () => {
    expect(fmt(results({ event_type: "vote_monthly", media_type: "movie", period: "bad" }))).toBe(
      g.t("notifications.titles.event_results"),
    )
  })
})

describe("getNotificationActionUrl", () => {
  it("routes tv and review targets to the title page", () => {
    expect(getNotificationActionUrl(make({ target_ref: "tv", target_id: 5 }))).toBe("/tv/5")
    expect(getNotificationActionUrl(make({ target_ref: "manga", target_id: 7 }))).toBe("/manga/7")
    expect(getNotificationActionUrl(make({ target_ref: "review", target_id: 9, params: { media_id: 7, media_type: "manga" } }))).toBe("/manga/7")
    expect(getNotificationActionUrl(make({ target_ref: "review", target_id: 9, params: { media_id: 603, media_type: "tv" } }))).toBe("/tv/603")
    expect(getNotificationActionUrl(make({ target_ref: "review", target_id: 9, params: null }))).toBeNull()
  })

  it("routes event results to the event page", () => {
    expect(getNotificationActionUrl(make({ target_ref: "event", target_id: 7, actor: null }))).toBe("/events/7")
  })
})
