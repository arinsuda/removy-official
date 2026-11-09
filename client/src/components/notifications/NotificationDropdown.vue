<template>
  <div ref="root" class="noti-dropdown" role="dialog" :aria-label="t('notifications.latest')" @keydown.esc="emit('close')">
    <header class="noti-dropdown__head">
      <h2 class="noti-dropdown__title">{{ t("notifications.title") }}</h2>
      <AppButton v-if="store.unreadCount > 0" variant="ghost" size="sm" @click="store.markAllAsRead()">
        <CheckCheck :size="13" aria-hidden="true" />
        {{ t("notifications.markAllAsRead") }}
      </AppButton>
    </header>

    <div class="noti-dropdown__body">
      <SkeletonList v-if="store.loading" avatar :count="3" />

      <div v-else-if="store.error" class="noti-dropdown__state">
        <p>{{ t("notifications.error") }}</p>
        <AppButton variant="outline" size="sm" @click="store.refresh()">{{ t("common.retry") }}</AppButton>
      </div>

      <div v-else-if="!latest.length" class="noti-dropdown__state">
        <BellOff :size="26" aria-hidden="true" />
        <p>{{ t("notifications.empty") }}</p>
      </div>

      <ul v-else class="noti-dropdown__list">
        <NotificationItem
          v-for="n in latest"
          :key="n.id"
          :notification="n"
          compact
          @open="open"
          @remove="(x) => store.removeNotification(x.id)"
          @responded="(x) => store.markAsRead(x.id)"
        />
      </ul>
    </div>

    <footer class="noti-dropdown__foot">
      <RouterLink :to="{ name: 'notifications' }" class="noti-dropdown__all link-nudge" @click="emit('close')">
        {{ t("notifications.viewAll") }}
        <ChevronRight :size="14" aria-hidden="true" />
      </RouterLink>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import { BellOff, CheckCheck, ChevronRight } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useNotificationStore } from "@/stores/notification"
import { getNotificationActionUrl, type AppNotification } from "@/types/notification"
import NotificationItem from "./NotificationItem.vue"
import { DROPDOWN_LIMIT } from "./notificationMeta"

const emit = defineEmits<{ close: [] }>()

const { t } = useI18n()
const store = useNotificationStore()
const router = useRouter()
const root = ref<HTMLElement | null>(null)

const latest = computed(() => store.notifications.slice(0, DROPDOWN_LIMIT))

async function open(n: AppNotification) {
  await store.markAsRead(n.id)
  const url = getNotificationActionUrl(n)
  if (url) {
    router.push(url)
    emit("close")
  }
}

// Bubble-phase listener on purpose: the bell button stops propagation, so its own click never closes the panel.
function onDocumentClick(e: MouseEvent) {
  if (root.value && !root.value.contains(e.target as Node)) emit("close")
}
onMounted(() => document.addEventListener("click", onDocumentClick))
onBeforeUnmount(() => document.removeEventListener("click", onDocumentClick))
</script>

<style scoped>
.noti-dropdown {
  position: fixed;
  top: 4rem;
  right: 0.75rem;
  left: 0.75rem;
  z-index: var(--z-above-header);
  display: flex;
  max-height: min(32rem, calc(100dvh - 5rem));
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  box-shadow: var(--shadow-popover);
}
@media (min-width: 640px) {
  .noti-dropdown {
    position: absolute;
    top: calc(100% + 0.625rem);
    right: -0.5rem;
    left: auto;
    width: 22rem;
  }
}
.noti-dropdown__head {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
}
.noti-dropdown__title {
  margin: 0;
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
}
.noti-dropdown__body {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}
.noti-dropdown__list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.noti-dropdown__state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem 1rem;
  text-align: center;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
.noti-dropdown__state p {
  margin: 0;
}
.noti-dropdown__foot {
  flex-shrink: 0;
  padding: 0.5rem;
  border-top: 1px solid var(--color-border);
}
.noti-dropdown__all {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0.5rem;
  border-radius: 0.5rem;
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-secondary);
  text-decoration: none;
}
.noti-dropdown__all:hover {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}
</style>
