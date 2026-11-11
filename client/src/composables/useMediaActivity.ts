import { computed, ref, watch, type Ref } from "vue"
import { useI18n } from "vue-i18n"
import { isAxiosError } from "axios"
import { libraryApi } from "@/api/endpoints/library"
import { mediaApi } from "@/api/endpoints/media"
import { reviewApi } from "@/api/endpoints/social"
import { watchLogApi } from "@/api/endpoints/watchLog"
import { useAuthGate, type AuthGateReason } from "@/composables/useAuthGate"
import { useToast } from "@/composables/useToast"
import { useAuthStore } from "@/stores/auth"
import type { MediaStats } from "@/types/movie"
import type { RatingBucket } from "@/types/review"
import type { WatchSummaryResponse } from "@/types/watchLog"

type MediaKind = "movie" | "tv"
type ListKind = "watchlist" | "watched"
type PendingAction = "like" | "watchlist" | "watched"

/**
 * The signed-in user's relationship to one title: stats, like, watchlist,
 * watched and watch-log summary. Actions always act on the id the ref holds
 * when they run; pass 0 to keep it idle.
 */
export function useMediaActivity(type: MediaKind, idRef: Ref<number>) {
  const { t } = useI18n()
  const auth = useAuthStore()
  const { requireAuth } = useAuthGate()
  const toast = useToast()

  const stats = ref<MediaStats | null>(null)
  const ratingDistribution = ref<RatingBucket[]>([])
  const itemIds = ref<Record<ListKind, number | null>>({ watchlist: null, watched: null })
  const watchSummary = ref<WatchSummaryResponse | null>(null)
  const pending = ref(new Set<PendingAction>())

  const isLiked = computed(() => Boolean(stats.value?.liked_at))
  const isWatchlisted = computed(() => itemIds.value.watchlist !== null)
  const isWatched = computed(() => itemIds.value.watched !== null)

  function reset() {
    stats.value = null
    ratingDistribution.value = []
    itemIds.value = { watchlist: null, watched: null }
    watchSummary.value = null
    pending.value = new Set()
  }

  async function loadStats(id: number) {
    try {
      const res = await mediaApi.getMediaStats(type, id)
      if (id !== idRef.value) return
      stats.value = res.data.stats
    } catch (err) {
      console.error("load media stats failed:", err)
    }
  }

  async function loadRating(id: number) {
    try {
      const res = await reviewApi.getMediaRating?.(type, id)
      if (id !== idRef.value) return
      ratingDistribution.value = res?.data?.rating?.distribution ?? []
    } catch {
      // Distribution is supplemental; keep empty on failure.
    }
  }

  async function loadLists(id: number) {
    if (!auth.isLoggedIn) return
    try {
      const res = await libraryApi.getOwnMediaStatus(id, type)
      if (id !== idRef.value) return
      const lists = res.data?.in_lists ?? []
      itemIds.value = {
        watchlist: lists.find(l => l.list_type === "watchlist")?.item_id ?? null,
        watched: lists.find(l => l.list_type === "watched")?.item_id ?? null,
      }
    } catch (err) {
      console.error("load library status failed:", err)
    }
  }

  async function loadWatchSummary(id: number) {
    if (!auth.isLoggedIn) return
    try {
      const res = await watchLogApi.getMyWatchLogs(type, id)
      if (id !== idRef.value) return
      watchSummary.value = res.data.summary
    } catch (err) {
      console.error("load watch summary failed:", err)
    }
  }

  async function load() {
    const id = idRef.value
    reset()
    if (!id) return
    void loadRating(id)
    await Promise.all([loadStats(id), loadLists(id), loadWatchSummary(id)])
  }

  /** Runs `fn` for the current id unless the caller is a guest or the same action is already in flight. */
  async function guarded(action: PendingAction, reason: AuthGateReason, fn: (id: number) => Promise<void>) {
    const id = idRef.value
    if (!id || !requireAuth(reason) || pending.value.has(action)) return
    pending.value.add(action)
    try {
      await fn(id)
    } catch (err) {
      console.error(`${action} failed:`, err)
      toast.error(t("media.detail.actionFailed"))
    } finally {
      pending.value.delete(action)
    }
  }

  function bumpStat(key: "like_count" | "watchlist_count", delta: number) {
    if (stats.value) stats.value = { ...stats.value, [key]: Math.max(0, stats.value[key] + delta) }
  }

  const toggleLike = () =>
    guarded("like", "like", async id => {
      if (isLiked.value) {
        await mediaApi.unlikeMedia(type, id)
        if (id !== idRef.value) return
        if (stats.value) stats.value = { ...stats.value, liked_at: null }
        bumpStat("like_count", -1)
      } else {
        await mediaApi.likeMedia(type, id)
        if (id !== idRef.value) return
        if (stats.value) stats.value = { ...stats.value, liked_at: new Date().toISOString() }
        bumpStat("like_count", 1)
      }
    })

  /** Adds to a list, treating "already in list" (409) as success. Returns the item id when known. */
  async function addToList(id: number, list: ListKind): Promise<number | null> {
    try {
      const res = await libraryApi.addItem({
        media_id: id,
        media_type: type,
        list_type: list,
        ...(list === "watched" ? { watched_at: new Date().toISOString().split("T")[0] } : {}),
      })
      return res.data?.item?.id ?? null
    } catch (err) {
      if (isAxiosError(err) && err.response?.status === 409) return null
      throw err
    }
  }

  const toggleWatchlist = () =>
    guarded("watchlist", "watchlist", async id => {
      const current = itemIds.value.watchlist
      if (current !== null) {
        await libraryApi.removeItem(current)
        if (id !== idRef.value) return
        itemIds.value = { ...itemIds.value, watchlist: null }
        bumpStat("watchlist_count", -1)
        return
      }
      const itemId = await addToList(id, "watchlist")
      if (id !== idRef.value) return
      itemIds.value = { ...itemIds.value, watchlist: itemId }
      bumpStat("watchlist_count", 1)
      if (itemId === null) await loadLists(id)
    })

  /** Called after the WatchLogModal saved a log: mark watched too, then refresh the summary. */
  const markLogged = () =>
    guarded("watched", "watchLog", async id => {
      if (itemIds.value.watched === null) {
        const itemId = await addToList(id, "watched")
        if (id === idRef.value) {
          itemIds.value = { ...itemIds.value, watched: itemId }
          if (itemId === null) await loadLists(id)
        }
      }
      await loadWatchSummary(id)
    })

  watch(
    [idRef, () => auth.isLoggedIn],
    () => {
      load()
    },
    { immediate: true },
  )

  return {
    stats,
    ratingDistribution,
    watchSummary,
    isLiked,
    isWatchlisted,
    isWatched,
    pending,
    requireLogin: requireAuth,
    toggleLike,
    toggleWatchlist,
    markLogged,
    reload: load,
    refreshStats: () => {
      if (!idRef.value) return Promise.resolve()
      void loadRating(idRef.value)
      return loadStats(idRef.value)
    },
  }
}
