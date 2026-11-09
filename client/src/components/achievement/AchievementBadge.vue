<template>
  <button type="button" class="ab" :class="{ 'ab--on': achievement.is_unlocked, 'ab--shine': shine }" :aria-label="`${achievement.achievement.name}. ${status}`" @click="emit('open', achievement)">
    <span class="ab__icon">
      <component :is="getAchievementIcon(achievement.achievement.action_type)" :size="26" :stroke-width="1.7" aria-hidden="true" />
      <Lock v-if="!achievement.is_unlocked" :size="12" class="ab__lock" aria-hidden="true" />
    </span>
    <span class="ab__name">{{ achievement.achievement.name }}</span>
    <span v-if="achievement.is_unlocked" class="ab__sub">{{ status }}</span>
    <template v-else>
      <span class="ab__bar" role="progressbar" :aria-valuenow="pct" aria-valuemin="0" aria-valuemax="100" :aria-label="t('profile.tabContent.achievements.progressLabel')">
        <span class="ab__fill" :style="{ width: `${pct}%` }" />
      </span>
      <span class="ab__sub">{{ status }}</span>
    </template>
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { Lock } from "lucide-vue-next"
import type { UserAchievement } from "@/types/achievement"
import { getAchievementIcon } from "@/utils/achievementIcons"
import { progressOf } from "./achievements"

const props = defineProps<{ achievement: UserAchievement; shine?: boolean }>()
const emit = defineEmits<{ open: [achievement: UserAchievement] }>()

const { t, locale } = useI18n()
const pct = computed(() => progressOf(props.achievement))
const status = computed(() => {
  const a = props.achievement
  if (a.is_unlocked && a.unlocked_at) {
    return t("profile.tabContent.achievements.unlockedOn", { date: new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(a.unlocked_at)) })
  }
  if (a.is_unlocked) return t("profile.tabContent.achievements.filters.unlocked")
  return t("profile.tabContent.achievements.progress", { current: Math.min(a.current_count, a.achievement.target_count), target: a.achievement.target_count })
})
</script>

<style scoped>
.ab { display: flex; flex-direction: column; align-items: center; gap: 0.4rem; min-width: 0; padding: 1rem 0.625rem 0.875rem; font: inherit; text-align: center; color: var(--color-text-muted); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); cursor: pointer; transition: transform var(--motion-fast) var(--ease-out-soft), border-color var(--motion-fast) var(--ease-water); }
.ab__icon { position: relative; display: grid; place-items: center; width: 3.5rem; height: 3.5rem; color: var(--color-text-muted); background: var(--color-surface-3); border-radius: 50%; filter: grayscale(1); opacity: 0.8; }
.ab__lock { position: absolute; right: -0.125rem; bottom: -0.125rem; padding: 0.2rem; box-sizing: content-box; color: var(--color-text-secondary); background: var(--color-surface); border: 1px solid var(--color-border-strong); border-radius: 50%; }
.ab--on { border-color: color-mix(in srgb, var(--color-brand) 55%, transparent); background: linear-gradient(to bottom, color-mix(in srgb, var(--color-brand) 14%, transparent), transparent 70%), var(--color-surface-2); }
.ab--on .ab__icon { color: var(--color-accent); background: var(--color-brand); filter: none; opacity: 1; box-shadow: 0 0 20px color-mix(in srgb, var(--color-brand) 50%, transparent); }
.ab--shine.ab--on .ab__icon { overflow: hidden; }
.ab--shine.ab--on .ab__icon::after { content: ""; position: absolute; inset: 0; background: linear-gradient(115deg, transparent 35%, color-mix(in srgb, var(--color-text-primary) 45%, transparent) 50%, transparent 65%); transform: translateX(-120%); animation: ab-shine 3.2s var(--ease-water) 0.6s 2; }
@keyframes ab-shine { 0%, 55% { transform: translateX(-120%); } 100% { transform: translateX(120%); } }
.ab__name { display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; font-size: var(--text-small); font-weight: 700; line-height: 1.3; color: var(--color-text-primary); overflow-wrap: anywhere; }
.ab:not(.ab--on) .ab__name { color: var(--color-text-secondary); }
.ab__sub { font-size: var(--text-caption); }
.ab__bar { display: block; width: 100%; height: 0.3125rem; overflow: hidden; background: var(--color-surface-3); border-radius: 9999px; }
.ab__fill { display: block; height: 100%; background: var(--color-text-secondary); border-radius: inherit; transition: width var(--motion-base) var(--ease-out-soft); }
@media (hover: hover) {
  .ab:hover { transform: translateY(-3px); border-color: var(--color-border-strong); }
  .ab--on:hover { border-color: var(--color-brand); }
}
@media (prefers-reduced-motion: reduce) {
  .ab, .ab__fill { transition: none; }
  .ab:hover { transform: none; }
  .ab--shine.ab--on .ab__icon::after { animation: none; }
}
</style>
