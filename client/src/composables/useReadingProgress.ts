import { computed, onScopeDispose, type Ref } from "vue"
import { useI18n } from "vue-i18n"
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query"
import { readingProgressApi } from "@/api/endpoints/readingProgress"
import {
  MANGA_PROGRESS_QUERY_KEY,
  MANGA_PROGRESS_SAVE_DEBOUNCE_MS,
  MANGA_PROGRESS_STALE_MS,
} from "@/components/manga/mangaConfig"
import {
  isEmptyProgress,
  progressErrorKey,
  progressValues,
  stepProgress,
  type ProgressField,
  type ProgressTotals,
} from "@/components/manga/mangaProgress"
import { useToast } from "@/composables/useToast"
import { useAuthStore } from "@/stores/auth"
import type { ReadingProgress, ReadingProgressInput } from "@/types/readingProgress"

/** `values: null` clears the progress (DELETE). `revision` is the local change the request belongs to. */
interface SaveRequest {
  values: ReadingProgressInput | null
  revision: number
}

/**
 * The signed-in viewer's reading position in one manga. Changes show at once in the query cache
 * and are saved after a short pause. A failed save restores the last saved position and shows a
 * toast. The caller keys its component by manga id, so the id stays fixed for this composable.
 */
export function useReadingProgress(mangaId: number, totals: Ref<ProgressTotals>) {
  const { t } = useI18n()
  const auth = useAuthStore()
  const toast = useToast()
  const queryClient = useQueryClient()
  // The viewer is part of the key so another account in the same tab never reads or saves this cache.
  const queryKey = computed(() => [MANGA_PROGRESS_QUERY_KEY, auth.user?.id ?? null, "manga", mangaId])

  const query = useQuery({
    queryKey,
    queryFn: async () => (await readingProgressApi.get(mangaId)).data.progress,
    enabled: computed(() => auth.isLoggedIn && mangaId > 0),
    staleTime: MANGA_PROGRESS_STALE_MS,
  })

  const values = computed(() => progressValues(query.data.value ?? null))

  /** Bumped by every local change. A response only writes the cache when it answers the latest one. */
  let revision = 0
  /** Saved position from before the first change that is not saved yet; restored when that save fails. */
  let unsaved: { before: ReadingProgress | null } | null = null
  let saveTimer: ReturnType<typeof setTimeout> | undefined

  const writeCache = (progress: ReadingProgress | null) => queryClient.setQueryData(queryKey.value, progress)

  function showChange(progress: ReadingProgress | null) {
    unsaved ??= { before: query.data.value ?? null }
    revision += 1
    writeCache(progress)
  }

  function rollback() {
    if (unsaved) writeCache(unsaved.before)
    unsaved = null
  }

  const save = useMutation({
    scope: { id: `reading-progress-${mangaId}` },
    mutationFn: async ({ values: next }: SaveRequest): Promise<ReadingProgress | null> => {
      if (!next) {
        await readingProgressApi.remove(mangaId)
        return null
      }
      return (await readingProgressApi.put(mangaId, next)).data.progress
    },
    onSuccess: (saved, { values: sent, revision: answers }) => {
      if (answers === revision) {
        unsaved = null
        writeCache(saved)
      } else if (unsaved) {
        // A newer change is still unsaved: a later failure must roll back to what the server holds now.
        unsaved.before = saved
      }
      if (!sent) toast.success(t("manga.progress.cleared"))
    },
    onError: (err, { revision: answers }) => {
      if (answers === revision) rollback()
      toast.error(t(progressErrorKey(err)))
    },
  })

  function flush() {
    saveTimer = undefined
    save.mutate({ values: values.value, revision })
  }

  function scheduleSave() {
    clearTimeout(saveTimer)
    saveTimer = setTimeout(flush, MANGA_PROGRESS_SAVE_DEBOUNCE_MS)
  }

  function cancelPendingSave() {
    clearTimeout(saveTimer)
    saveTimer = undefined
  }

  /** One stepper click: show it now, save once the clicks pause. Clicks that change nothing are ignored. */
  function step(field: ProgressField, delta: number) {
    const next = stepProgress(values.value, field, delta, totals.value)
    if (!next) return
    showChange({ media_id: mangaId, media_type: "manga", ...next, updated_at: new Date().toISOString() })
    scheduleSave()
  }

  /** Saves a pair typed in the edit dialog right away. */
  function commit(next: ReadingProgressInput) {
    const unchanged = next.chapter === values.value.chapter && next.volume === values.value.volume
    if (isEmptyProgress(next) || unchanged) return
    cancelPendingSave()
    showChange({ media_id: mangaId, media_type: "manga", ...next, updated_at: new Date().toISOString() })
    save.mutate({ values: next, revision })
  }

  function clearProgress() {
    cancelPendingSave()
    showChange(null)
    save.mutate({ values: null, revision })
  }

  // Leaving the page must not drop a change the viewer just made: save it now.
  onScopeDispose(() => {
    if (saveTimer !== undefined) flush()
  })

  return { query, values, step, commit, clearProgress }
}
