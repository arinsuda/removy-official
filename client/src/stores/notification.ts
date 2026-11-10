import { defineStore } from "pinia"
import { ref, computed } from "vue"
import { notiApi } from "@/api/endpoints/notification"
import { useSocket } from "@/composables/useSocket"
import { i18n } from "@/i18n"
import { useUIStore } from "@/stores/ui"
import { logApiError } from "@/utils/logApiError"
import { formatNotification } from "@/utils/notificationText"
import {
  notificationTitleKey,
  type AppNotification,
  type NotificationPaginationMeta,
  type SocketNotificationDeletedPayload,
  type SocketNotificationNewPayload,
  type SocketNotificationReadPayload,
} from "@/types/notification"

export const useNotificationStore = defineStore("notification", () => {
  const DEFAULT_PAGINATION: NotificationPaginationMeta = {
    page: 1,
    limit: 20,
    total: 0,
    total_pages: 0,
  }
  const notifications = ref<AppNotification[]>([])
  const unreadCount = ref(0)
  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref<string | null>(null)
  const isPanelOpen = ref(false)

  const pagination = ref<NotificationPaginationMeta>({ ...DEFAULT_PAGINATION })

  const hasMore = computed(
    () => pagination.value.page < pagination.value.total_pages,
  )

  let socketBound = false

  async function fetchNotifications(page = 1) {
    if (loading.value || loadingMore.value) return

    const limit = pagination.value?.limit ?? DEFAULT_PAGINATION.limit

    if (page === 1) loading.value = true
    else loadingMore.value = true

    error.value = null

    try {
      const res = await notiApi.list({ page, limit })

      const nextNotifications = res.data.notifications ?? []

      notifications.value =
        page === 1
          ? nextNotifications
          : [...notifications.value, ...nextNotifications]

      unreadCount.value = res.data.unread_count ?? 0

      pagination.value = {
        ...DEFAULT_PAGINATION,
        ...res.data.pagination,
        page,
        limit,
      }
    } catch (err) {
      logApiError("fetchNotifications failed:", err)
      error.value = "load_failed"
    } finally {
      loading.value = false
      loadingMore.value = false
    }
  }

  async function loadMore() {
    if (!hasMore.value || loadingMore.value) return
    await fetchNotifications(pagination.value.page + 1)
  }

  async function fetchUnreadCount() {
    try {
      const res = await notiApi.getUnreadCount()
      unreadCount.value = res.data.unread_count
    } catch (err) {
      logApiError("fetchUnreadCount failed:", err)
    }
  }

  async function refresh() {
    await fetchNotifications(1)
  }

  async function markAsRead(id: number) {
    const target = notifications.value.find(n => n.id === id)
    if (!target || target.is_read) return

    target.is_read = true
    unreadCount.value = Math.max(0, unreadCount.value - 1)

    try {
      await notiApi.markAsRead(id)
    } catch (err) {
      target.is_read = false
      unreadCount.value++
      logApiError("markAsRead failed:", err)
    }
  }

  /**
   * ใช้จากหน้าอื่นที่มี list เป็นของตัวเอง (เช่น NotificationsView.vue)
   * ต่างจาก markAsRead ตรงที่ไม่ต้องพึ่งว่า id นั้นอยู่ใน `notifications` ของ panel หรือเปล่า
   * ยังไงก็ยิง API จริง แล้ว sync `unreadCount` กลับมาให้ badge ที่กระดิ่งถูกต้อง
   * คืนค่า boolean กลับไปให้ผู้เรียกใช้ตัดสินใจ revert local state ของตัวเอง
   */
  async function markAsReadRemote(id: number): Promise<boolean> {
    const target = notifications.value.find(n => n.id === id)
    const wasUnreadInPanel = !!target && !target.is_read

    if (target) target.is_read = true
    if (unreadCount.value > 0) unreadCount.value--

    try {
      await notiApi.markAsRead(id)
      return true
    } catch (err) {
      if (target && wasUnreadInPanel) target.is_read = false
      unreadCount.value++
      logApiError("markAsReadRemote failed:", err)
      return false
    }
  }

  async function markAllAsRead() {
    if (unreadCount.value === 0) return
    const prevState = notifications.value.map(n => n.is_read)
    const prevCount = unreadCount.value

    notifications.value.forEach(n => (n.is_read = true))
    unreadCount.value = 0

    try {
      await notiApi.markAllAsRead()
    } catch (err) {
      notifications.value.forEach((n, i) => (n.is_read = prevState[i] ?? false))
      unreadCount.value = prevCount
      logApiError("markAllAsRead failed:", err)
    }
  }

  async function removeNotification(id: number) {
    const idx = notifications.value.findIndex(n => n.id === id)
    if (idx === -1) return

    const [removed] = notifications.value.splice(idx, 1)
    if (!removed) return

    if (!removed.is_read) {
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
    pagination.value.total = Math.max(0, pagination.value.total - 1)

    try {
      await notiApi.deleteNotification(id)
    } catch (err) {
      notifications.value.splice(idx, 0, removed)
      if (!removed.is_read) unreadCount.value++
      pagination.value.total++
      logApiError("removeNotification failed:", err)
    }
  }

  async function clearAll() {
    const prev = notifications.value
    const prevUnread = unreadCount.value
    notifications.value = []
    unreadCount.value = 0

    try {
      await notiApi.clearAll()
    } catch (err) {
      notifications.value = prev
      unreadCount.value = prevUnread
      logApiError("clearAll failed:", err)
    }
  }

  function prependNotification(n: AppNotification) {
    if (notifications.value.some(x => x.id === n.id)) return
    notifications.value.unshift(n)
    if (!n.is_read) unreadCount.value++
    pagination.value.total++
  }

  function handleRemoteRead({ ids }: SocketNotificationReadPayload) {
    ids.forEach(id => {
      const target = notifications.value.find(n => n.id === id)
      if (target && !target.is_read) {
        target.is_read = true
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
    })
  }

  function handleRemoteDelete({ ids }: SocketNotificationDeletedPayload) {
    if (!ids || ids.length === 0) {
      notifications.value = []
      unreadCount.value = 0
      pagination.value.total = 0
      return
    }
    ids.forEach(id => {
      const idx = notifications.value.findIndex(n => n.id === id)
      if (idx === -1) return
      const [removed] = notifications.value.splice(idx, 1)
      if (removed && !removed.is_read) {
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
    })
  }

  // A dropped socket misses pushes, so when the tab becomes visible again, resync the
  // badge over HTTP (a 401 there refreshes the token, which resumes the socket).
  function onVisibilityChange() {
    if (document.visibilityState !== "visible") return
    const { isConnected, resume } = useSocket()
    if (isConnected.value) return
    resume()
    fetchUnreadCount()
  }

  function bindSocket() {
    if (socketBound) return
    const { connect, on } = useSocket()
    connect()
    document.addEventListener("visibilitychange", onVisibilityChange)

    on<SocketNotificationNewPayload>("notification:new", payload => {
      prependNotification(payload)
      useUIStore().addToast(
        i18n.global.t("notifications.toast", {
          title: i18n.global.te(notificationTitleKey(payload.type))
            ? i18n.global.t(notificationTitleKey(payload.type))
            : i18n.global.t("notifications.title"),
          message: formatNotification(payload, i18n.global),
        }),
        "info",
      )
    })

    on<SocketNotificationReadPayload>("notification:read", handleRemoteRead)

    on<SocketNotificationDeletedPayload>(
      "notification:deleted",
      handleRemoteDelete,
    )

    on<{ unread_count: number }>(
      "notification:unread_count",
      ({ unread_count }) => {
        unreadCount.value = unread_count
      },
    )
    socketBound = true
  }

  function unbindSocket() {
    const { disconnect } = useSocket()
    disconnect()
    document.removeEventListener("visibilitychange", onVisibilityChange)
    socketBound = false
  }

  function togglePanel() {
    const nextOpen = !isPanelOpen.value
    isPanelOpen.value = nextOpen

    if (!nextOpen) return

    if (
      notifications.value.length === 0 &&
      !loading.value &&
      !loadingMore.value
    ) {
      fetchNotifications(1)
    }
  }

  function closePanel() {
    isPanelOpen.value = false
  }

  function reset() {
    notifications.value = []
    unreadCount.value = 0
    pagination.value = { ...DEFAULT_PAGINATION }
    isPanelOpen.value = false
    unbindSocket()
  }

  return {
    notifications,
    unreadCount,
    loading,
    loadingMore,
    error,
    pagination,
    hasMore,
    isPanelOpen,

    fetchNotifications,
    loadMore,
    fetchUnreadCount,
    refresh,
    markAsRead,
    markAsReadRemote,
    markAllAsRead,
    removeNotification,
    clearAll,
    prependNotification,

    bindSocket,
    unbindSocket,
    togglePanel,
    closePanel,
    reset,
  }
})
