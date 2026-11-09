<template>
  <div class="noti-filters">
    <div class="noti-filters__tabs">
      <AppTabs v-model="category" :tabs="tabs" :label="t('notifications.filters.label')" :controls="false" />
    </div>
    <button type="button" class="noti-filters__unread" :class="{ 'is-on': unreadOnly }" :aria-pressed="unreadOnly" @click="unreadOnly = !unreadOnly">
      <span class="noti-filters__dot" aria-hidden="true" />
      {{ t("notifications.filters.unreadOnly") }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import AppTabs from "@/components/ui/AppTabs.vue"
import { NOTIFICATION_CATEGORIES } from "@/types/notification"

const { t } = useI18n()

/** "all" or a NotificationCategory. */
const category = defineModel<string>("category", { required: true })
const unreadOnly = defineModel<boolean>("unreadOnly", { required: true })

const tabs = computed(() =>
  ["all", ...NOTIFICATION_CATEGORIES].map((key) => ({ key, label: t(`notifications.categories.${key}`) })),
)
</script>

<style scoped>
.noti-filters {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}
.noti-filters__tabs {
  overflow-x: auto;
  scrollbar-width: none;
}
.noti-filters__tabs::-webkit-scrollbar {
  display: none;
}
.noti-filters__tabs :deep(.app-tabs) {
  width: max-content;
}
.noti-filters__tabs :deep(.app-tabs__tab) {
  white-space: nowrap;
}
.noti-filters__unread {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  gap: 0.5rem;
  padding: 0.375rem 0.875rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-btn);
  background: var(--color-surface-2);
  font-size: var(--text-caption);
  font-weight: 500;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition:
    color var(--motion-fast) ease,
    border-color var(--motion-fast) ease;
}
.noti-filters__unread:hover {
  border-color: var(--color-border-strong);
  color: var(--color-text-primary);
}
.noti-filters__unread.is-on {
  border-color: var(--color-brand);
  color: var(--color-text-primary);
}
.noti-filters__dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 9999px;
  background: var(--color-text-muted);
}
.is-on .noti-filters__dot {
  background: var(--color-brand);
}

@media (min-width: 768px) {
  .noti-filters {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}
</style>
