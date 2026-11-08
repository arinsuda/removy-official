import { onScopeDispose, ref, shallowRef, watch, type Ref } from "vue"
import { isAxiosError } from "axios"
import { mediaApi } from "@/api/endpoints/media"
import type { DetailConfig, DetailData } from "@/components/detail/detailConfig"

const RATINGS_RETRY_DELAY = 1500
const RATINGS_MAX_RETRIES = 2

export type DetailStatus = "loading" | "ready" | "error" | "notFound"

/**
 * Loads one title for DetailPage. When the id changes, the previous data stays
 * in `data` (status "loading") so the page can show a progress bar instead of
 * blanking; only the newest request may write state.
 */
export function useMediaDetail(config: DetailConfig, idRef: Ref<number>) {
  const status = ref<DetailStatus>("loading")
  const data = shallowRef<DetailData | null>(null)
  let requestSeq = 0
  const viewRecorded = new Set<number>()
  let retryTimer: ReturnType<typeof setTimeout> | null = null

  function cancelRatingsRetry() {
    if (retryTimer) clearTimeout(retryTimer)
    retryTimer = null
  }

  /** IMDb / Rotten Tomatoes may still be loading server-side; re-ask and merge only ratings. */
  function scheduleRatingsRetry(id: number, seq: number, attempt: number) {
    if (attempt >= RATINGS_MAX_RETRIES) return
    cancelRatingsRetry()
    retryTimer = setTimeout(async () => {
      retryTimer = null
      try {
        const fresh = await config.fetch(id)
        if (seq !== requestSeq || !data.value) return
        data.value = { ...data.value, ratings: fresh.ratings }
        if (fresh.ratings?.external_pending) scheduleRatingsRetry(id, seq, attempt + 1)
      } catch {
        // Silent: keep the ratings already shown.
      }
    }, RATINGS_RETRY_DELAY)
  }

  async function load() {
    const seq = ++requestSeq
    cancelRatingsRetry()
    const id = idRef.value
    if (!Number.isInteger(id) || id <= 0) {
      data.value = null
      status.value = "notFound"
      return
    }

    status.value = "loading"
    try {
      const result = await config.fetch(id)
      if (seq !== requestSeq) return
      data.value = result
      status.value = "ready"
      if (result.ratings?.external_pending) scheduleRatingsRetry(id, seq, 0)
      if (!viewRecorded.has(id)) {
        viewRecorded.add(id)
        mediaApi.recordMediaView(config.type, id).catch(err => console.warn("record view failed:", err))
      }
    } catch (err) {
      if (seq !== requestSeq) return
      const code = isAxiosError(err) ? err.response?.status : undefined
      if (code === 404 || code === 400) {
        data.value = null
        status.value = "notFound"
      } else {
        console.error("load media detail failed:", err)
        status.value = "error"
      }
    }
  }

  onScopeDispose(cancelRatingsRetry)
  watch(idRef, load, { immediate: true })

  return { status, data, reload: load }
}
