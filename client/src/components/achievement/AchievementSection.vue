<template>
  <section class="as" :aria-labelledby="headId">
    <header class="as__head">
      <span class="as__icon"><component :is="CATEGORIES[section.key].icon" :size="18" aria-hidden="true" /></span>
      <div class="as__text">
        <h2 :id="headId" class="as__title">{{ t(`achievements.categories.${section.key}`) }}</h2>
        <p class="as__count">{{ t("achievements.counter", { unlocked: n(section.unlocked), total: n(section.total) }) }}</p>
      </div>
      <span class="as__bar" aria-hidden="true"><span class="as__fill" :style="{ width: `${pct}%` }" /></span>
    </header>
    <ul class="as__grid" :class="{ 'as__grid--few': isFew, 'as__grid--single': section.units.length === 1 && !hasFamily }">
      <li v-for="(unit, i) in section.units" :key="unit.key" class="as__cell" :class="{ 'as__cell--wide': unit.tiers.length > 1, 'as__cell--solo': soloIndices.has(i) }" :style="{ '--i': Math.min(i, STAGGER_STEPS) }">
        <AchievementFamily v-if="unit.tiers.length > 1" :unit="unit" @open="emit('open', $event)" />
        <AchievementBadge v-else :achievement="unit.current" shine @open="emit('open', $event)" />
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { UserAchievement } from "@/types/achievement"
import AchievementBadge from "./AchievementBadge.vue"
import AchievementFamily from "./AchievementFamily.vue"
import { CATEGORIES, FEW_UNITS, STAGGER_STEPS } from "./config"
import type { CategorySummary } from "./units"

const props = defineProps<{ section: CategorySummary }>()
const emit = defineEmits<{ open: [achievement: UserAchievement] }>()

const { t, n } = useI18n()
const headId = computed(() => `achievements-section-${props.section.key}`)
const isFew = computed(() => props.section.units.length <= FEW_UNITS)
const hasFamily = computed(() => props.section.units.some((u) => u.tiers.length > 1))
const pct = computed(() => (props.section.total ? (props.section.unlocked / props.section.total) * 100 : 0))
const soloIndices = computed(() => {
  const set = new Set<number>()
  const list = props.section.units
  let singlesRun = 0
  for (let i = 0; i < list.length; i++) {
    if (list[i]!.tiers.length > 1) {
      singlesRun = 0
      continue
    }
    singlesRun++
    const next = list[i + 1]
    if (singlesRun % 2 === 1 && (!next || next.tiers.length > 1)) {
      set.add(i)
    }
  }
  return set
})
</script>

<style scoped>
.as { display: flex; flex-direction: column; gap: 0.875rem; }
.as__head { display: grid; grid-template-columns: auto 1fr; align-items: center; gap: 0.25rem 0.75rem; }
.as__icon { display: grid; place-items: center; width: 2.25rem; height: 2.25rem; color: var(--color-accent); background: var(--color-surface-3); border-radius: 50%; }
.as__text { display: flex; flex-wrap: wrap; align-items: baseline; gap: 0.25rem 0.75rem; min-width: 0; }
.as__title { margin: 0; font-size: var(--text-body); font-weight: 700; color: var(--color-text-primary); }
.as__count { margin: 0; font-size: var(--text-caption); color: var(--color-text-secondary); }
.as__bar { grid-column: 1 / -1; display: block; height: 0.25rem; overflow: hidden; background: var(--color-surface-3); border-radius: 9999px; }
.as__fill { display: block; height: 100%; background: var(--color-brand); border-radius: inherit; transition: width var(--motion-base) var(--ease-out-soft); }
.as__grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr)); gap: 0.75rem; margin: 0; padding: 0; list-style: none; }
.as__cell { display: flex; flex-direction: column; min-width: 0; animation: as-in var(--motion-base) var(--ease-out-soft) calc(var(--i) * 30ms) both; }
.as__cell > * { flex: 1; }
.as__cell--wide, .as__cell--solo { grid-column: 1 / -1; }
@keyframes as-in { from { opacity: 0; transform: translateY(8px); } }
@media (min-width: 480px) {
  .as__cell--solo { grid-column: auto; }
  .as__cell--wide { grid-column: span 2; }
}
@media (min-width: 768px) {
  .as__grid--few .as__cell--wide { grid-column: auto; }
  .as__grid--single { max-width: 20rem; }
}
@media (prefers-reduced-motion: reduce) {
  .as__cell { animation: none; }
  .as__fill { transition: none; }
}
</style>
