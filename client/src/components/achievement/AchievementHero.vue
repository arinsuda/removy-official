<template>
  <header class="ah">
    <RouterLink :to="{ name: 'user-profile', params: { userId } }" class="ah__back"><ChevronLeft :size="16" aria-hidden="true" />{{ t("navigation.profile") }}</RouterLink>

    <div class="ah__card">
      <div class="ah__who">
        <AppAvatar :src="profile?.avatar_url" :name="displayName" size="xl" decorative />
        <div class="ah__text">
          <h1 class="ah__title">{{ isOwner ? t("achievements.title.own") : t("achievements.title.other", { name: displayName }) }}</h1>
          <p class="ah__sub">{{ isOwner ? t("achievements.subtitle.own") : t("achievements.subtitle.other", { name: displayName }) }}</p>
          <ul class="ah__chips">
            <li v-if="profile" class="ah__chip"><Star :size="14" aria-hidden="true" />{{ t("achievements.hero.level", { level: profile.level }) }}</li>
            <li class="ah__chip"><Zap :size="14" aria-hidden="true" />{{ t("achievements.hero.xp", { n: n(summary.xp) }) }}</li>
          </ul>
        </div>
      </div>
      <AchievementRing :pct="summary.pct" :label="t('achievements.hero.progressLabel')" :sub="t('achievements.hero.counter', { unlocked: n(summary.unlocked), total: n(summary.total) })" />
    </div>

    <div v-if="latest || next" class="ah__spot">
      <button v-if="latest" type="button" class="ah__spot-btn ah__spot-btn--latest" @click="emit('open', latest)">
        <span class="ah__spot-icon"><component :is="getAchievementIcon(latest.achievement.action_type)" :size="22" :stroke-width="1.7" aria-hidden="true" /></span>
        <span class="ah__spot-text">
          <span class="ah__spot-label">{{ t("achievements.hero.latest") }}</span>
          <span class="ah__spot-name">{{ latest.achievement.name }}</span>
          <span v-if="latest.unlocked_at" class="ah__spot-meta">{{ formatDate(latest.unlocked_at) }}</span>
        </span>
      </button>
      <button v-if="next" type="button" class="ah__spot-btn" @click="emit('open', next)">
        <span class="ah__spot-icon ah__spot-icon--muted"><component :is="getAchievementIcon(next.achievement.action_type)" :size="22" :stroke-width="1.7" aria-hidden="true" /></span>
        <span class="ah__spot-text">
          <span class="ah__spot-label">{{ t("achievements.hero.next") }}</span>
          <span class="ah__spot-name">{{ next.achievement.name }}</span>
          <span class="ah__spot-bar" aria-hidden="true"><span :style="{ width: `${progressOf(next)}%` }" /></span>
          <span class="ah__spot-meta">{{ t("achievements.progress", { current: n(Math.min(next.current_count, next.achievement.target_count)), target: n(next.achievement.target_count) }) }}</span>
        </span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink } from "vue-router"
import { ChevronLeft, Star, Zap } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import type { UserAchievement } from "@/types/achievement"
import type { UserProfile } from "@/types/user"
import { getAchievementIcon } from "@/utils/achievementIcons"
import AchievementRing from "./AchievementRing.vue"
import { progressOf } from "./achievements"
import type { OverallSummary } from "./units"

const props = defineProps<{
  userId: number
  profile: UserProfile | null
  isOwner: boolean
  summary: OverallSummary
  latest: UserAchievement | null
  next: UserAchievement | null
}>()
const emit = defineEmits<{ open: [achievement: UserAchievement] }>()

const { t, n, locale } = useI18n()
const displayName = computed(() => props.profile?.display_name || props.profile?.username || "")
const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso))
</script>

<style scoped>
.ah { display: flex; flex-direction: column; gap: 1rem; padding: 1rem 0 0; }
.ah__back { display: inline-flex; align-items: center; align-self: flex-start; gap: 0.25rem; min-height: 2.75rem; font-size: var(--text-small); color: var(--color-text-secondary); text-decoration: none; }
.ah__card { display: flex; flex-direction: column; align-items: center; gap: 1.25rem; padding: 1.5rem 1rem; background: radial-gradient(120% 140% at 0% 0%, color-mix(in srgb, var(--color-brand) 22%, transparent), transparent 60%), var(--color-surface-2); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card); }
.ah__who { display: flex; flex-direction: column; align-items: center; gap: 0.875rem; min-width: 0; text-align: center; }
.ah__text { min-width: 0; }
.ah__title { margin: 0; font-size: var(--text-heading); font-weight: 700; line-height: 1.2; color: var(--color-text-primary); overflow-wrap: anywhere; }
.ah__sub { margin: 0.25rem 0 0; font-size: var(--text-small); color: var(--color-text-secondary); }
.ah__chips { display: flex; flex-wrap: wrap; justify-content: center; gap: 0.5rem; margin: 0.75rem 0 0; padding: 0; list-style: none; }
.ah__chip { display: inline-flex; align-items: center; gap: 0.375rem; min-height: 1.75rem; padding: 0 0.75rem; font-size: var(--text-caption); font-weight: 700; color: var(--color-accent); background: color-mix(in srgb, var(--color-surface) 60%, transparent); border: 1px solid var(--color-border-strong); border-radius: var(--radius-btn); }
.ah__spot { display: grid; gap: 0.75rem; }
.ah__spot-btn { display: flex; align-items: center; gap: 0.875rem; min-width: 0; min-height: 4.5rem; padding: 0.75rem 1rem; font: inherit; text-align: start; color: inherit; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); cursor: pointer; transition: border-color var(--motion-fast) var(--ease-water), transform var(--motion-fast) var(--ease-out-soft); }
.ah__spot-btn--latest { border-color: color-mix(in srgb, var(--color-brand) 45%, transparent); }
.ah__spot-icon { display: grid; flex: none; place-items: center; width: 2.75rem; height: 2.75rem; color: var(--color-accent); background: var(--color-brand); border-radius: 50%; box-shadow: 0 0 18px color-mix(in srgb, var(--color-brand) 45%, transparent); }
.ah__spot-icon--muted { color: var(--color-text-secondary); background: var(--color-surface-3); box-shadow: none; }
.ah__spot-text { display: flex; flex: 1; flex-direction: column; gap: 0.2rem; min-width: 0; }
.ah__spot-label { font-size: var(--text-caption); font-weight: 600; letter-spacing: 0.04em; color: var(--color-text-muted); }
.ah__spot-name { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); overflow-wrap: anywhere; }
.ah__spot-meta { font-size: var(--text-caption); color: var(--color-text-secondary); }
.ah__spot-bar { display: block; height: 0.3125rem; overflow: hidden; background: var(--color-surface-3); border-radius: 9999px; }
.ah__spot-bar > span { display: block; height: 100%; background: var(--color-brand); border-radius: inherit; transition: width var(--motion-base) var(--ease-out-soft); }
@media (hover: hover) {
  .ah__back:hover { color: var(--color-text-primary); }
  .ah__spot-btn:hover { border-color: var(--color-border-strong); transform: translateY(-2px); }
}
@media (min-width: 640px) {
  .ah__spot { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (min-width: 768px) {
  .ah__card { flex-direction: row; justify-content: space-between; padding: 2rem; }
  .ah__who { flex-direction: row; gap: 1.25rem; text-align: start; }
  .ah__chips { justify-content: flex-start; }
  .ah__title { font-size: var(--text-page); }
}
@media (prefers-reduced-motion: reduce) {
  .ah__spot-btn, .ah__spot-bar > span { transition: none; }
  .ah__spot-btn:hover { transform: none; }
}
</style>
