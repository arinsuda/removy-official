import { computed } from "vue"
import { useRoute, useRouter, type LocationQueryRaw, type RouteLocationRaw } from "vue-router"
import type { LibraryListKey, LibraryTab } from "@/types"
import { DEFAULT_LIBRARY_TAB, LIBRARY_LIST_KEYS, LIBRARY_TABS, LIST_QUERY_KEY, TAB_QUERY_KEY } from "./config"

const first = (v: unknown) => (Array.isArray(v) ? v[0] : v)

/** ?tab=<tab>&list=<list> <-> the active section. Updates replace history and keep the hash. */
export function useLibraryRoute() {
  const route = useRoute()
  const router = useRouter()

  // Old links used ?tab=<list key> (e.g. ?tab=watchlist); treat them as that list.
  const legacyList = computed<LibraryListKey | null>(() => {
    const q = first(route.query[TAB_QUERY_KEY])
    return LIBRARY_LIST_KEYS.find((k) => k === q) ?? null
  })
  const tab = computed<LibraryTab>(() => {
    if (legacyList.value) return "lists"
    const q = first(route.query[TAB_QUERY_KEY])
    return LIBRARY_TABS.find((t) => t === q) ?? DEFAULT_LIBRARY_TAB
  })
  const list = computed<LibraryListKey | null>(() => {
    if (tab.value !== "lists") return null
    if (legacyList.value) return legacyList.value
    const q = first(route.query[LIST_QUERY_KEY])
    return LIBRARY_LIST_KEYS.find((k) => k === q) ?? null
  })

  /** Target for a section; unrelated query keys and the hash are kept. */
  function target(next: { tab: LibraryTab; list?: LibraryListKey | null }): RouteLocationRaw {
    const { [TAB_QUERY_KEY]: _t, [LIST_QUERY_KEY]: _l, ...rest } = route.query
    const query: LocationQueryRaw = { ...rest }
    if (next.tab !== DEFAULT_LIBRARY_TAB) query[TAB_QUERY_KEY] = next.tab
    if (next.tab === "lists" && next.list) query[LIST_QUERY_KEY] = next.list
    return { query, hash: route.hash }
  }

  const setTab = (t: LibraryTab) => router.replace(target({ tab: t }))
  const openList = (k: LibraryListKey | null) => router.replace(target({ tab: "lists", list: k }))
  const listTo = (k: LibraryListKey) => target({ tab: "lists", list: k })
  const tabTo = (t: LibraryTab) => target({ tab: t })
  return { tab, list, setTab, openList, listTo, tabTo }
}
