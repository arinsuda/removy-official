<template>
  <div class="feed-filters">
    <AppTabs v-model="model" :tabs="tabs" :label="t('feed.filters.label')" :controls="false" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import AppTabs from "@/components/ui/AppTabs.vue"
import { FEED_CATEGORIES } from "./feedConfig"

const { t } = useI18n()

/** "" is the API's "all" category. */
const model = defineModel<string>({ required: true })

const tabs = computed(() => FEED_CATEGORIES.map((c) => ({ key: c, label: t(`feed.filters.${c || "all"}`) })))
</script>

<style scoped>
.feed-filters {
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
