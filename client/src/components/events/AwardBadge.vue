<template>
  <span class="award-badge" role="img" :aria-label="label" :title="label">
    <span aria-hidden="true">🏆</span>
    <span aria-hidden="true">{{ shortPeriod }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { MediaAward } from "@/types/event"
import { formatAwardPeriod } from "@/utils/events"

const props = defineProps<{ award: MediaAward }>()
const { t, locale } = useI18n()

const shortPeriod = computed(() => formatAwardPeriod(props.award.period, locale.value, "short"))
const label = computed(() =>
  t("awards.badge.label", { period: formatAwardPeriod(props.award.period, locale.value, "long") }),
)
</script>

<style scoped>
.award-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.15rem 0.6rem;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-rating);
  background: color-mix(in srgb, var(--color-surface-3) 80%, transparent);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-btn);
  white-space: nowrap;
}
</style>
