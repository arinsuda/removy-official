<template>
  <section class="activity" :aria-labelledby="titleId">
    <h2 :id="titleId" class="activity__title">{{ t("media.detail.activity.title") }}</h2>

    <p v-if="!loggedIn" class="activity__guest">
      {{ t("media.detail.activity.guest") }}
      <RouterLink :to="{ name: 'login', query: { redirect: route.fullPath } }" class="activity__link">
        {{ t("media.detail.activity.signIn") }}
      </RouterLink>
    </p>

    <dl v-else class="activity__list">
      <div class="activity__item">
        <dt>{{ t("media.detail.activity.rating") }}</dt>
        <dd v-if="myRating" data-test="activity-rating" class="activity__rating">
          <StarRatingInput :model-value="myRating" readonly :size="16" :show-label="false" :aria-label="t('media.detail.activity.rating')" />
          <span>{{ myRating }}/5</span>
        </dd>
        <dd v-else class="activity__none">{{ t("media.detail.activity.notRated") }}</dd>
      </div>
      <div class="activity__item">
        <dt>{{ t("media.detail.activity.watched") }}</dt>
        <dd data-test="activity-watch">
          <template v-if="watchCount > 0">
            {{ t("media.detail.activity.watchCount", { count: watchCount }, watchCount) }}
            <span v-if="lastDate" class="activity__none">· {{ t("media.detail.activity.last", { date: lastDate }) }}</span>
          </template>
          <span v-else class="activity__none">{{ t("media.detail.activity.notWatched") }}</span>
        </dd>
      </div>
      <div class="activity__item">
        <dt>{{ t("media.detail.activity.lists") }}</dt>
        <dd>
          <ul class="activity__chips">
            <li v-if="isWatchlisted" class="activity__chip">{{ t("media.detail.actions.inWatchlist") }}</li>
            <li v-if="isWatched" class="activity__chip">{{ t("media.detail.actions.inWatched") }}</li>
            <li v-if="isLiked" class="activity__chip activity__chip--like">{{ t("media.detail.actions.liked") }}</li>
            <li v-if="!isWatchlisted && !isWatched && !isLiked" class="activity__none">{{ t("media.detail.activity.noLists") }}</li>
          </ul>
        </dd>
      </div>
    </dl>
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink, useRoute } from "vue-router"
import StarRatingInput from "@/components/movie/StarRatingInput.vue"
import type { WatchSummaryResponse } from "@/types/watchLog"

const props = defineProps<{
  loggedIn: boolean
  myRating: number | null
  watchSummary: WatchSummaryResponse | null
  isWatchlisted: boolean
  isWatched: boolean
  isLiked: boolean
}>()

const { t, locale } = useI18n()
const route = useRoute()
const titleId = useId()

const watchCount = computed(() => props.watchSummary?.watch_count ?? 0)
const lastDate = computed(() => {
  const iso = props.watchSummary?.last_watched_on
  if (!iso) return ""
  const d = new Date(`${iso.slice(0, 10)}T00:00:00`)
  return Number.isNaN(d.getTime()) ? "" : new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(d)
})
</script>

<style scoped>
.activity { padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.activity__title { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.activity__guest { margin-top: 0.75rem; font-size: var(--text-small); color: var(--color-text-secondary); }
.activity__link { color: var(--color-brand); font-weight: 600; text-decoration: underline; text-underline-offset: 3px; }
.activity__list { display: flex; flex-direction: column; gap: 0.75rem; margin-top: 0.75rem; }
.activity__item { display: flex; flex-direction: column; gap: 0.2rem; }
.activity__item dt { font-size: var(--text-caption); font-weight: 700; letter-spacing: 0.04em; text-transform: uppercase; color: var(--color-text-muted); }
.activity__item dd { margin: 0; font-size: var(--text-small); color: var(--color-text-primary); }
.activity__rating { display: flex; align-items: center; gap: 0.5rem; font-weight: 700; color: var(--color-rating); }
.activity__none { font-size: var(--text-small); color: var(--color-text-muted); }
.activity__chips { display: flex; flex-wrap: wrap; gap: 0.4rem; margin: 0; padding: 0; list-style: none; }
.activity__chip { padding: 0.15rem 0.6rem; font-size: var(--text-caption); font-weight: 600; background: var(--color-surface-3); border: 1px solid var(--color-border-strong); border-radius: var(--radius-btn); }
.activity__chip--like { color: var(--color-brand); border-color: var(--color-brand); }
</style>
