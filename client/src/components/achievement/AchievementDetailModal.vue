<template>
  <AppModal :model-value="!!achievement" size="sm" :title="achievement?.achievement.name" @close="emit('close')">
    <div v-if="achievement" class="ad">
      <span class="ad__icon" :class="{ 'is-on': achievement.is_unlocked }">
        <component :is="getAchievementIcon(achievement.achievement.action_type)" :size="40" :stroke-width="1.6" aria-hidden="true" />
        <Lock v-if="!achievement.is_unlocked" :size="14" class="ad__lock" aria-hidden="true" />
      </span>

      <ul class="ad__tags">
        <li class="ad__tag">{{ t(`achievements.categories.${category}`) }}</li>
        <li v-if="tiers.length > 1" class="ad__tag">{{ t("achievements.family.tier", { tier: TIER_NUMERALS[tierIndex] ?? tierIndex + 1 }) }}</li>
        <li v-if="achievement.achievement.exp_reward" class="ad__tag ad__tag--xp">{{ t("achievements.xp", { n: achievement.achievement.exp_reward }) }}</li>
      </ul>

      <section class="ad__block">
        <h3 class="ad__label">{{ t("achievements.detail.how") }}</h3>
        <p class="ad__desc">{{ achievement.achievement.description }}</p>
        <p class="ad__goal"><Target :size="14" aria-hidden="true" />{{ t("achievements.detail.goal", { target: n(achievement.achievement.target_count) }) }}</p>
      </section>

      <section class="ad__block">
        <h3 class="ad__label">{{ t("achievements.progressLabel") }}</h3>
        <span class="ad__bar" role="progressbar" :aria-valuenow="pct" aria-valuemin="0" aria-valuemax="100" :aria-label="t('achievements.progressLabel')">
          <span class="ad__fill" :style="{ width: `${pct}%` }" />
        </span>
        <p class="ad__state">
          {{ t("achievements.progress", { current: n(Math.min(achievement.current_count, achievement.achievement.target_count)), target: n(achievement.achievement.target_count) }) }}
          <span v-if="achievement.is_unlocked && achievement.unlocked_at" class="ad__date"><CheckCircle2 :size="14" aria-hidden="true" />{{ t("achievements.unlockedOn", { date: formatDate(achievement.unlocked_at) }) }}</span>
          <span v-else-if="achievement.is_unlocked" class="ad__date"><CheckCircle2 :size="14" aria-hidden="true" />{{ t("achievements.filters.unlocked") }}</span>
        </p>
      </section>

      <section v-if="tiers.length > 1" class="ad__block">
        <h3 class="ad__label">{{ t("achievements.detail.tiers") }}</h3>
        <ol class="ad__tiers">
          <li v-for="(tier, i) in tiers" :key="tier.achievement_id">
            <button type="button" class="ad__tier" :class="{ 'is-current': tier.achievement_id === achievement.achievement_id }" :aria-current="tier.achievement_id === achievement.achievement_id ? 'true' : undefined" @click="emit('select', tier)">
              <span class="ad__tier-n">{{ TIER_NUMERALS[i] ?? i + 1 }}</span>
              <span class="ad__tier-name">{{ tier.achievement.name }}</span>
              <CheckCircle2 v-if="tier.is_unlocked" :size="16" class="ad__tier-done" aria-hidden="true" />
              <span v-else class="ad__tier-pct">{{ n(progressOf(tier) / 100, "percent") }}</span>
            </button>
          </li>
        </ol>
      </section>

      <p v-if="isOwner && cta && !achievement.is_unlocked" class="ad__tip"><Lightbulb :size="14" aria-hidden="true" />{{ t("achievements.detail.tip") }}</p>
    </div>
    <template #footer>
      <AppButton v-if="isOwner && cta && achievement && !achievement.is_unlocked" variant="primary" :to="cta.to(userId)" @click="emit('close')">{{ t(cta.labelKey) }}</AppButton>
      <AppButton variant="outline" @click="emit('close')">{{ t("common.close") }}</AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { CheckCircle2, Lightbulb, Lock, Target } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import type { UserAchievement } from "@/types/achievement"
import { getAchievementIcon } from "@/utils/achievementIcons"
import { progressOf } from "./achievements"
import { TIER_NUMERALS, categoryOf, ctaFor } from "./config"

const props = defineProps<{
  achievement: UserAchievement | null
  /** Every tier of the same family, ascending; a single badge passes just itself. */
  tiers: UserAchievement[]
  userId: number
  isOwner: boolean
}>()
const emit = defineEmits<{ close: []; select: [achievement: UserAchievement] }>()

const { t, n, locale } = useI18n()
const category = computed(() => categoryOf(props.achievement?.achievement.action_type ?? ""))
const cta = computed(() => (props.achievement ? ctaFor(props.achievement.achievement.action_type) : null))
const pct = computed(() => (props.achievement ? progressOf(props.achievement) : 0))
const tierIndex = computed(() => props.tiers.findIndex((x) => x.achievement_id === props.achievement?.achievement_id))
const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "long" }).format(new Date(iso))
</script>

<style scoped>
.ad { display: flex; flex-direction: column; align-items: center; gap: 1rem; text-align: center; }
.ad__icon { position: relative; display: grid; place-items: center; width: 6rem; height: 6rem; overflow: hidden; color: var(--color-text-muted); background: var(--color-surface-3); border-radius: 50%; }
.ad__icon.is-on { color: var(--color-accent); background: var(--color-brand); box-shadow: 0 0 36px color-mix(in srgb, var(--color-brand) 55%, transparent); }
.ad__icon.is-on::after { content: ""; position: absolute; inset: 0; background: linear-gradient(115deg, transparent 35%, color-mix(in srgb, var(--color-text-primary) 45%, transparent) 50%, transparent 65%); transform: translateX(-120%); animation: ad-shine 3.2s var(--ease-water) 0.4s 2; }
@keyframes ad-shine { 0%, 55% { transform: translateX(-120%); } 100% { transform: translateX(120%); } }
.ad__lock { position: absolute; right: 0.5rem; bottom: 0.5rem; padding: 0.25rem; box-sizing: content-box; color: var(--color-text-secondary); background: var(--color-surface); border: 1px solid var(--color-border-strong); border-radius: 50%; }
.ad__tags { display: flex; flex-wrap: wrap; justify-content: center; gap: 0.5rem; margin: 0; padding: 0; list-style: none; }
.ad__tag { padding: 0.125rem 0.75rem; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-secondary); border: 1px solid var(--color-border-strong); border-radius: 9999px; }
.ad__tag--xp { color: var(--color-accent); }
.ad__block { display: flex; flex-direction: column; gap: 0.5rem; width: 100%; }
.ad__label { margin: 0; font-size: var(--text-caption); font-weight: 600; letter-spacing: 0.04em; color: var(--color-text-muted); }
.ad__desc { margin: 0; font-size: var(--text-small); line-height: 1.6; color: var(--color-text-secondary); }
.ad__goal { display: inline-flex; align-items: center; justify-content: center; gap: 0.375rem; margin: 0; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-primary); }
.ad__bar { display: block; width: 100%; height: 0.5rem; overflow: hidden; background: var(--color-surface-3); border-radius: 9999px; }
.ad__fill { display: block; height: 100%; background: var(--color-brand); border-radius: inherit; transition: width var(--motion-wave) var(--ease-out-soft); }
.ad__state { display: flex; flex-wrap: wrap; align-items: center; justify-content: center; gap: 0.25rem 0.75rem; margin: 0; font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.ad__date { display: inline-flex; align-items: center; gap: 0.25rem; font-size: var(--text-caption); font-weight: 600; color: var(--color-success); }
.ad__tiers { display: flex; flex-direction: column; gap: 0.375rem; margin: 0; padding: 0; list-style: none; }
.ad__tier { display: flex; align-items: center; gap: 0.75rem; width: 100%; min-height: 2.75rem; padding: 0 0.75rem; font: inherit; font-size: var(--text-small); text-align: start; color: var(--color-text-secondary); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); cursor: pointer; }
.ad__tier.is-current { color: var(--color-text-primary); border-color: var(--color-accent); }
.ad__tier-n { flex: none; min-width: 1.5rem; font-size: var(--text-caption); font-weight: 700; color: var(--color-text-muted); }
.ad__tier-name { flex: 1; min-width: 0; overflow-wrap: anywhere; }
.ad__tier-done { flex: none; color: var(--color-success); }
.ad__tier-pct { flex: none; font-size: var(--text-caption); }
.ad__tip { display: flex; align-items: center; gap: 0.5rem; margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
@media (hover: hover) {
  .ad__tier:hover { border-color: var(--color-border-strong); }
}
@media (prefers-reduced-motion: reduce) {
  .ad__icon.is-on::after { animation: none; }
  .ad__fill { transition: none; }
}
</style>
