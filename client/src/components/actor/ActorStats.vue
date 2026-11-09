<template>
  <dl class="actor-stats" :aria-label="t('actor.stats.label')">
    <div class="actor-stats__tile">
      <dt>{{ t("actor.stats.credits") }}</dt>
      <dd>{{ n(stats.total) }}</dd>
    </div>
    <div class="actor-stats__tile">
      <dt>{{ t("actor.stats.movies") }}</dt>
      <dd>{{ n(stats.movies) }}</dd>
    </div>
    <div class="actor-stats__tile">
      <dt>{{ t("actor.stats.tv") }}</dt>
      <dd>{{ n(stats.tv) }}</dd>
    </div>
    <div class="actor-stats__tile">
      <dt>{{ t("actor.stats.yearsActive") }}</dt>
      <dd>{{ years }}</dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { CreditStats } from "./actorCredits"

const props = defineProps<{ stats: CreditStats }>()
const { t, n } = useI18n()

// Years are plain digits in every locale, so they bypass number grouping ("2,010").
const years = computed(() => {
  const { firstYear, lastYear } = props.stats
  if (firstYear === null || lastYear === null) return t("actor.stats.none")
  if (firstYear === lastYear) return t("actor.stats.single", { year: firstYear })
  return t("actor.stats.range", { from: firstYear, to: lastYear })
})
</script>

<style scoped>
.actor-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin: 0;
}
.actor-stats__tile {
  padding: 0.875rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.actor-stats__tile dt {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.actor-stats__tile dd {
  margin: 0.125rem 0 0;
  font-size: var(--text-title);
  font-weight: 800;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

@media (min-width: 768px) {
  .actor-stats {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
