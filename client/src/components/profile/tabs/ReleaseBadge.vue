<template>
  <span v-if="status !== 'unknown'" class="rb" :class="`rb--${status}`" data-release-status :data-status="status">
    <component :is="status === 'upcoming' ? CalendarClock : CircleCheck" :size="12" aria-hidden="true" />
    <span class="rb__text">{{ status === "upcoming" ? t("media.release.upcoming", { date: formatted }) : t("media.release.released") }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { CalendarClock, CircleCheck } from "lucide-vue-next"
import { formatReleaseDate, releaseStatus } from "@/utils/releaseStatus"

const props = defineProps<{ date: string | null | undefined }>()
const { t, locale } = useI18n()
const status = computed(() => releaseStatus(props.date))
const formatted = computed(() => formatReleaseDate(props.date, locale.value))
</script>

<style scoped>
.rb { display: inline-flex; align-items: center; gap: 0.25rem; min-width: 0; padding: 0.125rem 0.5rem; font-size: var(--text-caption); font-weight: 600; border-radius: 9999px; }
.rb__text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rb--upcoming { color: var(--color-accent); background: color-mix(in srgb, var(--color-accent) 14%, transparent); border: 1px solid color-mix(in srgb, var(--color-accent) 40%, transparent); }
.rb--released { color: var(--color-text-secondary); background: var(--color-surface-3); border: 1px solid var(--color-border); }
</style>
