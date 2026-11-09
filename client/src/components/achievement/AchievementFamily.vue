<template>
  <article class="af" :class="{ 'af--done': unit.status === 'unlocked' }">
    <button type="button" class="af__main" :aria-label="`${current.name}. ${summary}`" @click="emit('open', unit.current)">
      <span class="af__icon" :class="{ 'af__icon--on': doneCount > 0 }">
        <component :is="getAchievementIcon(current.action_type)" :size="24" :stroke-width="1.7" aria-hidden="true" />
      </span>
      <span class="af__text">
        <span class="af__name">{{ current.name }}</span>
        <span class="af__sub">{{ summary }}</span>
      </span>
    </button>

    <ol class="af__track" :aria-label="t('achievements.family.tiers', { done: doneCount, total: unit.tiers.length })">
      <li v-for="(tier, i) in unit.tiers" :key="tier.achievement_id" class="af__step" :class="{ 'is-done': tier.is_unlocked, 'is-current': tier === unit.current && unit.status !== 'unlocked', 'is-linked': i > 0 && unit.tiers[i - 1]?.is_unlocked, 'is-dim': isDim(tier) }">
        <button type="button" class="af__node" :aria-current="tier === unit.current ? 'step' : undefined" :aria-label="`${tier.achievement.name}. ${t('achievements.family.tier', { tier: TIER_NUMERALS[i] ?? i + 1 })}`" @click="emit('open', tier)">
          <span class="af__dot">
            <Check v-if="tier.is_unlocked" :size="14" :stroke-width="3" aria-hidden="true" />
            <template v-else>{{ TIER_NUMERALS[i] ?? i + 1 }}</template>
          </span>
        </button>
      </li>
    </ol>

    <span class="af__bar" role="progressbar" :aria-valuenow="unit.pct" aria-valuemin="0" aria-valuemax="100" :aria-label="t('achievements.progressLabel')">
      <span class="af__fill" :style="{ width: `${unit.pct}%` }" />
    </span>
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { Check } from "lucide-vue-next"
import type { UserAchievement } from "@/types/achievement"
import { getAchievementIcon } from "@/utils/achievementIcons"
import { TIER_NUMERALS } from "./config"
import type { AchievementUnit } from "./units"

const props = defineProps<{ unit: AchievementUnit }>()
const emit = defineEmits<{ open: [achievement: UserAchievement] }>()

const { t, n } = useI18n()
const current = computed(() => props.unit.current.achievement)
/** Tiers outside the active status/search filter fade so the matching ones stand out. */
const isDim = (tier: UserAchievement) => !!props.unit.matched && props.unit.matched.size < props.unit.tiers.length && !props.unit.matched.has(tier.achievement_id)
const doneCount = computed(() => props.unit.tiers.filter((x) => x.is_unlocked).length)
const summary = computed(() =>
  props.unit.status === "unlocked"
    ? t("achievements.family.complete", { total: props.unit.tiers.length })
    : t("achievements.progress", { current: n(Math.min(props.unit.current.current_count, current.value.target_count)), target: n(current.value.target_count) }),
)
</script>

<style scoped>
.af { display: flex; flex-direction: column; gap: 0.875rem; min-width: 0; padding: 1rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.af--done { border-color: color-mix(in srgb, var(--color-brand) 55%, transparent); background: linear-gradient(to bottom, color-mix(in srgb, var(--color-brand) 14%, transparent), transparent 70%), var(--color-surface-2); }
.af__main { display: flex; align-items: center; gap: 0.875rem; min-width: 0; padding: 0; font: inherit; text-align: start; color: inherit; background: none; border: 0; cursor: pointer; }
.af__icon { display: grid; flex: none; place-items: center; width: 3.25rem; height: 3.25rem; color: var(--color-text-muted); background: var(--color-surface-3); border-radius: 50%; }
.af__icon--on { color: var(--color-accent); background: var(--color-brand); box-shadow: 0 0 18px color-mix(in srgb, var(--color-brand) 45%, transparent); }
.af__text { display: flex; flex-direction: column; gap: 0.125rem; min-width: 0; }
.af__name { font-size: var(--text-small); font-weight: 700; line-height: 1.3; color: var(--color-text-primary); overflow-wrap: anywhere; }
.af__sub { font-size: var(--text-caption); color: var(--color-text-secondary); }
.af__track { display: flex; align-items: center; margin: 0; padding: 0.25rem 0; overflow-x: auto; list-style: none; scrollbar-width: none; }
.af__track::-webkit-scrollbar { display: none; }
.af__step { position: relative; display: flex; flex: 1 0 auto; justify-content: flex-end; min-width: 2.75rem; }
.af__step:first-child { flex: 0 0 auto; min-width: 0; }
.af__step::before { content: ""; position: absolute; inset-block-start: 50%; inset-inline: 0 2.75rem; height: 2px; background: var(--color-surface-3); transform: translateY(-50%); }
.af__step:first-child::before { display: none; }
.af__step.is-dim .af__node { opacity: 0.4; }
.af__step.is-linked::before { background: var(--color-brand); }
.af__node { position: relative; display: grid; place-items: center; width: 2.75rem; height: 2.75rem; padding: 0; font: inherit; font-size: var(--text-caption); font-weight: 700; color: var(--color-text-muted); background: none; border: 0; border-radius: 50%; cursor: pointer; }
.af__dot { display: grid; place-items: center; width: 2rem; height: 2rem; background: var(--color-surface-3); border: 1px solid var(--color-border-strong); border-radius: 50%; }
.af__step.is-done .af__node { color: var(--color-on-brand); }
.af__step.is-done .af__dot { background: var(--color-brand); border-color: var(--color-brand); }
.af__step.is-current .af__node { color: var(--color-text-primary); }
.af__step.is-current .af__dot { border-color: var(--color-accent); box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent) 25%, transparent); }
.af__bar { display: block; height: 0.3125rem; overflow: hidden; background: var(--color-surface-3); border-radius: 9999px; }
.af__fill { display: block; height: 100%; background: var(--color-brand); border-radius: inherit; transition: width var(--motion-base) var(--ease-out-soft); }
@media (hover: hover) {
  .af__node:hover .af__dot { border-color: var(--color-text-secondary); }
}
@media (prefers-reduced-motion: reduce) {
  .af__fill { transition: none; }
}
</style>
