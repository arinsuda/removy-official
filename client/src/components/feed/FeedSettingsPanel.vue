<template>
  <div class="feed-settings-panel">
    <h3 class="feed-settings-panel__title">{{ t("settings.activity.title") }}</h3>
    <p class="feed-settings-panel__desc">{{ t("settings.activity.description") }}</p>
    <ActivityToggleList />
    <RouterLink :to="{ name: 'settings', hash: '#activity' }" class="feed-settings-panel__link link-nudge">
      {{ t("feed.settings.manage") }}
      <ChevronRight :size="14" aria-hidden="true" />
    </RouterLink>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ChevronRight } from "lucide-vue-next"
import ActivityToggleList from "@/components/settings/ActivityToggleList.vue"

const { t } = useI18n()
</script>

<style scoped>
.feed-settings-panel {
  width: 20rem;
  max-width: 100%;
  padding: 1.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  box-shadow: var(--shadow-popover);
}
.feed-settings-panel__title {
  margin: 0 0 0.25rem;
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-primary);
}
.feed-settings-panel__desc {
  margin: 0 0 1rem;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.feed-settings-panel__link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 1rem;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-secondary);
  text-decoration: none;
}
.feed-settings-panel__link:hover {
  color: var(--color-brand);
}
</style>
