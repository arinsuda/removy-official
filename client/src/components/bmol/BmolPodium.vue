<template>
  <div class="podium">
    <section
      v-for="group in groups"
      :key="group.rank"
      class="podium__col"
      :class="[`podium__col--${tierOf(group.rank)}`, { 'is-target': targetRank === group.rank }]"
      :data-bmol-rank="group.rank"
      :aria-label="t('library.bmol.drag.rank', { rank: group.rank })"
    >
      <h3 class="podium__rank">{{ t("library.bmol.podium.rank", { rank: group.rank }) }}</h3>
      <div class="podium__cards">
        <slot :items="group.items" />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { tierOf, type RankGroup } from "./bmolRanks"

defineProps<{ groups: RankGroup[]; targetRank?: number | null }>()
const { t } = useI18n()
</script>

<style scoped>
.podium { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.625rem; align-items: start; }
.podium__col {
  --tier: var(--color-text-secondary);
  padding: 0.625rem;
  background: var(--color-surface-2);
  border: 1px solid color-mix(in srgb, var(--tier) 40%, var(--color-border));
  border-radius: var(--radius-card);
}
.podium__col--bronze { --tier: color-mix(in srgb, var(--color-warning) 55%, var(--color-brand)); }
.podium__col.is-target { outline: 2px dashed var(--tier); outline-offset: 3px; }
.podium__rank { margin-bottom: 0.75rem; font-size: var(--text-small); font-weight: 800; letter-spacing: 0.08em; text-transform: uppercase; color: var(--tier); }
.podium__cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(6rem, 1fr)); gap: 0.75rem; }
@media (min-width: 768px) {
  .podium { gap: 1rem; }
  .podium__col { padding: 1rem; }
  .podium__cards { grid-template-columns: repeat(auto-fill, minmax(8.5rem, 1fr)); gap: 1rem; }
}
</style>
