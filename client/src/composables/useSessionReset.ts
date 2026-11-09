import { useAdminStore } from "@/stores/admin"
import { useNotificationStore } from "@/stores/notification"

/** Clears the per-user state that outlives a session: admin mode, notifications and their socket. */
export function useSessionReset() {
  const admin = useAdminStore()
  const notifications = useNotificationStore()
  return () => {
    admin.reset()
    notifications.reset()
  }
}
