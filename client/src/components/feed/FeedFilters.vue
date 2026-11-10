<template>
  <div class="feed-filters">
    <AppTabs v-model="scope" :tabs="scopeTabs" :label="t('feed.scopes.label')" :controls="false" />
    <AppTabs v-model="model" :tabs="tabs" :label="t('feed.filters.label')" :controls="false" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import AppTabs from "@/components/ui/AppTabs.vue"
import type { FeedScope } from "@/types/feed"
import { DEFAULT_FEED_SCOPE, FEED_CATEGORIES, FEED_SCOPES } from "./feedConfig"

const { t } = useI18n()

/** "" is the API's "all" category. */
const model = defineModel<string>({ required: true })
const scope = defineModel<FeedScope>("scope", { default: DEFAULT_FEED_SCOPE })

const tabs = computed(() => FEED_CATEGORIES.map((c) => ({ key: c, label: t(`feed.filters.${c || "all"}`) })))
const scopeTabs = computed(() => FEED_SCOPES.map((s) => ({ key: s, label: t(`feed.scopes.${s}`) })))
</script>

<style scoped>
.feed-filters {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
  overflow-x: auto;
  scrollbar-width: none;
}
.feed-filters::-webkit-scrollbar {
  display: none;
}
.feed-filters :deep(.app-tabs) {
  width: max-content;
}
.feed-filters :deep(.app-tabs__tab) {
  white-space: nowrap;
}
</style>
