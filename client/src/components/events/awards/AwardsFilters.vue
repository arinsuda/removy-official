<template>
  <div class="awards-filters">
    <AppTabs
      v-model="mediaType"
      :tabs="mediaTabs"
      :label="t('awards.filters.label')"
      :controls="false"
    />
    <label class="awards-filters__year">
      <span class="awards-filters__label">{{ t("awards.filters.year") }}</span>
      <select
        class="awards-filters__select"
        :value="year"
        @change="onYearChange"
      >
        <option v-for="option in yearOptions" :key="option" :value="option">
          {{ formatAwardPeriod(String(option), locale, "long") }}
        </option>
      </select>
    </label>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import AppTabs from "@/components/ui/AppTabs.vue"
import type { AwardMediaType } from "@/types/event"
import { formatAwardPeriod } from "@/utils/events"

const props = defineProps<{ years: number[] }>()
const mediaType = defineModel<AwardMediaType>("mediaType", { required: true })
const year = defineModel<number>("year", { required: true })
const { t, locale } = useI18n()

const mediaTabs = computed<{ key: AwardMediaType; label: string }[]>(() => [
  { key: "movie", label: t("awards.filters.movie") },
  { key: "tv", label: t("awards.filters.tv") },
])

// The selected year is always an option, even before the first response lists the years.
const yearOptions = computed(() => (props.years.includes(year.value) ? props.years : [year.value, ...props.years]))

function onYearChange(event: Event) {
  year.value = Number((event.target as HTMLSelectElement).value)
}
</script>

<style scoped>
.awards-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem 1rem;
}

.awards-filters__year {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.awards-filters__label {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.awards-filters__select {
  padding: 0.375rem 0.75rem;
  font-size: var(--text-small);
  color: var(--color-text-primary);
  background-color: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-btn);
}
</style>
