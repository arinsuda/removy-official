import { useQueryClient } from "@tanstack/vue-query"
import { useI18n } from "vue-i18n"
import { listApi } from "@/api/endpoints/list"
import { useToast } from "@/composables/useToast"
import type { CreateListRequest, ListSummary, MediaType, UpdateListRequest } from "@/types"
import { listErrorKey } from "./listErrors"

/** Root of every list query key, so one invalidation refreshes cards, sheets and counts. */
export const LISTS_KEY = ["lists"] as const

/**
 * Mutations on custom lists. Each resolves to the result or null after toasting the
 * translated error, so callers only branch on success.
 */
export function useListActions() {
  const { t } = useI18n()
  const toast = useToast()
  const client = useQueryClient()

  async function run<T>(task: () => Promise<T>): Promise<T | null> {
    try {
      const result = await task()
      await client.invalidateQueries({ queryKey: LISTS_KEY })
      return result
    } catch (err) {
      toast.error(t(listErrorKey(err)))
      return null
    }
  }

  return {
    create: (data: CreateListRequest) => run(() => listApi.create(data).then((r) => r.data.list)),
    update: (id: number, data: UpdateListRequest): Promise<ListSummary | null> => run(() => listApi.update(id, data).then((r) => r.data.list)),
    remove: (id: number) => run(() => listApi.remove(id).then(() => true)),
    addItem: (id: number, media: { media_id: number; media_type: MediaType }) => run(() => listApi.addItem(id, media).then((r) => r.data.item)),
    removeItem: (id: number, itemId: number) => run(() => listApi.removeItem(id, itemId).then(() => true)),
  }
}
