import { computed, ref, toValue, watch, type MaybeRefOrGetter } from "vue"
import { listApi } from "@/api/endpoints/list"
import { httpStatus } from "@/components/lists/listErrors"
import type { ListItem, ListSummary } from "@/types"

export type ListLoadError = "notFound" | "forbidden" | "failed" | null

/** One custom list with its items, loaded page by page. Reordering needs every item, see loadAll. */
export function useCustomListDetail(listId: MaybeRefOrGetter<number>) {
  const list = ref<ListSummary | null>(null)
  const items = ref<ListItem[]>([])
  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref<ListLoadError>(null)
  const page = ref(1)
  const hasMore = ref(false)
  const total = ref(0)
  let seq = 0

  async function load() {
    const id = toValue(listId)
    const mine = ++seq
    loading.value = true
    error.value = null
    try {
      const { data } = await listApi.get(id, 1)
      if (mine !== seq) return
      list.value = data.list
      items.value = data.items ?? []
      page.value = data.pagination.page
      hasMore.value = data.pagination.has_more
      total.value = data.pagination.total
    } catch (err) {
      if (mine !== seq) return
      const status = httpStatus(err)
      error.value = status === 404 ? "notFound" : status === 403 ? "forbidden" : "failed"
      list.value = null
      items.value = []
    } finally {
      if (mine === seq) loading.value = false
    }
  }

  async function loadMore(): Promise<boolean> {
    if (!hasMore.value) return true
    if (loadingMore.value) return false
    const mine = seq
    loadingMore.value = true
    try {
      const { data } = await listApi.get(toValue(listId), page.value + 1)
      if (mine !== seq) return false
      const known = new Set(items.value.map((i) => i.id))
      items.value = [...items.value, ...data.items.filter((i) => !known.has(i.id))]
      page.value = data.pagination.page
      hasMore.value = data.pagination.has_more
      total.value = data.pagination.total
      return true
    } catch {
      return false
    } finally {
      loadingMore.value = false
    }
  }

  /** Reorder sends the whole list, so every page must be here first. Returns false if a page failed. */
  async function loadAll(): Promise<boolean> {
    while (hasMore.value) if (!(await loadMore())) return false
    return true
  }

  /** Applies an optimistic change; returns a function that restores the previous items. */
  function mutate(next: ListItem[]) {
    const prev = items.value
    items.value = next
    return () => (items.value = prev)
  }

  watch(() => toValue(listId), load, { immediate: true })

  return { list, items, loading, loadingMore, error, hasMore, total, load, loadMore, loadAll, mutate }
}
