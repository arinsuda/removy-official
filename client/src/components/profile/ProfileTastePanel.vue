<template>
  <aside class="taste" :aria-label="t('profile.taste.title')">
    <section class="card">
      <h2 class="card__title">{{ t("profile.taste.title") }}</h2>
      <div class="taste__avg">
        <Star :size="18" class="taste__star" aria-hidden="true" />
        <span class="taste__avg-value">{{ taste.average === null ? "-" : taste.average.toFixed(1) }}</span>
        <span class="taste__persona">{{ t(`profile.taste.persona.${taste.persona}`) }}</span>
      </div>
      <p v-if="taste.topGenres.length" class="taste__genres-label">{{ t("profile.taste.topGenres") }}</p>
      <ul v-if="taste.topGenres.length" class="taste__genres">
        <li v-for="g in taste.topGenres" :key="g.id" class="taste__genre">{{ g.name }}</li>
      </ul>
    </section>

    <RatingDistribution :reviews="reviews" :title="t('profile.taste.distribution')" include-all />

    <section v-if="stats" class="card">
      <div class="taste__counts">
        <div><span class="taste__num">{{ stats.watched_count }}</span><span class="taste__lbl">{{ t("profile.taste.watched") }}</span></div>
        <div><span class="taste__num">{{ stats.watchlist_count }}</span><span class="taste__lbl">{{ t("profile.taste.watchlist") }}</span></div>
        <div><span class="taste__num">{{ stats.like_count }}</span><span class="taste__lbl">{{ t("profile.taste.likes") }}</span></div>
      </div>
    </section>

    <section class="card">
      <h2 class="card__title">{{ t("profile.taste.level", { level: stats?.level ?? level }) }}</h2>
      <ul v-if="achievements.length" class="taste__ach">
        <li v-for="a in achievements" :key="a.achievement_id" class="taste__ach-item">
          <Award :size="14" aria-hidden="true" />{{ a.achievement.name }}
        </li>
      </ul>
      <p v-else class="taste__empty">{{ t("profile.taste.noAchievements") }}</p>
    </section>
  </aside>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { useQuery } from "@tanstack/vue-query"
import { Award, Star } from "lucide-vue-next"
import { achievementApi } from "@/api/endpoints/achievement"
import type { UserStats } from "@/api/endpoints/user"
import RatingDistribution from "@/components/detail/RatingDistribution.vue"
import type { ReviewResponse } from "@/types/review"
import type { TasteSummary } from "./tasteProfile"

const props = defineProps<{
  userId: number
  taste: TasteSummary
  reviews: ReviewResponse[]
  stats?: UserStats
  level: number
}>()
const { t } = useI18n()

const ach = useQuery({
  queryKey: computed(() => ["user-achievements-latest", props.userId]),
  queryFn: () => achievementApi.listByUser(props.userId, { unlocked: true, limit: 20 }),
})
const achievements = computed(() =>
  [...(ach.data.value?.data ?? [])]
    .sort((a, b) => (b.unlocked_at ?? "").localeCompare(a.unlocked_at ?? ""))
    .slice(0, 3),
)
</script>

<style scoped>
.taste { display: flex; flex-direction: column; gap: 1rem; }
.card { padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.card__title { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.taste__avg { display: flex; align-items: center; gap: 0.5rem; margin-top: 0.75rem; }
.taste__star { color: var(--color-rating); }
.taste__avg-value { font-size: 1.75rem; font-weight: 700; color: var(--color-text-primary); }
.taste__persona { padding: 0.125rem 0.5rem; font-size: 0.75rem; font-weight: 600; color: var(--color-text-secondary); background: var(--color-surface-3); border-radius: 999px; }
.taste__genres-label { margin-top: 0.75rem; font-size: var(--text-caption); color: var(--color-text-muted); }
.taste__genres { display: flex; flex-wrap: wrap; gap: 0.375rem; margin-top: 0.375rem; padding: 0; list-style: none; }
.taste__genre { padding: 0.125rem 0.625rem; font-size: 0.75rem; color: var(--color-text-secondary); border: 1px solid var(--color-border-strong); border-radius: 999px; }
.taste__counts { display: grid; grid-template-columns: repeat(3, 1fr); text-align: center; }
.taste__counts div { display: flex; flex-direction: column; }
.taste__num { font-size: 1.125rem; font-weight: 700; color: var(--color-text-primary); }
.taste__lbl, .taste__empty { font-size: var(--text-caption); color: var(--color-text-muted); }
.taste__ach { display: flex; flex-direction: column; gap: 0.375rem; margin-top: 0.5rem; padding: 0; list-style: none; }
.taste__ach-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8125rem; color: var(--color-text-secondary); }
.taste__empty { margin-top: 0.5rem; }
</style>
