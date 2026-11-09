<template>
  <div class="w-76 rounded-2xl border border-border bg-surface-2 p-5 shadow-2xl">
    <h3 class="mb-1 text-sm font-semibold text-text-primary">{{ t("settings.activity.title") }}</h3>
    <p class="mb-4 text-xs leading-relaxed text-text-muted">{{ t("settings.activity.description") }}</p>
    <ActivityToggleList />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import ActivityToggleList from "@/components/settings/ActivityToggleList.vue"

const { t } = useI18n()
</script>
