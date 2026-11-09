import { computed } from "vue"
import { useRoute, useRouter, type LocationQueryRaw, type RouteLocationRaw } from "vue-router"
import type { LibraryListKey, LibraryTab } from "@/types"
import { CUSTOM_LIST_PREFIX, DEFAULT_LIBRARY_TAB, LIBRARY_LIST_KEYS, LIBRARY_TABS, LIST_QUERY_KEY, TAB_QUERY_KEY } from "./config"

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

  // ?list=custom-<id> opens one of the user's own made lists.
  const customListId = computed<number | null>(() => {
    if (tab.value !== "lists") return null
    const q = first(route.query[LIST_QUERY_KEY])
    if (typeof q !== "string" || !q.startsWith(CUSTOM_LIST_PREFIX)) return null
    const id = Number(q.slice(CUSTOM_LIST_PREFIX.length))
    return Number.isInteger(id) && id > 0 ? id : null
  })

  /** Target for a section; unrelated query keys and the hash are kept. */
  function target(next: { tab: LibraryTab; list?: LibraryListKey | null; customId?: number | null }): RouteLocationRaw {
    const { [TAB_QUERY_KEY]: _t, [LIST_QUERY_KEY]: _l, ...rest } = route.query
    const query: LocationQueryRaw = { ...rest }
    if (next.tab !== DEFAULT_LIBRARY_TAB) query[TAB_QUERY_KEY] = next.tab
    if (next.tab === "lists" && next.list) query[LIST_QUERY_KEY] = next.list
    if (next.tab === "lists" && next.customId) query[LIST_QUERY_KEY] = `${CUSTOM_LIST_PREFIX}${next.customId}`
    return { query, hash: route.hash }
  }

  const setTab = (t: LibraryTab) => router.replace(target({ tab: t }))
  const openList = (k: LibraryListKey | null) => router.replace(target({ tab: "lists", list: k }))
  const openCustomList = (id: number | null) => router.replace(target({ tab: "lists", customId: id }))
  const customListTo = (id: number) => target({ tab: "lists", customId: id })
  const listTo = (k: LibraryListKey) => target({ tab: "lists", list: k })
  const tabTo = (t: LibraryTab) => target({ tab: t })
  return { tab, list, customListId, setTab, openList, openCustomList, listTo, customListTo, tabTo }
}
