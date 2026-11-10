<template>
  <li class="noti" :class="{ 'is-unread': !notification.is_read, 'is-fresh': fresh, 'is-compact': compact }">
    <button type="button" class="noti__main" @click="emit('open', notification)">
      <span class="noti__icon" :class="`noti__icon--${notification.category}`" aria-hidden="true">
        <component :is="notificationIcon(notification.type)" :size="compact ? 16 : 18" />
      </span>
      <span class="noti__body">
        <span class="noti__title">
          {{ title }}
          <span v-if="fresh" class="noti__new">{{ t("notifications.new") }}</span>
        </span>
        <span class="noti__message">{{ message }}</span>
        <span class="noti__time">
          <time :datetime="notification.created_at">{{ relativeTime }}</time>
          <span v-if="!notification.is_read" class="sr-only">{{ t("notifications.unread") }}</span>
        </span>
      </span>
      <span v-if="!notification.is_read" class="noti__dot" aria-hidden="true" />
    </button>

    <button type="button" class="noti__remove" :aria-label="t('notifications.remove')" @click="emit('remove', notification)">
      <X :size="14" aria-hidden="true" />
    </button>

    <div v-if="isRequest" class="noti__actions">
      <template v-if="requestState === 'accepted' || requestState === 'declined'">
        <p class="noti__result" role="status">{{ t(`notifications.actions.${requestState}`) }}</p>
      </template>
      <template v-else>
        <AppButton size="sm" :loading="requestState === 'busy'" @click="respond(true)">{{ t("notifications.actions.accept") }}</AppButton>
        <AppButton size="sm" variant="outline" :disabled="requestState === 'busy'" @click="respond(false)">
          {{ t("notifications.actions.decline") }}
        </AppButton>
      </template>
    </div>
  </li>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { X } from "lucide-vue-next"
import { followApi } from "@/api/api"
import AppButton from "@/components/ui/AppButton.vue"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import { notificationTitleKey, type AppNotification } from "@/types/notification"
import { logApiError } from "@/utils/logApiError"
import { formatRelativeTime } from "@/utils/relativeTime"
import { formatNotification } from "@/utils/notificationText"
import { notificationIcon } from "./notificationMeta"

const props = defineProps<{
  notification: AppNotification
  /** Tighter layout for the header dropdown. */
  compact?: boolean
  /** Just arrived over the socket; plays the arrival highlight. */
  fresh?: boolean
}>()

const emit = defineEmits<{
  open: [notification: AppNotification]
  remove: [notification: AppNotification]
  /** A follow request was answered; the parent marks the notification read. */
  responded: [notification: AppNotification]
}>()

const { t, te, n, locale } = useI18n()
const auth = useAuthStore()
const ui = useUIStore()

const title = computed(() => {
  const key = notificationTitleKey(props.notification.type)
  return te(key) ? t(key) : t("notifications.title")
})
const message = computed(() => formatNotification(props.notification, { t, te, n }))
const relativeTime = computed(() => formatRelativeTime(props.notification.created_at, locale.value))

const isRequest = computed(() => props.notification.type === "follow_requested" && !!props.notification.actor)
const requestState = ref<"idle" | "busy" | "accepted" | "declined">("idle")

async function respond(accept: boolean) {
  const me = auth.user?.id
  const follower = props.notification.actor?.id
  if (!me || !follower || requestState.value === "busy") return
  requestState.value = "busy"
  try {
    await (accept ? followApi.acceptRequest(me, follower) : followApi.rejectRequest(me, follower))
    requestState.value = accept ? "accepted" : "declined"
    emit("responded", props.notification)
  } catch (err) {
    requestState.value = "idle"
    logApiError("follow request response failed:", err)
    ui.addToast(t("notifications.actions.failed"), "error")
  }
}
</script>

<style scoped>
.noti {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  transition:
    background-color var(--motion-fast) ease,
    border-color var(--motion-fast) ease;
}
.noti:hover {
  background: var(--color-surface-3);
}
.noti.is-unread {
  border-color: color-mix(in srgb, var(--color-brand) 35%, transparent);
  background: color-mix(in srgb, var(--color-brand) 7%, var(--color-surface-2));
}
.noti.is-fresh {
  animation: noti-arrive var(--motion-wave-text) var(--ease-out-soft), noti-glow 2.4s ease-out;
}
.noti__main {
  display: flex;
  flex: 1;
  min-width: 0;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.875rem 2.75rem 0.875rem 0.875rem;
  border: 0;
  border-radius: inherit;
  background: none;
  text-align: left;
  font: inherit;
  color: inherit;
  cursor: pointer;
}
.noti__icon {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 9999px;
  background: var(--color-surface-3);
  color: var(--color-text-secondary);
}
.is-compact .noti__icon {
  width: 2rem;
  height: 2rem;
}
.noti__icon--social { color: var(--color-info); }
.noti__icon--media { color: var(--color-brand); }
.noti__icon--achievement { color: var(--color-warning); }
.noti__icon--system { color: var(--color-success); }
.noti__body {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 0.125rem;
}
.noti__title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
}
.noti__new {
  padding: 0 0.5rem;
  border-radius: var(--radius-btn);
  background: var(--color-brand);
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-on-brand);
}
.noti__message {
  font-size: var(--text-small);
  line-height: 1.45;
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}
.is-compact .noti__message {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.noti__time {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.noti__dot {
  position: absolute;
  top: 1.1rem;
  right: 2.75rem;
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 9999px;
  background: var(--color-brand);
}
.noti__remove {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  border: 0;
  border-radius: 9999px;
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color var(--motion-fast) ease,
    background-color var(--motion-fast) ease;
}
.noti__remove:hover {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}
.noti__actions {
  display: flex;
  flex-basis: 100%;
  align-items: center;
  gap: 0.5rem;
  padding: 0 0.875rem 0.875rem 4.125rem;
}
.is-compact .noti__actions {
  padding-left: 3.625rem;
}
.noti__result {
  margin: 0;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-success);
}

@keyframes noti-arrive {
  from {
    opacity: 0;
    transform: translateY(-12px);
  }
}
@keyframes noti-glow {
  from {
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-brand) 60%, transparent);
  }
  to {
    box-shadow: 0 0 0 2px transparent;
  }
}

@media (prefers-reduced-motion: reduce) {
  .noti.is-fresh {
    animation: none;
  }
}
</style>
