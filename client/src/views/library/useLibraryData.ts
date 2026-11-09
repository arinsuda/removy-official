import { ref, watch, type Ref } from "vue"
import { isAxiosError } from "axios"
import { libraryApi, reviewApi, userApi } from "@/api/api"
import { mediaApi } from "@/api/endpoints/media"
import { summaryToHomeMedia } from "@/components/profile/tabs/profileMedia"
import type { ReviewResponse } from "@/types"
import type { UserProfile } from "@/types/user"
import { toEntries, toWatchedEntries, type LibraryEntry } from "./libraryStats"

/** Everything the library overview needs. Each source fails on its own so one bad call never blanks the page. */
export function useLibraryData(userId: Ref<number>) {
  const profile = ref<UserProfile | null>(null)
  const watchlist = ref<LibraryEntry[]>([])
  const watched = ref<LibraryEntry[]>([])
  const likes = ref<LibraryEntry[]>([])
  const reviews = ref<ReviewResponse[]>([])
  const loading = ref(true)
  const profileFailed = ref(false)
  const listsFailed = ref(false)
  /** The owner keeps the library private from this viewer (403 on a list). */
  const locked = ref(false)

  /** `quiet` refreshes counts in place (e.g. after leaving a list) without the skeleton. */
  async function load(quiet = false) {
    const id = userId.value
    if (!quiet) loading.value = true
    const [p, wl, wd, lk, rv] = await Promise.allSettled([
      userApi.getProfile(id),
      libraryApi.getVisibleUserLibrary(id, { list_type: "watchlist" }),
      libraryApi.getVisibleUserLibrary(id, { list_type: "watched" }),
      mediaApi.getLikeByUserId(id),
      reviewApi.getUserReviews(id),
    ])
    if (id !== userId.value) return
    if (p.status === "fulfilled") profile.value = p.value.data.user
    profileFailed.value = p.status === "rejected"
    if (wl.status === "fulfilled") watchlist.value = toEntries(wl.value.data.items)
    if (wd.status === "fulfilled") watched.value = toWatchedEntries(wd.value.data.items)
    if (lk.status === "fulfilled") {
      likes.value = lk.value.data.likes.map((l) => ({ id: l.media.id, summary: l.media, media: summaryToHomeMedia(l.media), at: l.created_at }))
    }
    if (rv.status === "fulfilled") reviews.value = rv.value.data.reviews ?? []
    const lists = [wl, wd, lk, rv]
    locked.value = lists.some((r) => r.status === "rejected" && isAxiosError(r.reason) && r.reason.response?.status === 403)
    listsFailed.value = lists.some((r) => r.status === "rejected")
    loading.value = false
  }

  watch(userId, () => void load(), { immediate: true })
  return { profile, watchlist, watched, likes, reviews, loading, profileFailed, listsFailed, locked, load }
}
