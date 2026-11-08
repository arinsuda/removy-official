import { ref, shallowRef, watch, type Ref } from "vue"
import { isAxiosError } from "axios"
import { mediaApi } from "@/api/endpoints/media"
import type { DetailConfig, DetailData } from "@/components/detail/detailConfig"

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

  async function load() {
    const seq = ++requestSeq
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

  watch(idRef, load, { immediate: true })

  return { status, data, reload: load }
}
