import { computed, onBeforeUnmount, ref, watch, type Ref } from "vue"
import { notiApi } from "@/api/endpoints/notification"
import { FRESH_HIGHLIGHT_MS } from "@/components/notifications/notificationMeta"
import { useNotificationStore } from "@/stores/notification"
import type { AppNotification, NotificationCategory, NotificationPaginationMeta } from "@/types/notification"
import { logApiError } from "@/utils/logApiError"

const DEFAULT_PAGINATION: NotificationPaginationMeta = { page: 1, limit: 20, total: 0, total_pages: 0 }

/**
 * Filtered notification list for the full page. It owns its own page of items and
 * merges notifications that the socket pushes into the shared store while it is open.
 */
export function useNotificationList(category: Ref<NotificationCategory | null>, unreadOnly: Ref<boolean>) {
  const store = useNotificationStore()

  const items = ref<AppNotification[]>([])
  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref(false)
  const pagination = ref<NotificationPaginationMeta>({ ...DEFAULT_PAGINATION })
  const freshIds = ref<Set<number>>(new Set())
  const timers = new Set<ReturnType<typeof setTimeout>>()

  // Bumped on every reload so a slow response for an old filter cannot overwrite the new list.
  let generation = 0

  const hasMore = computed(() => pagination.value.page < pagination.value.total_pages)

  async function fetchList(page = 1) {
    if (page > 1 && (loading.value || loadingMore.value)) return
    const gen = generation
    if (page === 1) loading.value = true
    else loadingMore.value = true
    error.value = false

    try {
      const res = await notiApi.list({
        page,
        limit: pagination.value.limit,
        category: category.value ?? undefined,
        unread: unreadOnly.value || undefined,
      })
      if (gen !== generation) return
      const next = res.data.notifications ?? []
      items.value = page === 1 ? next : [...items.value, ...next]
      pagination.value = { ...DEFAULT_PAGINATION, ...res.data.pagination, page, limit: pagination.value.limit }
    } catch (err) {
      if (gen !== generation) return
      logApiError("fetchList (notifications page) failed:", err)
      error.value = true
    } finally {
      if (gen === generation) {
        loading.value = false
        loadingMore.value = false
      }
    }
  }

  function reload() {
    generation++
    pagination.value = { ...DEFAULT_PAGINATION }
    items.value = []
    return fetchList(1)
  }

  function loadMore() {
    if (hasMore.value && !loadingMore.value) return fetchList(pagination.value.page + 1)
  }

  async function markRead(n: AppNotification) {
    if (n.is_read) return
    n.is_read = true
    if (!(await store.markAsReadRemote(n.id))) n.is_read = false
  }

  async function markAllRead() {
    await store.markAllAsRead()
    items.value.forEach((n) => (n.is_read = true))
  }

  async function remove(n: AppNotification) {
    const idx = items.value.findIndex((x) => x.id === n.id)
    if (idx === -1) return
    items.value.splice(idx, 1)
    // The store only knows ids that are in its own list; keep the bell badge in sync either way.
    if (store.notifications.some((x) => x.id === n.id)) await store.removeNotification(n.id)
    else {
      try {
        await notiApi.deleteNotification(n.id)
        if (!n.is_read) await store.fetchUnreadCount()
      } catch (err) {
        items.value.splice(idx, 0, n)
        logApiError("remove notification failed:", err)
      }
    }
  }

  function matches(n: AppNotification) {
    if (category.value && n.category !== category.value) return false
    return !unreadOnly.value || !n.is_read
  }

  // A notification pushed over the socket lands at the top of the store list; mirror it here.
  watch(
    () => store.notifications[0]?.id,
    () => {
      const incoming = store.notifications[0]
      if (!incoming || loading.value || items.value.some((x) => x.id === incoming.id)) return
      const newest = items.value[0]?.id ?? 0
      if (incoming.id < newest || !matches(incoming)) return
      items.value.unshift(incoming)
      pagination.value.total++
      freshIds.value = new Set(freshIds.value).add(incoming.id)
      const timer = setTimeout(() => {
        const next = new Set(freshIds.value)
        next.delete(incoming.id)
        freshIds.value = next
        timers.delete(timer)
      }, FRESH_HIGHLIGHT_MS)
      timers.add(timer)
    },
  )

  watch([category, unreadOnly], reload, { immediate: true })
  onBeforeUnmount(() => timers.forEach(clearTimeout))

  return { items, loading, loadingMore, error, hasMore, freshIds, fetchList, reload, loadMore, markRead, markAllRead, remove }
}
