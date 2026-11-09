import { computed, ref, watch, type Ref } from "vue"

export interface TabListConfig<T> {
  items: Ref<T[]>
  /** Text matched by the in-tab search. Omit to disable searching. */
  searchText?: (item: T) => string
  /** Named predicates. The key "all" (or any unknown key) keeps everything. */
  filters?: Record<string, (item: T) => boolean>
  sorts: Record<string, (a: T, b: T) => number>
  defaultSort: string
  /** Items shown before "Load more". Omit to show everything. */
  pageSize?: number
}

export interface TabListState {
  query: string
  filter: string
  sort: string
}

const normalize = (s: string) => s.trim().toLocaleLowerCase()

/** Pure search -> filter -> sort pipeline, shared by the composable and its tests. */
export function applyTabList<T>(items: T[], state: TabListState, cfg: Omit<TabListConfig<T>, "items" | "pageSize" | "defaultSort">): T[] {
  const q = normalize(state.query)
  const keep = cfg.filters?.[state.filter]
  const cmp = cfg.sorts[state.sort]
  const out = items.filter((item) => {
    if (keep && !keep(item)) return false
    return !q || !cfg.searchText || normalize(cfg.searchText(item)).includes(q)
  })
  // Array.prototype.sort is stable, so equal items keep the API order.
  return cmp ? out.sort(cmp) : out
}

export function useTabList<T>(cfg: TabListConfig<T>) {
  const query = ref("")
  const filter = ref("all")
  const sort = ref(cfg.defaultSort)
  const limit = ref(cfg.pageSize ?? Infinity)

  const filtered = computed(() => applyTabList(cfg.items.value, { query: query.value, filter: filter.value, sort: sort.value }, cfg))
  const visible = computed(() => filtered.value.slice(0, limit.value))
  const hasMore = computed(() => filtered.value.length > limit.value)
  const isFiltering = computed(() => normalize(query.value) !== "" || filter.value !== "all")

  const resetLimit = () => (limit.value = cfg.pageSize ?? Infinity)
  watch([query, filter, sort], resetLimit)

  function loadMore() {
    if (cfg.pageSize) limit.value += cfg.pageSize
  }
  function clear() {
    query.value = ""
    filter.value = "all"
  }

  /** Per-filter match counts over the unfiltered items, for chip badges. */
  const filterCounts = computed(() => {
    const counts: Record<string, number> = { all: cfg.items.value.length }
    for (const [key, fn] of Object.entries(cfg.filters ?? {})) counts[key] = cfg.items.value.filter(fn).length
    return counts
  })

  return { query, filter, sort, filtered, visible, hasMore, isFiltering, filterCounts, loadMore, clear }
}
